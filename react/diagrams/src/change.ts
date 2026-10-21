/** The kinds of change a change description can carry. */
export type ChangeDescriptionKind = 'add' | 'change' | 'remove' | 'reference'

/** What a tech design changes at a node, an edge, or an API field. `kind` defaults to change. */
export type ChangeDescription = { kind?: ChangeDescriptionKind; reason: string }

/** What a mark looks like in a table or a chip. One hue per kind. */
export const HIGHLIGHT_TAG_COLOR: Record<ChangeDescriptionKind, string> = {
  add: 'green',
  change: 'blue',
  remove: 'red',
  reference: 'default',
}

/** Badge text, capitalised for a table cell. */
export const HIGHLIGHT_LABEL: Record<ChangeDescriptionKind, string> = {
  add: 'Add',
  change: 'Change',
  remove: 'Remove',
  reference: 'Existing',
}

/**
 * The kind of change a node, an edge, or an API field carries, or null when it is not
 * marked. A mark with no kind reads as a change, which is how the store treats it.
 */
export function changeKind(
  entry: { change?: { kind?: string; reason: string } } | undefined,
): ChangeDescriptionKind | null {
  const reason = entry?.change?.reason
  if (!reason) return null
  const kind = (entry?.change?.kind || '').trim().toLowerCase()
  if (kind === 'add' || kind === 'change' || kind === 'remove' || kind === 'reference') return kind
  return 'change'
}
