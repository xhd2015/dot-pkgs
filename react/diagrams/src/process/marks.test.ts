/**
 * How a change mark reads on a node or an edge. Run:
 * bun src/process/marks.test.ts
 */
import type { ChangeDescription } from './model'
import { markLabel, markOf } from './marks'

let failed = 0
function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

{
  const add = markOf({ kind: 'add', reason: 'new step' })
  assert(add.className === 'pf-mark-add', `an addition claims the add class (${add.className})`)
  assert(add.kind === 'add', 'the kind is reported, so a chip can name it')
  assert(add.title === 'new step', 'the reason becomes the tooltip')
}
{
  assert(markOf({ kind: 'remove', reason: 'gone' }).className === 'pf-mark-remove', 'a removal claims its class')
  assert(markOf({ kind: 'reference', reason: 'context' }).className === 'pf-mark-reference', 'a reference keeps its kind')
  assert(markOf(undefined).className === '', 'no change is not a mark')
  assert(markOf(undefined).title === '', 'no change has no tooltip')
  assert(markOf(undefined).kind === null, 'no change reports no kind')
  assert(markOf({ kind: 'add', reason: '' }).kind === null, 'a mark with no reason reports no kind')
}
{
  // The store treats a mark with no kind as a change, and so does the diagram.
  const bare = markOf({ reason: 'edited' })
  assert(bare.className === 'pf-mark-change', `a mark without a kind reads as a change (${bare.className})`)
  assert(bare.kind === 'change', 'and it reports the change kind')
  // The store can send a kind this type does not name; the diagram still has to
  // read it as a change rather than drop the mark.
  const loose = (kind: string) => markOf({ kind, reason: 'x' } as unknown as ChangeDescription)
  assert(loose('').className === 'pf-mark-change', 'an empty kind reads as a change')
  assert(loose('ADD').className === 'pf-mark-add', 'the kind is case-insensitive')
  assert(loose('nonsense').className === 'pf-mark-change', 'an unknown kind reads as a change')
  assert(markOf({ kind: 'add', reason: '' }).className === '', 'a mark with no reason is not a mark')
}
{
  assert(markLabel('add') === 'Add' && markLabel('change') === 'Change', 'labels come from the shared vocabulary')
  assert(markLabel('remove') === 'Remove' && markLabel('reference') === 'Existing', 'every kind has a word')
}

if (failed) {
  console.error(`${failed} assertion(s) failed`)
  throw new Error(`${failed} assertion(s) failed`)
}
console.log('marks: all assertions passed')
