/**
 * A drag across a card's label selects text and must not also expand the card.
 * Run: bun src/process/cardActivation.test.ts
 */
import { selectionInside } from './cardActivation'

let failed = 0

function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

function node(contains: boolean): Node {
  return { contains: () => contains } as unknown as Node
}

function main() {
  const root = node(true)
  assert(!selectionInside(null, { isCollapsed: false, anchorNode: node(true) }), 'no card, no selection')
  assert(!selectionInside(root, null), 'a click with no selection still activates')
  assert(!selectionInside(root, { isCollapsed: true, anchorNode: node(true) }), 'a plain click is a collapsed caret')
  assert(!selectionInside(root, { isCollapsed: false, anchorNode: null }), 'a range with no anchor is not a selection')
  assert(
    !selectionInside(node(false), { isCollapsed: false, anchorNode: node(true) }),
    'a selection outside the card does not block it',
  )
  assert(
    selectionInside(root, { isCollapsed: false, anchorNode: node(true) }),
    'a drag that highlighted text inside the card is not an activation',
  )

  if (failed) {
    console.error(`${failed} assertion(s) failed`)
    process.exit(1)
  }
  console.log('all assertions passed')
}

main()
