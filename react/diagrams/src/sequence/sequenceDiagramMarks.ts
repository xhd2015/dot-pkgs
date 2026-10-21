import type { Sequence } from './model'
import {
  HIGHLIGHT_CLASS,
  NODE_HIGHLIGHT_CLASS,
  stepMarkPlan,
  type ChangeDescriptionKind,
  type StepMark,
} from './sequenceHighlights'

const SVG_NS = 'http://www.w3.org/2000/svg'

/** Class Mermaid puts on a step's arrow, whichever shape it drew. */
const ARROW_CLASSES = ['messageLine0', 'messageLine1']

/** Marker Mermaid stamps the step number onto, drawn in near-black. */
const BADGE_MARKER_ID = 'sequencenumber'

/** How far to look along the siblings for a step's arrow and its step number. */
const LOOKAHEAD = 5
const LOOKBEHIND = 3

/**
 * Colour a rendered diagram in place, so a marked step is green or red on the arrow
 * itself rather than inside a background band, and a marked lifeline is filled.
 *
 * Mermaid draws every note before every arrow, so each group is matched against its
 * own plan in source order. When the counts disagree the mapping can no longer be
 * trusted, and nothing is coloured rather than colouring the wrong step.
 *
 * Returns how many marks were applied.
 */
const DESC_CLASS = 'aiw-seq-desc'

/**
 * Colour the description of an arrow that has both an API name and a description.
 * Mermaid draws the whole label as one line; the description becomes the next
 * tspan on that same line so it can take the secondary text colour.
 * Counts must match or nothing is split, same rule as the mark pass.
 */
export function colorSequenceDescriptions(root: ParentNode, seq: Sequence | null | undefined): number {
  const messages = (seq?.messages || []).filter((m) => !m.fragment)
  const arrows = messages.filter((m) => (m.from || '').trim() !== (m.to || '').trim())
  const notes = messages.filter((m) => (m.from || '').trim() !== '' && m.from === m.to)
  return (
    paintDescriptions(root.querySelectorAll('text.messageText'), arrows) +
    paintDescriptions(root.querySelectorAll('text.noteText'), notes)
  )
}

function paintDescriptions(labels: ArrayLike<Element>, messages: Sequence['messages']): number {
  const rows = messages || []
  if (labels.length !== rows.length) return 0
  let applied = 0
  for (let i = 0; i < rows.length; i++) {
    const api = (rows[i].api || '').trim()
    const desc = (rows[i].description || '').trim()
    if (!api || !desc) continue
    if (splitDescription(labels[i], api, desc)) applied++
  }
  return applied
}

function splitDescription(textEl: Element, api: string, desc: string): boolean {
  if (textEl.querySelector(`tspan.${DESC_CLASS}`)) return true
  const spans = Array.from(textEl.children).filter((el) => el.tagName.toLowerCase() === 'tspan')
  const holder = spans[0] ?? null
  const raw = ((holder ? holder.textContent : textEl.textContent) || '').replace(/\s+/g, ' ').trim()
  const prefix = api.replace(/\s+/g, ' ').trim()
  const detail = desc.replace(/\s+/g, ' ').trim()
  if (!raw.startsWith(prefix)) return false
  const rest = raw.slice(prefix.length).trim()
  if (!rest.startsWith(detail)) return false
  if (holder) {
    holder.textContent = prefix
    const tail = document.createElementNS(SVG_NS, 'tspan')
    tail.setAttribute('class', DESC_CLASS)
    tail.textContent = ` ${detail}`
    holder.insertAdjacentElement('afterend', tail)
    return true
  }
  textEl.textContent = ''
  const head = document.createElementNS(SVG_NS, 'tspan')
  head.textContent = prefix
  const tail = document.createElementNS(SVG_NS, 'tspan')
  tail.setAttribute('class', DESC_CLASS)
  tail.textContent = ` ${detail}`
  textEl.append(head, tail)
  return true
}

export function colorSequenceMarks(root: ParentNode, seq: Sequence | null | undefined): number {
  const plan = stepMarkPlan(seq)
  return (
    markSteps(root.querySelectorAll('text.messageText'), plan.arrows) +
    markNotes(root.querySelectorAll('text.noteText'), plan.notes) +
    markNodes(root, plan.nodes)
  )
}

