import test from 'node:test'
import assert from 'node:assert/strict'
import { controlChar, keyboardInset, terminalKeys } from './terminalKeys.ts'

test('control keys are the bytes the shell expects', () => {
    assert.equal(terminalKeys.c, '\x03')
    assert.equal(terminalKeys.a, '\x01')
    assert.equal(terminalKeys.r, '\x12')
    assert.equal(terminalKeys.l, '\x0c')
    assert.equal(terminalKeys.tab, '\t')
    assert.equal(terminalKeys.esc, '\x1b')
    assert.equal(terminalKeys.up, '\x1b[A')
    assert.equal(controlChar('d'), '\x04')
    assert.equal(controlChar('D'), '\x04')
    assert.equal(controlChar('1'), null)
})

test('keyboard inset is the layout height the visual viewport no longer covers', () => {
    assert.equal(keyboardInset(800, 800, 0), 0)
    assert.equal(keyboardInset(800, 430, 0), 370)
    assert.equal(keyboardInset(800, 430, 20), 350)
    assert.equal(keyboardInset(800, 900, 0), 0)
})
