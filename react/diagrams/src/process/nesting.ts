import type { APIFlowchart } from './model'

/**
 * The bookkeeping behind expanding a node in place. A nested breakdown is
 * addressed by the path of node ids from the root, so every question here is
 * about paths, not about charts: which node is open, what a click does to the
 * path, and what the breadcrumb reads.
 */

/** True when `full` is inside `prefix` (or equal to it): a node is inside an open path. */
export function pathStartsWith(full: string[], prefix: string[]): boolean {
  return full.length >= prefix.length && prefix.every((value, i) => full[i] === value)
}

export function pathEquals(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((value, i) => value === b[i])
}

/**
 * The path after clicking a node: open it, or close it when it is already open.
 * Sibling paths replace each other, so one branch is open at a time.
 */
export function togglePath(current: string[], path: string[], parent: string[]): string[] {
  return pathEquals(current, path) ? parent.slice() : path
}

/**
 * The breadcrumb for an open path: the chart's own label, then the label of each
 * node the path opens, falling back to the node id.
 */
export function crumbsFor(chart: APIFlowchart, expandPath: string[]): string[] {
  const labels = [chart.label || 'Process']
  let current = chart
  for (const id of expandPath) {
    const node = (current.nodes || []).find((item) => item.id === id)
    labels.push(node?.flowchart_breakdown?.label || node?.label || id)
    if (node?.flowchart_breakdown) current = node.flowchart_breakdown
  }
  return labels
}

/**
 * The path to the first chart that carries a change mark, so a technical design
 * opens on the part it changes instead of making the reader hunt for it. Returns
 * the container path when this chart is the marked one, or null when nothing
 * below it is marked.
 */
export function revealPath(chart: APIFlowchart, container: string[] = []): string[] | null {
  const markedHere =
    (chart.nodes || []).some((node) => node.change) || (chart.edges || []).some((edge) => edge.change)
  if (markedHere) return container
  for (const node of chart.nodes || []) {
    if (!node.flowchart_breakdown) continue
    const found = revealPath(node.flowchart_breakdown, [...container, node.id])
    if (found) return found
  }
  return null
}
