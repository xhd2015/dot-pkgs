/**
 * Who owns which class, and the guard that keeps it that way. Run:
 * bun src/process/cssOwnership.test.ts
 */
import { readFileSync, readdirSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))

let failed = 0
function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

const OWNER: Record<string, string> = {}
function owns(file: string, classes: string[]) {
  for (const name of classes) {
    if (OWNER[name]) throw new Error(`${name} is already owned by ${OWNER[name]}`)
    OWNER[name] = file
  }
}

owns('ProcessDiagram.css', [
  'pf-shell',
  'pf-embed',
  'pf-empty',
  'pf-crumbs',
  'pf-crumb',
  'pf-crumb-item',
  'pf-sep',
  'pf-canvas',
  'pf-hint',
])
owns('ProcessGraph.css', [
  'pf-graph',
  'pf-graph-node',
  'pf-edges',
  'pf-edge',
  'pf-edge-layer',
  'pf-edge-label',
  'pf-nest',
])
owns('ProcessNodeCard.css', [
  'pf-node',
  'pf-has',
  'pf-response',
  'pf-decision',
  'pf-expanded',
  'pf-head',
  'pf-icon',
  'pf-ttl',
  'pf-sub',
  'pf-pills',
  'pf-pill',
  'pf-pill-mark',
  'pf-fold',
  'pf-status',
])
owns('ProcessDiamond.css', ['pd-branch', 'pd-branch-outline', 'pd-branch-copy', 'pd-branch-title', 'pd-branch-sub'])

const MARKS = ['pf-mark-add', 'pf-mark-change', 'pf-mark-remove', 'pf-mark-reference']
const MODIFIERS = new Set(['current'])

function selectors(css: string): string[] {
  const out: string[] = []
  let i = 0
  while (i < css.length) {
    if (css.startsWith('/*', i)) {
      const end = css.indexOf('*/', i + 2)
      if (end < 0) break
      i = end + 2
      continue
    }
    if (/\s/.test(css[i])) {
      i += 1
      continue
    }
    if (css[i] === '@') {
      const open = css.indexOf('{', i)
      if (open < 0) break
      let depth = 1
      let k = open + 1
      while (k < css.length && depth) {
        if (css[k] === '{') depth += 1
        else if (css[k] === '}') depth -= 1
        k += 1
      }
      out.push(...selectors(css.slice(open + 1, k - 1)))
      i = k
      continue
    }
    const open = css.indexOf('{', i)
    if (open < 0) break
    out.push(css.slice(i, open))
    let depth = 1
    let k = open + 1
    while (k < css.length && depth) {
      if (css[k] === '{') depth += 1
      else if (css[k] === '}') depth -= 1
      k += 1
    }
    i = k
  }
  return out
}

const sheets = readdirSync(here)
  .filter((name) => name.endsWith('.css'))
  .map((name) => ({ owner: name, path: join(here, name), css: readFileSync(join(here, name), 'utf8') }))

{
  const declared = new Set(Object.values(OWNER))
  for (const sheet of sheets) {
    assert(declared.has(sheet.owner), `${sheet.owner} declares the classes it owns`)
  }
}

for (const sheet of sheets) {
  const foreign: string[] = []
  for (const selector of selectors(sheet.css)) {
    for (const match of selector.matchAll(/\.(-?[_a-zA-Z][\w-]*)/g)) {
      const name = match[1]
      if (MODIFIERS.has(name)) continue
      if (MARKS.includes(name)) continue
      if (!name.startsWith('pf-') && !name.startsWith('pd-')) continue
      if (OWNER[name] !== sheet.owner) foreign.push(`${name} (owned by ${OWNER[name] ?? 'nothing'})`)
    }
  }
  assert(
    foreign.length === 0,
    `${sheet.owner} selects only its own classes${foreign.length ? ` — found ${[...new Set(foreign)].join(', ')}` : ''}`,
  )
}

{
  const wanted: Array<[string, string]> = [
    ['ProcessDiagram.tsx', 'ProcessDiagram.css'],
    ['ProcessGraph.tsx', 'ProcessGraph.css'],
    ['ProcessNodeCard.tsx', 'ProcessNodeCard.css'],
    ['ProcessDiamond.tsx', 'ProcessDiamond.css'],
  ]
  for (const [tsx, css] of wanted) {
    const source = readFileSync(join(here, tsx), 'utf8')
    assert(source.includes(`import './${css}'`), `${tsx} imports ${css}`)
    const others = [...source.matchAll(/import '\.\/([\w.]+\.css)'/g)].map((m) => m[1]).filter((name) => name !== css)
    assert(others.length === 0, `${tsx} imports no other component's stylesheet${others.length ? `: ${others.join(', ')}` : ''}`)
  }
}

if (failed) {
  console.error(`${failed} assertion(s) failed`)
  throw new Error(`${failed} assertion(s) failed`)
}
console.log('cssOwnership: all assertions passed')
