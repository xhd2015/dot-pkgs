import test from 'node:test'
import assert from 'node:assert/strict'
import { terminalClose, terminalCreate, terminalInput, terminalResize } from './protocol.ts'

test('create uses the fitted size and falls back when the element is not measurable yet', () => {
    assert.deepEqual(terminalCreate('term-1', 92.8, 28.2), { type: 'create', tabId: 'term-1', cols: 92, rows: 28 })
    assert.deepEqual(terminalCreate('term-1', 0, null), { type: 'create', tabId: 'term-1', cols: 80, rows: 24 })
})

test('resize, input, and close keep the terminal identity', () => {
    assert.deepEqual(terminalResize('term-1', 110, 32), { type: 'resize', tabId: 'term-1', cols: 110, rows: 32 })
    assert.deepEqual(terminalInput('term-1', 'ls\r'), { type: 'input', tabId: 'term-1', data: 'ls\r' })
    assert.deepEqual(terminalClose('term-1'), { type: 'close', tabId: 'term-1' })
})
