import type { NodeSize } from './layout'

export type GraphNodeSizes = Record<string, NodeSize>

/** The browser scheduling boundary, injectable for deterministic resize tests. */
export type GraphSizeEnvironment = {
  createObserver: (onResize: () => void) => Pick<ResizeObserver, 'observe' | 'disconnect'>
  requestFrame: (run: () => void) => number
  cancelFrame: (id: number) => void
}

const browserEnvironment: GraphSizeEnvironment = {
  createObserver: (onResize) => new ResizeObserver(onResize),
  requestFrame: (run) => requestAnimationFrame(run),
  cancelFrame: (id) => cancelAnimationFrame(id),
}

function sameSizes(a: GraphNodeSizes, b: GraphNodeSizes): boolean {
  const ids = Object.keys(a)
  return ids.length === Object.keys(b).length && ids.every((id) => a[id].width === b[id]?.width && a[id].height === b[id]?.height)
}

/**
 * Observe only this graph's direct cards. A nested graph can settle after its
 * parent rendered, so its enclosing card must keep reporting its real box.
 * Position transforms do not affect offsetWidth/Height and cannot feed back
 * into these measurements. Resize bursts share a frame; unchanged boxes do
 * not schedule another layout.
 */
export function observeGraphNodes(
  root: HTMLElement,
  onChange: (sizes: GraphNodeSizes) => void,
  environment: GraphSizeEnvironment = browserEnvironment,
): () => void {
  const cards = [...root.querySelectorAll<HTMLElement>(':scope > [data-node-id] > .pf-node')]
  let previous: GraphNodeSizes | null = null
  let frame: number | null = null
  let stopped = false

  function measure() {
    frame = null
    if (stopped) return
    const next: GraphNodeSizes = {}
    for (const card of cards) {
      const id = card.parentElement?.dataset.nodeId
      if (id != null) next[id] = { width: card.offsetWidth, height: card.offsetHeight }
    }
    if (!previous || !sameSizes(previous, next)) {
      previous = next
      onChange(next)
    }
  }

  const observer = environment.createObserver(() => {
    if (!stopped && frame == null) frame = environment.requestFrame(measure)
  })
  for (const card of cards) observer.observe(card, { box: 'border-box' })
  // Initial placement and expand/collapse should not wait for an observer tick.
  measure()

  return () => {
    stopped = true
    observer.disconnect()
    if (frame != null) environment.cancelFrame(frame)
  }
}
