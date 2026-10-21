/**
 * The diagram's expansion bookkeeping. Run:
 * bun src/process/nesting.test.ts
 */
import type { APIFlowchart, APIFlowchartNode } from './model'
import { crumbsFor, pathEquals, pathStartsWith, revealPath, togglePath } from './nesting'

let failed = 0
function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

function node(id: string, label: string, breakdown?: APIFlowchart): APIFlowchartNode {
  return { id, label, ...(breakdown ? { flowchart_breakdown: breakdown } : {}) }
}

function chart(id: string, nodes: APIFlowchartNode[]): APIFlowchart {
  return { id, label: id, nodes, edges: [] }
}

// --- paths ----------------------------------------------------------------
{
  assert(pathStartsWith(['a', 'b'], ['a']), 'a nested path is inside its parent')
  assert(pathStartsWith(['a'], ['a']), 'a path is inside itself')
  assert(!pathStartsWith(['a'], ['a', 'b']), 'a parent is not inside its child')
  assert(!pathStartsWith(['a', 'c'], ['a', 'b']), 'a sibling is not inside the open path')
  assert(pathEquals(['a', 'b'], ['a', 'b']), 'equal paths compare equal')
  assert(!pathEquals(['a'], ['a', 'b']), 'a shorter path is not equal')
  assert(!pathEquals(['a', 'b'], ['b', 'a']), 'order matters')
}

// --- clicking a node ------------------------------------------------------
{
  const root = ['a']
  const child = ['a', 'b']
  assert(togglePath([], root, []).join('/') === 'a', 'clicking a closed node opens it')
  assert(togglePath(root, root, []).length === 0, 'clicking an open node closes it to its parent')
  assert(
    togglePath(child, child, root).join('/') === 'a',
    'clicking a nested open node closes it to its own parent, not to the root',
  )
  assert(
    togglePath(['a', 'b'], ['a', 'c'], ['a']).join('/') === 'a/c',
    'opening a sibling replaces the branch that was open',
  )
}

// --- breadcrumbs ----------------------------------------------------------
{
  const inner = chart('inner-chart', [node('x', 'Inner step')])
  const root = chart('root-chart', [node('a', 'Outer', inner), node('b', 'Other')])
  assert(crumbsFor(root, []).join(' / ') === 'root-chart', 'a closed diagram reads as its own label')
  assert(
    crumbsFor(root, ['b']).join(' / ') === 'root-chart / Other',
    'a node with no breakdown reads as its own label',
  )
  assert(
    crumbsFor(root, ['a']).join(' / ') === 'root-chart / inner-chart',
    'a node with a breakdown reads as the breakdown chart label, which is what it opens',
  )
  assert(
    crumbsFor(root, ['a', 'x']).join(' / ') === 'root-chart / inner-chart / Inner step',
    'a nested crumb follows the opened chart into its own node labels',
  )
  const noLabel = chart('', [node('a', 'Outer')])
  noLabel.label = undefined
  assert(crumbsFor(noLabel, []).join(' / ') === 'Process', 'an unlabelled chart reads as Process')
  assert(
    crumbsFor(root, ['missing']).join(' / ') === 'root-chart / missing',
    'an unknown id still produces a crumb rather than dropping it',
  )
}

// --- revealing the first marked chart ------------------------------------
{
  const unmarked = chart('plain', [node('a', 'A'), node('b', 'B')])
  assert(revealPath(unmarked) === null, 'nothing marked anywhere returns null')

  const markedTop = chart('top', [node('a', 'A')])
  markedTop.nodes[0].change = { kind: 'add', reason: 'new' }
  assert(revealPath(markedTop)?.length === 0, 'a mark on the top chart opens nothing (already visible)')

  const deep = chart('deep', [node('x', 'X')])
  deep.nodes[0].change = { kind: 'change', reason: 'edited' }
  const middle = chart('middle', [node('m', 'M', deep)])
  const root = chart('root', [node('a', 'A'), node('b', 'B', middle)])
  assert(
    revealPath(root)?.join('/') === 'b/m',
    `a mark two levels down opens the path to it (${revealPath(root)?.join('/')})`,
  )

  const markedEdge = chart('edge-marked', [node('a', 'A'), node('b', 'B')])
  markedEdge.edges = [{ from: 'a', to: 'b', change: { kind: 'remove', reason: 'gone' } }]
  assert(revealPath(markedEdge)?.length === 0, 'a marked edge counts as a mark on its chart')

  const siblingFirst = chart('sib', [node('p', 'P', deep)])
  const twoBranches = chart('two', [node('first', 'First'), node('second', 'Second', siblingFirst)])
  assert(
    revealPath(twoBranches)?.join('/') === 'second/p',
    'the first branch that contains a mark wins, and an unmarked branch is skipped',
  )
}

if (failed) {
  console.error(`${failed} assertion(s) failed`)
  throw new Error(`${failed} assertion(s) failed`)
}
console.log('nesting: all assertions passed')
