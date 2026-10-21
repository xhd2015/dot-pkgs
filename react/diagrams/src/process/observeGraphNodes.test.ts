/** Run: bun src/process/observeGraphNodes.test.ts */
import { observeGraphNodes, type GraphNodeSizes, type GraphSizeEnvironment } from './observeGraphNodes'

let failed = 0
function assert(condition: unknown, message: string) {
  if (condition) console.log('ok:', message)
  else {
    failed++
    console.error('FAIL:', message)
  }
}

function card(id: string, width = 300, height = 88) {
  return { parentElement: { dataset: { nodeId: id } }, offsetWidth: width, offsetHeight: height } as unknown as HTMLElement
}

function harness(cards: HTMLElement[]) {
  const changes: GraphNodeSizes[] = []
  const frames = new Map<number, () => void>()
  const observed: Array<{ element: Element; box?: ResizeObserverBoxOptions }> = []
  let notify = () => {}
  let serial = 0
  let disconnected = false
  let selected = ''
  const root = {
    querySelectorAll: (selector: string) => {
      selected = selector
      return cards
    },
  } as unknown as HTMLElement
  const environment: GraphSizeEnvironment = {
    createObserver: (run) => {
      notify = run
      return {
        observe: (element, options) => observed.push({ element, box: options?.box }),
        disconnect: () => { disconnected = true },
      }
    },
    requestFrame: (run) => {
      const id = serial++
      frames.set(id, run)
      return id
    },
    cancelFrame: (id) => { frames.delete(id) },
  }
  const stop = observeGraphNodes(root, (sizes) => changes.push(sizes), environment)
  return {
    changes, frames, observed, stop,
    notify: () => notify(),
    selected: () => selected,
    disconnected: () => disconnected,
    flush: () => {
      const pending = [...frames.values()]
      frames.clear()
      for (const run of pending) run()
    },
  }
}

function resize(element: HTMLElement, width: number, height: number) {
  Object.assign(element, { offsetWidth: width, offsetHeight: height })
}

// --- actual box changes must keep reaching the parent layout ---------------
{
  const batch = card('batch')
  const h = harness([batch])
  assert(h.changes[0]?.batch.height === 88 && h.frames.size === 0, 'initial border box is reported synchronously')
  for (const [width, height] of [[690, 800], [703, 1567], [690, 1827], [740, 1990]]) {
    resize(batch, width, height)
    h.notify()
    h.flush()
  }
  assert(h.changes.length === 5 && h.changes[4].batch.height === 1990, 'late child growth is reported beyond two passes')
  resize(batch, 380, 83)
  h.notify()
  h.flush()
  assert(h.changes.at(-1)?.batch.height === 83, 'collapse shrinks the reported box again')
  resize(batch, 420, 83)
  h.notify()
  h.flush()
  assert(h.changes.at(-1)?.batch.width === 420, 'a width-only change also updates placement')
  h.stop()
}

// --- duplicate notifications share a frame, unchanged boxes do not relayout -
{
  const batch = card('batch')
  const h = harness([batch])
  resize(batch, 690, 1200)
  h.notify()
  h.notify()
  resize(batch, 703, 1567)
  h.notify()
  assert(h.frames.size === 1 && h.changes.length === 1, 'resize bursts queue only one frame')
  h.flush()
  assert(h.changes.length === 2 && h.changes[1].batch.height === 1567, 'the frame reads the latest box, not an earlier event')
  for (let i = 0; i < 4; i++) {
    h.notify()
    h.flush()
  }
  assert(h.changes.length === 2 && h.frames.size === 0, 'unchanged measurements stop without a layout loop')
  h.stop()
}

// --- one graph owns only its direct cards, even with reused descendant ids --
{
  const outer = harness([card('batch', 690, 1827), card('exec', 229, 55)])
  const inner = harness([card('batch', 260, 96)])
  assert(outer.selected() === ':scope > [data-node-id] > .pf-node', 'measurement excludes descendant graph cards')
  assert(outer.changes[0].batch.height === 1827 && inner.changes[0].batch.height === 96, 'reused node ids have independent graph measurements')
  assert(outer.observed.length === 2 && outer.observed.every((entry) => entry.box === 'border-box'), 'only direct card border boxes are observed')
  outer.stop()
  inner.stop()
}

// --- lifecycle changes cannot deliver stale sizes --------------------------
{
  const h = harness([card('removed')])
  h.notify()
  const staleFrame = [...h.frames.values()][0]
  h.stop()
  assert(h.disconnected() && h.frames.size === 0, 'cleanup disconnects and cancels pending work, including frame id zero')
  h.notify()
  staleFrame()
  assert(h.frames.size === 0 && h.changes.length === 1, 'late observer callbacks and cancelled frames cannot publish after cleanup')
  const replacement = harness([card('replacement', 400, 150)])
  assert(Object.keys(replacement.changes[0]).join(',') === 'replacement', 'a replacement graph reports no removed-node cache entries')
  replacement.stop()
}
{
  const h = harness([])
  assert(Object.keys(h.changes[0]).length === 0 && h.observed.length === 0, 'an empty graph reports an empty measurement set')
  h.stop()
}

if (failed) throw new Error(`${failed} assertion(s) failed`)
console.log('observeGraphNodes: all assertions passed')
