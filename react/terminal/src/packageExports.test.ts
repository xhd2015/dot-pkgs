import assert from 'node:assert/strict'
import { readFileSync, statSync } from 'node:fs'
import test from 'node:test'

const rootURL = new URL('../../../', import.meta.url)
const root = JSON.parse(readFileSync(new URL('package.json', rootURL), 'utf8'))
const terminal = JSON.parse(readFileSync(new URL('../package.json', import.meta.url), 'utf8'))

test('GitHub root package exposes the terminal source entry', () => {
  assert.equal(root.exports['./terminal'], './react/terminal/src/index.ts')
  assert.ok(statSync(new URL(root.exports['./terminal'], rootURL)).isFile())
})

test('root consumers receive every terminal runtime dependency', () => {
  for (const [name, version] of Object.entries(terminal.dependencies)) {
    assert.equal(root.dependencies[name], version, `${name} must not rely on the consumer installing it accidentally`)
  }
  assert.equal(root.peerDependencies.react, terminal.peerDependencies.react)
  assert.equal(root.peerDependencies['react-dom'], terminal.peerDependencies['react-dom'])
  assert.ok(root.sideEffects.includes('**/*.css'), 'xterm styles must remain side-effectful')
})

test('existing diagrams import paths remain source-compatible', () => {
  for (const name of ['sequence', 'process', 'drawio']) {
    const target = root.exports[`./${name}`]
    assert.ok(typeof target === 'string' && target.startsWith('./react/diagrams/'), name)
    assert.ok(statSync(new URL(target, rootURL)).isFile(), name)
  }
})
