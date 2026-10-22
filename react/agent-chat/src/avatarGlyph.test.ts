import assert from 'node:assert/strict'
import { avatarGlyph } from './avatarGlyph'

assert.equal(avatarGlyph('agent', '老板'), '老')
assert.equal(avatarGlyph('user', '我'), '我')
assert.equal(avatarGlyph('agent'), 'AI')
assert.equal(avatarGlyph('user'), 'U')
console.log('avatarGlyph.test.ts ok')
