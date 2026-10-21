import type { KeyboardEvent, MouseEvent } from 'react'

/** What a card needs to be operable by mouse and keyboard. */
export type CardActivation = {
  /** Click, Enter or Space on the card body. Nothing happens without it. */
  onActivate?: () => void
}

/** The bits of a selection a click needs, so a drag that highlights text is not an activation. */
export type TextSelection = {
  isCollapsed: boolean
  anchorNode: Node | null
}

/**
 * True when the pointer came up on a selection inside the card. A plain click has a
 * collapsed selection, and a selection that lives somewhere else must not block the card.
 */
export function selectionInside(root: Node | null, selection: TextSelection | null): boolean {
  if (!root || !selection || selection.isCollapsed || !selection.anchorNode) return false
  return root.contains(selection.anchorNode)
}

/**
 * The interaction props for a card body. A card is only a button when the
 * surface has something to expand, so a static preview stays a plain div.
 */
export function activationProps({ onActivate }: CardActivation) {
  if (!onActivate) return {}
  return {
    role: 'button' as const,
    tabIndex: 0,
    onClick: (event: MouseEvent<Element>) => {
      // A nested card sits inside this one; its click must not reach this body.
      // A drag that selected the label is not a request to expand the card.
      event.stopPropagation()
      const root = event.currentTarget instanceof Node ? event.currentTarget : null
      if (selectionInside(root, window.getSelection())) return
      onActivate()
    },
    onKeyDown: (event: KeyboardEvent) => {
      if (event.key !== 'Enter' && event.key !== ' ') return
      event.preventDefault()
      event.stopPropagation()
      onActivate()
    },
  }
}
