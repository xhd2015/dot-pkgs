/**
 * Composer Enter sends, Shift+Enter does not.
 * Run: bun src/composerKey.test.ts
 */
import { composerSubmitKey } from './composerKey'

let failed = 0
function assertEqual(actual: unknown, expected: unknown, label: string) {
  if (actual !== expected) {
    failed++
    console.error(`✗ ${label}\n  actual: ${actual}\n  expected: ${expected}`)
  } else {
    console.log(`✓ ${label}`)
  }
}

assertEqual(composerSubmitKey({ key: 'Enter', shiftKey: false }), true, 'Enter sends')
assertEqual(composerSubmitKey({ key: 'Enter', shiftKey: true }), false, 'Shift+Enter is a newline')
assertEqual(composerSubmitKey({ key: 'a', shiftKey: false }), false, 'other keys do not send')
assertEqual(
  composerSubmitKey({ key: 'Enter', shiftKey: false, nativeEvent: { isComposing: true } }),
  false,
  'IME composition does not send',
)
assertEqual(
  composerSubmitKey({ key: 'Enter', shiftKey: false, keyCode: 229 }),
  false,
  'keyCode 229 does not send',
)

if (failed > 0) {
  console.error(`\n${failed} assertion(s) failed`)
  process.exit(1)
}
