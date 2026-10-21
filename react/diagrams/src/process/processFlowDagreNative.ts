import { Graph, layout } from '@dagrejs/dagre'
import type { APIFlowchart, APIFlowchartNode } from './model'
import type { DagreLayout, LaidEdge, LaidNode, Point } from './layout'

const DECISION = { width: 240, height: 240 }
const STEP = { width: 300, height: 88 }

/**
 * A fresh box per node: dagre writes the placement onto the label object it was
 * handed, so returning one shared constant would make every step alias the same
 * position.
 */
function defaultSize(node: APIFlowchartNode): { width: number; height: number } {
  return node.kind === 'branch' ? { ...DECISION } : { ...STEP }
}

/** The point where the segment to `outside` leaves the card `node`. */
function clipToBox(node: LaidNode, inside: Point, outside: Point): Point {
  const dx = outside.x - inside.x
  const dy = outside.y - inside.y
  const hx = node.width / 2
  const hy = node.height / 2
  if (!dx && !dy) return { x: node.x, y: node.y + hy }
  const scale = Math.min(dx ? hx / Math.abs(dx) : Infinity, dy ? hy / Math.abs(dy) : Infinity)
  return { x: node.x + dx * scale, y: node.y + dy * scale }
}

/**
 * Dagre's own dummy-node polyline: the placement *and* the connectors come from
 * dagre, so a rank order dagre chose is drawn the way dagre expects. Endpoints
 * are clipped to the card so an arrow stays visible where it meets the box.
 *
 * The names differ from `layoutDagre` on purpose: that one keeps dagre's ranks
 * but replaces the connectors with the channel router, which is what the
 * "dagre" panel of the debugging page shows.
 */
export function layoutDagreNative(
  chart: APIFlowchart,
  sizeOf?: (id: string) => { width: number; height: number } | undefined,
): DagreLayout {
  const nodes = chart.nodes || []
  const edges = chart.edges || []
  const g = new Graph({ multigraph: true })
  g.setGraph({ rankdir: 'TB', nodesep: 56, ranksep: 72, edgesep: 20, marginx: 32, marginy: 32 })
  g.setDefaultEdgeLabel(() => ({}))
  for (const node of nodes) {
    const measured = sizeOf?.(node.id)
    // Dagre writes x/y onto this object, so it must never be the caller's cache.
    g.setNode(node.id, { ...(measured || defaultSize(node)) })
  }
  edges.forEach((edge, i) => {
    const label = edge.label || ''
    g.setEdge(edge.from, edge.to, { label, width: label ? Math.max(24, label.length * 7) : 0, height: label ? 16 : 0 }, `e${i}`)
  })
  layout(g)
  const graph = g.graph()
  const byId = Object.fromEntries(nodes.map((node) => [node.id, node]))
  const laidNodes: LaidNode[] = nodes.map((node) => {
    const placed = g.node(node.id)
    return {
      id: node.id,
      node: byId[node.id] || node,
      x: placed?.x || 0,
      y: placed?.y || 0,
      width: placed?.width || 0,
      height: placed?.height || 0,
      rank: 0,
    }
  })
  const nodeById = Object.fromEntries(laidNodes.map((node) => [node.id, node]))
  const laidEdges: LaidEdge[] = edges.map((edge, i) => {
    const raw = g.edge(edge.from, edge.to, `e${i}`)
    const points = ((raw?.points || []) as Point[]).map((point) => ({ x: point.x, y: point.y }))
    const src = nodeById[edge.from]
    const dst = nodeById[edge.to]
    if (src && points.length >= 2) points[0] = clipToBox(src, points[0], points[1])
    if (dst && points.length >= 2) points[points.length - 1] = clipToBox(dst, points[points.length - 1], points[points.length - 2])
    return {
      key: `${edge.from}->${edge.to}:${edge.label || ''}:${i}`,
      edge,
      points,
      labelX: raw?.x,
      labelY: raw?.y,
    }
  })
  // An edge whose target is not a node makes dagre invent one with no size, and
  // every coordinate downstream of it is NaN. The canvas stays finite instead.
  const xs = laidEdges.flatMap((edge) => edge.points.map((point) => point.x)).filter(Number.isFinite)
  const ys = laidEdges.flatMap((edge) => edge.points.map((point) => point.y)).filter(Number.isFinite)
  // An empty chart reports -Infinity, which is truthy, so it cannot be defaulted with ||.
  const graphWidth = typeof graph.width === 'number' && Number.isFinite(graph.width) ? graph.width : 1
  const graphHeight = typeof graph.height === 'number' && Number.isFinite(graph.height) ? graph.height : 1
  return {
    width: Math.max(graphWidth, ...xs, 1) + 32,
    height: Math.max(graphHeight, ...ys, 1) + 32,
    nodes: laidNodes,
    edges: laidEdges,
  }
}
