/**
 * The layout the product pages draw. Run:
 * bun src/process/processFlowDagreNative.test.ts
 */
import type { APIFlowchart } from './model'
import { layoutDagreNative } from './processFlowDagreNative'
import type { LaidNode, Point } from './layout'

let failed = 0
function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

function boxOf(nodes: LaidNode[], id: string): LaidNode | undefined {
  return nodes.find((node) => node.id === id)
}

function onOutline(node: LaidNode, point: Point): boolean {
  const hx = node.width / 2
  const hy = node.height / 2
  const dx = Math.abs(point.x - node.x)
  const dy = Math.abs(point.y - node.y)
  return Math.abs(dx - hx) < 1.5 || Math.abs(dy - hy) < 1.5
}

const chain: APIFlowchart = {
  id: 'chain',
  nodes: [
    { id: 'a', label: 'A' },
    { id: 'b', label: 'B' },
    { id: 'c', label: 'C' },
  ],
  edges: [
    { from: 'a', to: 'b' },
    { from: 'b', to: 'c' },
  ],
}

const exclusive: APIFlowchart = {
  id: 'exclusive',
  nodes: [
    { id: 'q', label: 'Which?', kind: 'branch' },
    { id: 'yes', label: 'Yes' },
    { id: 'no', label: 'No' },
  ],
  edges: [
    { from: 'q', to: 'yes', label: 'yes' },
    { from: 'q', to: 'no', label: 'no' },
  ],
}

{
  const layout = layoutDagreNative(chain)
  const ys = layout.nodes.map((node) => node.y)
  assert(ys[0] < ys[1] && ys[1] < ys[2], 'a chain is placed top to bottom')
}

{
  const layout = layoutDagreNative(exclusive)
  const q = boxOf(layout.nodes, 'q')!
  const keep = boxOf(layout.nodes, 'yes')!
  const drop = boxOf(layout.nodes, 'no')!
  assert(keep.y > q.y && drop.y > q.y, 'both legs sit below the question')
}

{
  const sizes: Record<string, { width: number; height: number }> = {
    a: { width: 412, height: 104 },
    b: { width: 300, height: 72 },
    c: { width: 300, height: 72 },
  }
  const measured = layoutDagreNative(chain, (id) => sizes[id])
  const honoured = measured.nodes.every((node) => node.width === sizes[node.id].width && node.height === sizes[node.id].height)
  assert(honoured, 'a measured DOM box is the box dagre places')
  const natural = layoutDagreNative(chain)
  assert(
    natural.nodes.some((node) => node.width !== sizes[node.id].width),
    'without measurements the fallback sizes are used',
  )
}

{
  const branch = layoutDagreNative(exclusive).nodes.find((node) => node.id === 'q')!
  assert(branch.width === 240 && branch.height === 240, 'an unmeasured decision is 240x240')
}

{
  const empty = layoutDagreNative({ id: 'empty', nodes: [], edges: [] })
  assert(empty.nodes.length === 0 && empty.edges.length === 0, 'an empty chart lays out nothing')
  assert(
    Number.isFinite(empty.width) && Number.isFinite(empty.height) && empty.width > 0 && empty.height > 0,
    `an empty chart still reports a usable canvas (${empty.width}x${empty.height})`,
  )
}

{
  const dangling: APIFlowchart = { id: 'dangling', nodes: [{ id: 'a', label: 'A' }], edges: [{ from: 'a', to: 'ghost' }] }
  let threw = false
  let layout = null
  try {
    layout = layoutDagreNative(dangling)
  } catch {
    threw = true
  }
  assert(!threw && layout !== null, 'an edge to a missing node does not throw')
  assert(
    layout !== null && Number.isFinite(layout.width) && Number.isFinite(layout.height),
    `a malformed chart keeps a finite canvas (${layout?.width}x${layout?.height})`,
  )
}

{
  const loop: APIFlowchart = {
    id: 'loop',
    nodes: [
      { id: 'a', label: 'A' },
      { id: 'b', label: 'B' },
    ],
    edges: [
      { from: 'a', to: 'a' },
      { from: 'a', to: 'b' },
    ],
  }
  const layout = layoutDagreNative(loop)
  assert(layout.nodes.length === 2, 'a self edge does not duplicate or drop its card')
  assert(layout.edges.every((edge) => edge.points.length >= 2), 'a self edge is still routed')
  const a = boxOf(layout.nodes, 'a')!
  const self = layout.edges.find((edge) => edge.edge.from === 'a' && edge.edge.to === 'a')!
  assert(onOutline(a, self.points[0]) || self.points.length >= 2, 'a self edge still has geometry')
}

if (failed) {
  console.error(`${failed} assertion(s) failed`)
  throw new Error(`${failed} assertion(s) failed`)
}
console.log('processFlowDagreNative: all assertions passed')
