import { useId, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { line as d3Line, select } from 'd3'
import type { LaidEdge, Point } from './layout'
import { layoutDagreNative } from './processFlowDagreNative'
import { ProcessNodeCard } from './ProcessNodeCard'
import type { APIFlowchart, APIFlowchartNode } from './model'
import { markOf } from './marks'
import { pathStartsWith, togglePath } from './nesting'
import { observeGraphNodes, type GraphNodeSizes } from './observeGraphNodes'
import './ProcessGraph.css'

const edgeLine = d3Line<Point>()
  .x((p) => p.x)
  .y((p) => p.y)

type Expand = {
  /** The node ids open above this graph, from the root. */
  pathPrefix: string[]
  expandPath: string[]
  onExpandPath: (next: string[]) => void
}

/**
 * One node of a graph: the card, plus the nested breakdown it opens. The path
 * from the root decides whether this node is the open one, so a sibling opening
 * closes this one.
 */
function NodeSlot({ node, allowExpand, pathPrefix, expandPath, onExpandPath }: { node: APIFlowchartNode; allowExpand: boolean } & Expand) {
  const path = [...pathPrefix, node.id]
  const has = Boolean(node.flowchart_breakdown)
  const expanded = allowExpand && has && pathStartsWith(expandPath, path)

  return (
    <ProcessNodeCard
      node={node}
      hasDetail={has}
      expanded={expanded}
      onActivate={allowExpand ? () => onExpandPath(togglePath(expandPath, path, pathPrefix)) : undefined}
    >
      {expanded && node.flowchart_breakdown ? (
        <div className="pf-nest">
          <ProcessGraph
            chart={node.flowchart_breakdown}
            pathPrefix={path}
            expandPath={expandPath}
            onExpandPath={onExpandPath}
          />
        </div>
      ) : null}
    </ProcessNodeCard>
  )
}

/**
 * A process chart drawn to scale: each graph observes its own cards, including
 * the enclosing box of an expanded breakdown. Late child resizing therefore
 * updates the parent placement and connectors too.
 */
export function ProcessGraph({ chart, pathPrefix, expandPath, onExpandPath }: { chart: APIFlowchart } & Expand) {
  const rootRef = useRef<HTMLDivElement>(null)
  const svgRef = useRef<SVGSVGElement>(null)
  const [sizes, setSizes] = useState<GraphNodeSizes>({})
  const laid = useMemo(() => layoutDagreNative(chart, (id) => sizes[id]), [chart, sizes])

  const markerRaw = useId()
  const markerId = `pf-arrow-${markerRaw.replace(/:/g, '')}`

  useLayoutEffect(() => {
    const root = rootRef.current
    if (!root) return
    return observeGraphNodes(root, setSizes)
  }, [chart, expandPath])

  useLayoutEffect(() => {
    const svg = svgRef.current
    if (!svg) return
    const edges = laid.edges.filter((item) => item.points.length >= 2)
    select(svg)
      .select('g.pf-edge-layer')
      .selectAll<SVGPathElement, LaidEdge>('path.pf-edge')
      .data(edges, (item) => item.key)
      .join('path')
      .attr('class', (item) => ['pf-edge', markOf(item.edge.change).className].filter(Boolean).join(' '))
      .attr('d', (item) => edgeLine(item.points))
      .attr('data-edge-from', (item) => item.edge.from)
      .attr('data-edge-to', (item) => item.edge.to)
      .attr('data-edge-label', (item) => item.edge.label || '')
      .attr('marker-end', `url(#${markerId})`)
  }, [laid, markerId])

  return (
    <div className="pf-graph" ref={rootRef} style={{ width: laid.width, height: laid.height }}>
      <svg className="pf-edges" ref={svgRef} width={laid.width} height={laid.height} aria-hidden>
        <defs>
          <marker id={markerId} markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto">
            <path d="M0 0 L8 4 L0 8 z" fill="#bfbfbf" />
          </marker>
        </defs>
        <g className="pf-edge-layer" />
      </svg>
      {laid.edges.map((item) =>
        item.edge.label && item.labelX != null && item.labelY != null ? (
          <div
            key={`${item.key}-label`}
            className={['pf-edge-label', markOf(item.edge.change).className].filter(Boolean).join(' ')}
            style={{ left: item.labelX, top: item.labelY }}
          >
            {item.edge.label}
          </div>
        ) : null,
      )}
      {laid.nodes.map((item) => (
        <div
          key={item.id}
          className="pf-graph-node"
          data-node-id={item.id}
          style={{ left: item.x, top: item.y }}
        >
          <NodeSlot
            node={item.node}
            allowExpand
            pathPrefix={pathPrefix}
            expandPath={expandPath}
            onExpandPath={onExpandPath}
          />
        </div>
      ))}
    </div>
  )
}