function classOf(el: Element): string {
  return el.getAttribute('class') || ''
}

function markSteps(labels: ArrayLike<Element>, marks: StepMark[]): number {
  if (labels.length !== marks.length) return 0
  let applied = 0
  for (let i = 0; i < marks.length; i++) {
    const kind = marks[i].kind
    const cls = kind ? HIGHLIGHT_CLASS[kind] : null
    // An unmarked step and a reference step both leave the drawing alone: one carries no
    // mark, the other carries one that says nothing changed.
    if (!kind || !cls) continue
    const label = labels[i]
    label.classList.add(cls)
    const arrow = sibling(label, 1, (el) => ARROW_CLASSES.some((c) => classOf(el).includes(c)))
    if (arrow) arrow.classList.add(cls)
    const number = sibling(label, 1, (el) => classOf(el).includes('sequenceNumber'))
    if (number) {
      number.classList.add(cls)
      colorStepBadge(number, kind, cls)
    }
    applied++
  }
  return applied
}

/**
 * Mermaid draws every step number on one shared badge marker, and its number is
 * already white, so a marked step needs a badge of its own in the mark's colour —
 * recolouring the number alone would print it dark on the near-black badge it sits
 * on. The clone is kept in the drawing's defs and reused on later passes.
 */
function colorStepBadge(number: Element, kind: ChangeDescriptionKind, cls: string): void {
  const drawing = number.parentElement
  const line = sibling(number, -1, (el) => el.getAttribute('marker-start') !== null)
  if (!drawing || !line) return
  const id = `aiw-seq-badge-${kind}`
  if (!drawing.querySelector(`marker#${id}`)) {
    const source = drawing.querySelector(`marker#${BADGE_MARKER_ID}`)
    if (!source) return
    const badge = source.cloneNode(true) as Element
    badge.setAttribute('id', id)
    badge.setAttribute('class', `aiw-seq-badge ${cls}`)
    source.parentElement?.appendChild(badge)
  }
  line.setAttribute('marker-start', `url(#${id})`)
}

function markNotes(labels: ArrayLike<Element>, marks: StepMark[]): number {
  if (labels.length !== marks.length) return 0
  let applied = 0
  for (let i = 0; i < marks.length; i++) {
    const kind = marks[i].kind
    const cls = kind ? HIGHLIGHT_CLASS[kind] : null
    if (!kind || !cls) continue
    const label = labels[i]
    label.classList.add(cls)
    // A note's box is the nearest rect before its text; colouring the text alone
    // still reads as marked if Mermaid ever moves the box.
    const box = sibling(label, -1, (el) => el.tagName.toLowerCase() === 'rect')
    if (box) box.classList.add(cls)
    applied++
  }
  return applied
}

function markNodes(
  root: ParentNode,
  nodes: Array<{ label: string; kind: keyof typeof NODE_HIGHLIGHT_CLASS; reason: string }>,
): number {
  if (nodes.length === 0) return 0
  const byLabel = new Map(nodes.map((n) => [n.label, n]))
  let applied = 0
  for (const group of Array.from(root.querySelectorAll('svg > g'))) {
    if (!group.querySelector('rect.actor')) continue
    // The label we compiled with is the participant's identity in the drawing.
    const label = (group.querySelector('text.actor')?.textContent || '').trim()
    const node = byLabel.get(label)
    if (!node) continue
    const cls = NODE_HIGHLIGHT_CLASS[node.kind]
    // A participant included for reference is described, not coloured.
    if (!cls) continue
    group.classList.add(cls)
    applied++
    if (node.reason && !group.querySelector('title')) {
      const title = document.createElementNS(SVG_NS, 'title')
      title.textContent = node.reason
      group.prepend(title)
    }
  }
  return applied
}

/** The nearest sibling in one direction that satisfies match, within a bounded walk. */
function sibling(from: Element, direction: 1 | -1, match: (el: Element) => boolean): Element | null {
  const limit = direction === 1 ? LOOKAHEAD : LOOKBEHIND
  let el = direction === 1 ? from.nextElementSibling : from.previousElementSibling
  for (let i = 0; el && i < limit; i++) {
    if (match(el)) return el
    el = direction === 1 ? el.nextElementSibling : el.previousElementSibling
  }
  return null
}
