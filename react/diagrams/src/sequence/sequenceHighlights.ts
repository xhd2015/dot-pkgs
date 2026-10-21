import type { Sequence, SequenceMessage } from './model'
import { HIGHLIGHT_LABEL, HIGHLIGHT_TAG_COLOR, changeKind, type ChangeDescriptionKind } from '../change'

export type { ChangeDescriptionKind }
export { HIGHLIGHT_LABEL, HIGHLIGHT_TAG_COLOR, changeKind }

/** Class the diagram pass puts on a marked step or lifeline. A reference mark colours
 * nothing — the diagram is there to show what changed — so its class is null and the pass
 * skips it rather than adding an empty class name. */
export const HIGHLIGHT_CLASS: Record<ChangeDescriptionKind, string | null> = {
  add: 'aiw-step-add',
  change: 'aiw-step-change',
  remove: 'aiw-step-remove',
  reference: null,
}

export const NODE_HIGHLIGHT_CLASS: Record<ChangeDescriptionKind, string | null> = {
  add: 'aiw-seq-node-add',
  change: 'aiw-seq-node-change',
  remove: 'aiw-seq-node-remove',
  reference: null,
}

/** One rendered step of the diagram, in source order. */
export type StepMark = {
  /** The source message this step came from. */
  messageID: string
  kind: ChangeDescriptionKind | null
}

/**
 * How the diagram should be coloured. Mermaid draws notes and arrows in two
 * separate passes — every note first, then every arrow — so each group is planned
 * on its own and zipped with its own rendered group, never against document order.
 *
 * A self-message with no API compiles to a note. One that names an API compiles
 * to an arrow, the same as a call to another participant.
 */
export type StepMarkPlan = {
  arrows: StepMark[]
  notes: StepMark[]
  nodes: Array<{ label: string; kind: ChangeDescriptionKind; reason: string }>
}

export function stepMarkPlan(seq: Sequence | null | undefined): StepMarkPlan {
  const plan: StepMarkPlan = { arrows: [], notes: [], nodes: [] }
  if (!seq) return plan
  for (const p of seq.participants || []) {
    const kind = changeKind(p)
    const label = (p.label || '').trim()
    if (kind && label) plan.nodes.push({ label, kind, reason: p.change?.reason || '' })
  }
  for (const m of seq.messages || []) {
    if (m.fragment) continue
    const mark: StepMark = { messageID: m.id || '', kind: changeKind(m) }
    if (isSelfNote(m)) plan.notes.push(mark)
    else plan.arrows.push(mark)
  }
  return plan
}

/** True when anything in the sequence is marked. */
export function hasHighlights(seq: Sequence | null | undefined): boolean {
  const plan = stepMarkPlan(seq)
  return plan.nodes.length > 0 || plan.arrows.some((s) => s.kind) || plan.notes.some((s) => s.kind)
}

/** One described change, as the list under the diagram reads it. */
export type HighlightEntry = {
  key: string
  kind: ChangeDescriptionKind
  /** Participant label for a node; "FROM → TO" for an edge. */
  title: string
  /** The number the diagram prints on the step, 0 when it prints none. */
  ordinal: number
  reason: string
}

/**
 * Everything this design marks, as descriptions: the nodes in declaration order,
 * then the edges in diagram order, so the list reads top-down like the diagram.
 *
 * The step number is the one the diagram prints. Mermaid numbers the arrows only.
 * A self-message with no API is a note and carries no number. A named self-call
 * is an arrow, so it is numbered with the others.
 */
export function highlightEntries(seq: Sequence | null | undefined): HighlightEntry[] {
  if (!seq) return []
  const labels = new Map<string, string>()
  for (const p of seq.participants || []) {
    labels.set(p.id, (p.label || '').trim() || p.id)
  }
  const entries: HighlightEntry[] = []
  for (const p of seq.participants || []) {
    const kind = changeKind(p)
    if (!kind) continue
    entries.push({
      key: `node:${p.id}`,
      kind,
      title: (p.label || '').trim() || p.id,
      ordinal: 0,
      reason: p.change?.reason || '',
    })
  }
  let stepNumber = 0
  for (const m of seq.messages || []) {
    const isNote = isSelfNote(m)
    if (!isNote) stepNumber++
    const kind = changeKind(m)
    if (!kind) continue
    const from = labels.get(m.from) || m.from
    const to = labels.get(m.to) || m.to
    entries.push({
      key: `edge:${m.id}`,
      kind,
      title: `${from} → ${to}`,
      ordinal: isNote ? 0 : stepNumber,
      reason: m.change?.reason || '',
    })
  }
  return entries
}

/** A self-message is a note only when it names no API. */
function isSelfNote(message: SequenceMessage): boolean {
  return message.from === message.to && !(message.api || '').trim()
}

/** True when this step is a removal, which the diagram draws with a crossed arrow. */
export function isRemoved(message: SequenceMessage): boolean {
  return changeKind(message) === 'remove'
}
