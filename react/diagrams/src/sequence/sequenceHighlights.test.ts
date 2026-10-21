/**
 * Marking the nodes and edges a tech design changes, and how the diagram is
 * coloured from that plan.
 * Run: bun src/sequence/sequenceHighlights.test.ts
 *
 * There is no DOM in this test runner, so the diagram pass is driven against a
 * small fake shaped like Mermaid's output, and `document.createElementNS` — which
 * the pass uses only for a tooltip — is stubbed.
 */
import type { Sequence, SequenceMessage } from './model'
import {
  HIGHLIGHT_CLASS,
  HIGHLIGHT_LABEL,
  HIGHLIGHT_TAG_COLOR,
  hasHighlights,
  highlightEntries,
  changeKind,
  isRemoved,
  stepMarkPlan,
} from './sequenceHighlights'
import { colorSequenceMarks } from './sequenceDiagramMarks'
import { toSequenceMermaid } from './sequenceMermaid'

let failed = 0

function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

function assertEqual(got: unknown, want: unknown, msg: string) {
  assert(JSON.stringify(got) === JSON.stringify(want), `${msg} (got ${JSON.stringify(got)})`)
}

/** DP → PC real call, PC → DP answer, PC → PC self step, DP → PC last call. */
function sequence(marks: Record<string, string> = {}): Sequence {
  const hit = (id: string) => (marks[id] ? { kind: marks[id] as never, reason: `why ${id}` } : undefined)
  return {
    autonumber: true,
    participants: [
      { id: 'dp', label: 'DP', kind: 'actor', change: hit('dp') },
      { id: 'pc', label: 'credit.pricing.center', kind: 'participant', change: hit('pc') },
    ],
    messages: [
      { id: 'm1', from: 'dp', to: 'pc', text: 'QueryProductCardInfo', arrow: 'solid', change: hit('m1') },
      { id: 'm2', from: 'pc', to: 'dp', text: 'tenor options', arrow: 'dashed', change: hit('m2') },
      { id: 'm3', from: 'pc', to: 'pc', text: 'Group E price-band', arrow: 'solid', change: hit('m3') },
      { id: 'm4', from: 'dp', to: 'pc', text: 'confirm', arrow: 'solid', change: hit('m4') },
    ],
  }
}

// ---- a fake drawing, shaped like Mermaid's output ----

class FakeEl {
  tagName: string
  cls: string
  attrs: Record<string, string>
  text: string | null = null
  kids: FakeEl[] = []
  parentEl: FakeEl | null = null

  constructor(tag: string, cls = '', attrs: Record<string, string> = {}) {
    this.tagName = tag.toUpperCase()
    this.cls = cls
    this.attrs = { ...attrs }
  }

  append(...nodes: FakeEl[]): void {
    for (const n of nodes) {
      n.parentEl = this
      this.kids.push(n)
    }
  }

  prepend(...nodes: FakeEl[]): void {
    for (const n of nodes) {
      n.parentEl = this
      this.kids.unshift(n)
    }
  }

  appendChild(node: FakeEl): FakeEl {
    this.append(node)
    return node
  }

  get textContent(): string {
    return this.text ?? ''
  }

  set textContent(value: string) {
    this.text = value
  }

  // Mirrors the calls the pass makes: classList.add, getAttribute, setAttribute
  // and a deep clone of Mermaid's badge marker.
  get classList() {
    return {
      add: (token: string) => {
        if (!this.cls.split(/\s+/).includes(token)) this.cls = this.cls ? `${this.cls} ${token}` : token
      },
    }
  }

  getAttribute(name: string): string | null {
    if (name === 'class') return this.cls || null
    return this.attrs[name] ?? null
  }

  setAttribute(name: string, value: string): void {
    if (name === 'class') this.cls = value
    else this.attrs[name] = value
  }

  cloneNode(deep: boolean): FakeEl {
    const copy = new FakeEl(this.tagName, this.cls, this.attrs)
    copy.text = this.text
    if (deep) for (const kid of this.kids) copy.append(kid.cloneNode(true))
    return copy
  }

