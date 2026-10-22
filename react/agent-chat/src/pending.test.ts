/**
 * Pending items drop once the server has the text.
 * Run: bun src/pending.test.ts
 */
import { pendingStillWaiting } from './pending'

let failed = 0
function assertEqual(actual: unknown, expected: unknown, label: string) {
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    failed++
    console.error(`✗ ${label}\n  actual: ${JSON.stringify(actual)}\n  expected: ${JSON.stringify(expected)}`)
  } else {
    console.log(`✓ ${label}`)
  }
}

const sending = { id: '1', text: 'hello', status: 'sending' as const, seen: 0 }
assertEqual(
  pendingStillWaiting(['hello'], [sending]).map((item) => item.id),
  [],
  'a sending item drops once the server has a new copy of its text',
)
assertEqual(
  pendingStillWaiting([], [sending]).map((item) => item.id),
  ['1'],
  'a sending item stays while the server has not seen it',
)

const failedItem = { id: '2', text: 'hello', status: 'failed' as const, seen: 0, error: 'down' }
assertEqual(
  pendingStillWaiting(['hello'], [failedItem]).map((item) => item.id),
  ['2'],
  'a failed item stays until retry even after the server has the text',
)

const second = { id: '3', text: 'hello', status: 'sending' as const, seen: 1 }
assertEqual(
  pendingStillWaiting(['hello'], [second]).map((item) => item.id),
  ['3'],
  'a later send of the same text waits until the server has more copies',
)

if (failed > 0) {
  console.error(`\n${failed} assertion(s) failed`)
  process.exit(1)
}
