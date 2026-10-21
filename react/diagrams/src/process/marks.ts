import { HIGHLIGHT_LABEL, changeKind, type ChangeDescription, type ChangeDescriptionKind } from '../change'

export type Mark = {
  /** Which kind of change it is, or null when this is not a mark. */
  kind: ChangeDescriptionKind | null
  /** Class the diagram puts on the card or the connector. */
  className: string
  /** Tooltip: why the change was made. */
  title: string
}

const NO_MARK: Mark = { kind: null, className: '', title: '' }

/**
 * How a marked node or edge reads in the diagram. One definition for both: the
 * card and the connector have to agree, and they used to be two copies that
 * could drift. This is the single kit file that touches the change vocabulary.
 */
export function markOf(change: ChangeDescription | undefined): Mark {
  const kind = changeKind({ change })
  if (!kind) return NO_MARK
  return { kind, className: `pf-mark-${kind}`, title: change?.reason || '' }
}

/** The word a mark carries, for the chip on a card. */
export function markLabel(kind: ChangeDescriptionKind): string {
  return HIGHLIGHT_LABEL[kind]
}