  get nextElementSibling(): FakeEl | null {
    const sibs = this.parentEl?.kids || []
    return sibs[sibs.indexOf(this) + 1] ?? null
  }

  get parentElement(): FakeEl | null {
    return this.parentEl
  }

  get previousElementSibling(): FakeEl | null {
    const sibs = this.parentEl?.kids || []
    const at = sibs.indexOf(this)
    return at > 0 ? sibs[at - 1] : null
  }

  querySelectorAll(sel: string): FakeEl[] {
    if (sel.includes('>')) {
      const [parentTok, childTok] = sel.split('>').map((s) => s.trim())
      return descendants(this).filter(
        (el) => matches(el, childTok) && el.parentEl !== null && matches(el.parentEl, parentTok),
      )
    }
    return descendants(this).filter((el) => matches(el, sel))
  }

  querySelector(sel: string): FakeEl | null {
    return this.querySelectorAll(sel)[0] ?? null
  }
}

function descendants(el: FakeEl, out: FakeEl[] = []): FakeEl[] {
  for (const kid of el.kids) {
    out.push(kid)
    descendants(kid, out)
  }
  return out
}

function matches(el: FakeEl, token: string): boolean {
  const [, tag = '', id = '', classes = ''] = token.match(/^([a-zA-Z]*)(#[^.#]*)?((?:\.[^.#]*)*)$/) || []
  if (tag && el.tagName !== tag.toUpperCase()) return false
  if (id && el.getAttribute('id') !== id.slice(1)) return false
  return classes.split('.').filter(Boolean).every((c) => el.cls.split(/\s+/).includes(c))
}

/**
 * Mermaid draws every note first and then every arrow, each group in source order —
 * that separation is exactly what the pass has to survive. Each step number hangs
 * on a zero-length badge line pointed at the marker Mermaid shares between steps.
 */
function fakeDrawing(arrowCount: number, noteCount: number, actorLabels: string[]): FakeEl {
  const root = new FakeEl('div')
  const svg = new FakeEl('svg')
  const defs = new FakeEl('defs')
  defs.append(new FakeEl('marker', '', { id: 'sequencenumber' }))
  svg.append(defs)
  for (let i = 0; i < noteCount; i++) {
    svg.append(new FakeEl('rect', 'note'), new FakeEl('text', 'noteText'))
  }
  for (let i = 0; i < arrowCount; i++) {
    svg.append(
      new FakeEl('text', 'messageText'),
      new FakeEl('line', 'messageLine0'),
      new FakeEl('line', '', { 'marker-start': 'url(#sequencenumber)' }),
      new FakeEl('text', 'sequenceNumber'),
    )
  }
  for (const label of actorLabels) {
    const group = new FakeEl('g')
    const box = new FakeEl('rect', 'actor actor-bottom')
    const text = new FakeEl('text', 'actor actor-box')
    text.textContent = label
    group.append(box, text)
    svg.append(group)
  }
  root.append(svg)
  return root
}

function colour(root: FakeEl, seq: Sequence | null): number {
  return colorSequenceMarks(root as unknown as ParentNode, seq)
}

function classOf(root: FakeEl, sel: string, index: number): string {
  return root.querySelectorAll(sel)[index].cls
}

function attrOf(root: FakeEl, sel: string, index: number, attr: string): string | null {
  return root.querySelectorAll(sel)[index].getAttribute(attr)
}

function badgeMarkers(root: FakeEl): Array<[string | null, string | null]> {
  return root.querySelectorAll('marker').map((m) => [m.getAttribute('id'), m.getAttribute('class')])
}

function main() {
  // The pass creates a tooltip element through the document; stub just that.
  ;(globalThis as { document?: unknown }).document = {
    createElementNS: (_ns: string, tag: string) => new FakeEl(tag),
  }

  assertEqual(
    HIGHLIGHT_LABEL,
    { add: 'Add', change: 'Change', remove: 'Remove', reference: 'Existing' },
    'the badge reads Add / Change / Remove, and Existing for a row kept for context',
  )
  assertEqual(
    HIGHLIGHT_TAG_COLOR,
    { add: 'green', change: 'blue', remove: 'red', reference: 'default' },
    'one hue per kind: green adds, blue changes, red removes, grey keeps',
  )
  assertEqual(
    [HIGHLIGHT_CLASS.add, HIGHLIGHT_CLASS.change, HIGHLIGHT_CLASS.remove, HIGHLIGHT_CLASS.reference],
    ['aiw-step-add', 'aiw-step-change', 'aiw-step-remove', null],
    'one class per kind, and none for a reference mark',
  )

  // ---- kind ----
  assertEqual(changeKind({ change: { reason: 'x' } }), 'change', 'a mark with no kind is a change')
  assertEqual(changeKind({ change: { kind: 'add', reason: 'x' } }), 'add', 'kind add')
  assertEqual(changeKind({ change: { kind: 'remove', reason: 'x' } }), 'remove', 'kind remove')
  assertEqual(changeKind({ change: { kind: 'CHANGE', reason: 'x' } }), 'change', 'kind is case-insensitive')
  assertEqual(changeKind({}), null, 'an unmarked entry has no kind')
  assertEqual(changeKind({ change: { reason: '' } }), null, 'a blank reason is not a mark')
  assertEqual(changeKind({ change: { kind: 'add', reason: '' } }), null, 'a kind without a reason is not a mark')
  assertEqual(changeKind({ change: { kind: 'reference', reason: 'shown for context' } }), 'reference', 'kind reference')

  assert(!hasHighlights(sequence()), 'an unmarked sequence has nothing to mark')
  assert(hasHighlights(sequence({ m1: 'add' })), 'a marked step counts')
  assert(hasHighlights(sequence({ pc: 'remove' })), 'a marked node counts')
  assert(hasHighlights(sequence({ m1: 'reference' })), 'a reference step is still something to describe')

  // ---- plan ----
  assertEqual(stepMarkPlan(sequence()).arrows.map((s) => s.kind), [null, null, null], 'three arrows, none marked')
  assertEqual(stepMarkPlan(sequence()).notes.map((s) => s.kind), [null], 'the self step is planned as a note')
  assertEqual(stepMarkPlan(sequence()).nodes, [], 'no marked nodes')

  const plan = stepMarkPlan(sequence({ pc: 'add', m2: 'change', m3: 'remove' }))
  assertEqual(
    plan.arrows,
    [
      { messageID: 'm1', kind: null },
      { messageID: 'm2', kind: 'change' },
      { messageID: 'm4', kind: null },
    ],
    'arrows keep source order and carry their kind',
  )
  assertEqual(plan.notes, [{ messageID: 'm3', kind: 'remove' }], 'a self step with no API is planned as a note')
  const namedSelf = sequence()
  namedSelf.messages[2] = { ...namedSelf.messages[2], api: 'RuleMatchService.BatchMatch' }
  assertEqual(
    stepMarkPlan(namedSelf).arrows.map((s) => s.messageID),
    ['m1', 'm2', 'm3', 'm4'],
    'a named self call is planned as an arrow',
  )
  assertEqual(
    plan.nodes,
    [{ label: 'credit.pricing.center', kind: 'add', reason: 'why pc' }],
    'marked nodes carry a reason',
  )
  assertEqual(stepMarkPlan(null).arrows, [], 'a null sequence plans nothing')

  // ---- descriptions ----
  assertEqual(highlightEntries(sequence()), [], 'nothing marked reads as an empty description list')
  assertEqual(
    highlightEntries(sequence({ pc: 'change', m2: 'remove' })),
    [
      { key: 'node:pc', kind: 'change', title: 'credit.pricing.center', ordinal: 0, reason: 'why pc' },
      { key: 'edge:m2', kind: 'remove', title: 'credit.pricing.center → DP', ordinal: 2, reason: 'why m2' },
    ],
    'nodes come first, then edges in diagram order',
  )
  const unknown = sequence({ m1: 'change' })
  unknown.messages.push({ id: 'm5', from: 'ghost', to: 'pc', text: 'x', arrow: 'solid', change: { reason: 'why m5' } })
  assertEqual(
    highlightEntries(unknown).map((e) => e.title),
    ['DP → credit.pricing.center', 'ghost → credit.pricing.center'],
    'edge titles use participant labels, falling back to ids',
  )
  assertEqual(
    highlightEntries(sequence({ m2: 'change', m3: 'remove', m4: 'change' })).map((e) => [e.key, e.ordinal]),
    [
      ['edge:m2', 2],
      ['edge:m3', 0],
      ['edge:m4', 3],
    ],
    'the step number counts arrows only, because a note carries none',
  )

  // ---- removal shape ----
  const removed: SequenceMessage = {
    id: 'm9',
    from: 'dp',
    to: 'pc',
    text: 'x',
    arrow: 'solid',
    change: { kind: 'remove', reason: 'r' },
  }
  assert(isRemoved(removed), 'a removed step is a removal')
  assert(!isRemoved(sequence().messages[0]), 'an unmarked step is not a removal')
  assert(!isRemoved(sequence({ m1: 'change' }).messages[0]), 'a changed step is not a removal')

  // ---- compiled source ----
  const plain = toSequenceMermaid(sequence())
  assert(
    plain ===
      [
        'sequenceDiagram',
        '  autonumber',
        '  actor dp as DP',
        '  participant pc as credit.pricing.center',
        '  dp->>pc: QueryProductCardInfo',
        '  pc-->>dp: tenor options',
        '  Note over pc: Group E price-band',
        '  dp->>pc: confirm',
        '',
      ].join('\n'),
    'an unmarked sequence compiles unchanged',
  )
  assert(!plain.includes('rect'), 'colour is never baked into the source as a band')
  assert(
    toSequenceMermaid(sequence({ m1: 'add', m2: 'change', pc: 'add' })) === plain,
    'colour-only marks leave the source alone',
  )
  assert(
    toSequenceMermaid(sequence({ m1: 'remove' })).includes('  dp-xpc: QueryProductCardInfo'),
    'a removed step is a crossed arrow',
  )
  assert(
    toSequenceMermaid(sequence({ m2: 'remove' })).includes('  pc--xdp: tenor options'),
    'a removed dashed step stays dashed',
  )
  assert(
    toSequenceMermaid(sequence({ m3: 'remove' })).includes('  Note over pc: Group E price-band'),
    'a removed self step stays a note — a note has no arrow to cross',
  )

  // ---- the diagram pass ----
  const marked = sequence({ m2: 'add', m3: 'remove', pc: 'change' })
  const drawing = fakeDrawing(3, 1, ['DP', 'credit.pricing.center'])
  assertEqual(colour(drawing, marked), 3, 'every mark is applied')
  // Notes are drawn before every arrow, so a document-order walk would colour the
  // wrong step here; the note mark must land on the note.
  assertEqual(classOf(drawing, 'text.noteText', 0), 'noteText aiw-step-remove', 'the removed self step is the note')
  assertEqual(classOf(drawing, 'rect.note', 0), 'note aiw-step-remove', 'and the note box is coloured')
  assertEqual(classOf(drawing, 'text.messageText', 0), 'messageText', 'the first arrow is left alone')
  assertEqual(classOf(drawing, 'text.messageText', 1), 'messageText aiw-step-add', 'the added step is the second arrow')
  assertEqual(classOf(drawing, 'line.messageLine0', 1), 'messageLine0 aiw-step-add', 'its arrow is coloured')
  assertEqual(classOf(drawing, 'text.sequenceNumber', 1), 'sequenceNumber aiw-step-add', 'and its step number')
  assertEqual(
    badgeMarkers(drawing),
    [
      ['sequencenumber', null],
      ['aiw-seq-badge-add', 'aiw-seq-badge aiw-step-add'],
    ],
    'a marked step gets a badge of its own, the shared one stays black',
  )
  // Lines run arrow, badge, arrow, badge … so step 1's badge line is the fourth.
  assertEqual(attrOf(drawing, 'line', 3, 'marker-start'), 'url(#aiw-seq-badge-add)', 'and its number sits on it')
  assertEqual(attrOf(drawing, 'line', 1, 'marker-start'), 'url(#sequencenumber)', 'an unmarked step keeps the shared badge')
  assertEqual(classOf(drawing, 'line.messageLine0', 0), 'messageLine0', 'no colour leaks to a neighbour')
  assertEqual(classOf(drawing, 'g', 1), 'aiw-seq-node-change', 'the marked lifeline takes its kind class')
  assertEqual(classOf(drawing, 'g', 0), '', 'the unmarked lifeline is untouched')
  assertEqual(drawing.querySelectorAll('title').length, 1, 'the reason rides along as a tooltip')
  assertEqual(
    drawing.querySelectorAll('title')[0].parentEl?.cls,
    'aiw-seq-node-change',
    'and hangs off the marked lifeline',
  )

  colour(drawing, marked)
  assertEqual(drawing.querySelectorAll('title').length, 1, 'a second pass does not add a second tooltip')
  assertEqual(drawing.querySelectorAll('marker').length, 2, 'nor a second badge')

  const removedArrow = fakeDrawing(3, 1, [])
  colour(removedArrow, sequence({ m1: 'remove' }))
  assertEqual(
    badgeMarkers(removedArrow)[1],
    ['aiw-seq-badge-remove', 'aiw-seq-badge aiw-step-remove'],
    'a removed step gets a badge in the removal colour',
  )

  // One group whose counts disagree is skipped on its own: the other groups are
  // still matched by count, so only the unmappable one goes uncoloured.
  const mismatched = fakeDrawing(2, 1, ['DP'])
  assertEqual(colour(mismatched, marked), 1, 'only the groups that still line up are coloured')
  assertEqual(classOf(mismatched, 'text.messageText', 0), 'messageText', 'the unmappable arrows are left alone')
  assertEqual(classOf(mismatched, 'text.noteText', 0), 'noteText aiw-step-remove', 'the note still marks correctly')

  const unmarked = fakeDrawing(3, 1, ['DP', 'credit.pricing.center'])
  assertEqual(colour(unmarked, sequence()), 0, 'an unmarked sequence marks nothing')
  assertEqual(colour(unmarked, null), 0, 'and neither does a null one')
  assertEqual(classOf(unmarked, 'text.messageText', 0), 'messageText', 'leaving the drawing untouched')

  // A reference mark says nothing changed, so the drawing must read exactly as an
  // unmarked one — while the description list still carries the entry.
  const reference = fakeDrawing(3, 1, ['DP', 'credit.pricing.center'])
  assertEqual(colour(reference, sequence({ m2: 'reference', pc: 'reference' })), 0, 'a reference mark colours nothing')
  assertEqual(classOf(reference, 'text.messageText', 1), 'messageText', 'its step keeps its plain text')
  assertEqual(classOf(reference, 'g', 1), '', 'and its lifeline stays plain')
  assertEqual(badgeMarkers(reference), [['sequencenumber', null]], 'no badge of its own is cloned')
  assertEqual(
    highlightEntries(sequence({ m2: 'reference' })),
    [{ key: 'edge:m2', kind: 'reference', title: 'credit.pricing.center → DP', ordinal: 2, reason: 'why m2' }],
    'but the description list still names it',
  )
}

main()

if (failed > 0) {
  console.error(`\n${failed} assertion(s) failed`)
  throw new Error(`${failed} assertion(s) failed`)
}
console.log('\nAll sequence highlight tests passed')
