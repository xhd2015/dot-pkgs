/**
 * Sequence JSON → Mermaid source: copy must not be parsed as diagram syntax.
 * Run: bun src/sequence/sequenceMermaid.test.ts
 */
import mermaid from 'mermaid'
import type { Sequence } from './model'
import { toSequenceMermaid } from './sequenceMermaid'

mermaid.initialize({ startOnLoad: false, securityLevel: 'loose' })

let failed = 0

function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

async function assertParses(src: string, msg: string) {
  try {
    await mermaid.parse(src)
    console.log('ok:', msg)
  } catch (e) {
    failed++
    console.error('FAIL:', msg, e instanceof Error ? e.message : e)
  }
}

function seq(partial: Pick<Sequence, 'participants' | 'messages'> & { autonumber?: boolean }): Sequence {
  return { autonumber: false, ...partial }
}

async function main() {
  const story269 = seq({
    autonumber: true,
    participants: [
      { id: 'user', label: 'User', kind: 'actor' },
      { id: 'page', label: 'AS Checkout Page', kind: 'participant' },
      { id: 'spl', label: 'SPL', kind: 'participant' },
      { id: 'pc', label: 'Pricing Center (PC rules & user quota)', kind: 'participant' },
    ],
    messages: [
      {
        id: 'm2',
        from: 'page',
        to: 'spl',
        text: '提交checkout下单请求: submit order request with selected rule',
        arrow: 'solid',
      },
      {
        id: 'm3',
        from: 'spl',
        to: 'pc',
        text: "validate the PC rule's user quota > 0",
        arrow: 'solid',
      },
      {
        id: 'm5',
        from: 'spl',
        to: 'page',
        text: 'pass → continue; quota = 0 → selected PC rule no longer valid',
        arrow: 'dashed',
      },
      {
        id: 'm6',
        from: 'page',
        to: 'spl',
        text: '确认checkout: confirm checkout, initiate transaction',
        arrow: 'solid',
      },
    ],
  })
  const src = toSequenceMermaid(story269)
  assert(
    src.includes('participant pc as Pricing Center (PC rules & user quota)'),
    'a participant alias with & is left as copy',
  )
  assert(
    src.includes('page->>spl: 提交checkout下单请求: submit order request with selected rule'),
    'a colon in the message stays as copy',
  )
  assert(
    src.includes("spl->>pc: validate the PC rule's user quota > 0"),
    'a > in the message stays as copy',
  )
  assert(
    src.includes('spl-->>page: pass → continue； quota = 0 → selected PC rule no longer valid'),
    'a semicolon becomes fullwidth so Mermaid does not start a new statement',
  )
  assert(!/continue; quota/.test(src), 'ASCII semicolon is not left in the mermaid source')
  await assertParses(src, 'story 269 Typical Path compiles to mermaid that parses')

  const messy = seq({
    participants: [{ id: 'dp', label: 'DP "app"', kind: 'actor' }],
    messages: [{ id: 'm1', from: 'dp', to: 'dp', text: 'line1\nline2 "quoted"', arrow: 'solid' }],
  })
  const messySrc = toSequenceMermaid(messy)
  assert(messySrc.includes("actor dp as DP 'app'"), 'inner double quotes become apostrophes')
  assert(messySrc.includes("Note over dp: line1 line2 'quoted'"), 'newlines become spaces')
  await assertParses(messySrc, 'sanitized quotes and newlines still parse')

  const named = seq({
    participants: [
      { id: 'pc', label: 'credit.pricing.center', kind: 'participant' },
      { id: 'bff', label: 'credit.usercenter.bff', kind: 'participant' },
    ],
    messages: [
      {
        id: 'm1',
        from: 'pc',
        to: 'bff',
        api: 'GetUserStatusForInactive',
        description: 'inactive users still continue',
        arrow: 'solid',
      },
    ],
  })
  const namedSrc = toSequenceMermaid(named)
  assert(
    namedSrc.includes('pc->>bff: GetUserStatusForInactive inactive users still continue'),
    'api and description render on one line',
  )
  assert(!namedSrc.includes('<br'), 'the description is not a second line')
  await assertParses(namedSrc, 'api plus description still parses')

  const selfCall = seq({
    participants: [{ id: 'pc', label: 'credit.pricing.center', kind: 'participant' }],
    messages: [
      {
        id: 'm1',
        from: 'pc',
        to: 'pc',
        api: 'RuleMatchService.BatchMatch',
        description: 'match pricing rules against the in-memory rule cache',
        arrow: 'solid',
      },
    ],
  })
  const selfSrc = toSequenceMermaid(selfCall)
  assert(
    selfSrc.includes('pc->>pc: RuleMatchService.BatchMatch match pricing rules against the in-memory rule cache'),
    'a named self call is an arrow on that lifeline',
  )
  assert(!selfSrc.includes('Note over'), 'a named self call is not a note')
  await assertParses(selfSrc, 'a self arrow parses')

  const branched = seq({
    participants: [
      { id: 'pc', label: 'credit.pricing.center', kind: 'participant' },
      { id: 'product', label: 'credit.core.product', kind: 'participant' },
    ],
    messages: [
      { id: 'a', from: '', to: '', fragment: 'alt', guard: 'activated or frozen', arrow: 'solid' },
      {
        id: 'p',
        from: 'pc',
        to: 'product',
        api: 'ProductQuery',
        description: 'products for the checkout amount',
        arrow: 'solid',
      },
      { id: 'e', from: '', to: '', fragment: 'else', guard: 'not activated', arrow: 'solid' },
      {
        id: 'r',
        from: 'pc',
        to: 'product',
        api: 'RateInfoQuery',
        description: 'term interest rates',
        arrow: 'solid',
      },
      { id: 'z', from: '', to: '', fragment: 'end', arrow: 'solid' },
    ],
  })
  const branchedSrc = toSequenceMermaid(branched)
  assert(branchedSrc.includes('alt activated or frozen'), 'alt guard')
  assert(branchedSrc.includes('else not activated'), 'else guard')
  assert(branchedSrc.includes('\n  end\n'), 'branch end')
  await assertParses(branchedSrc, 'alt branch parses')
}

await main()

if (failed > 0) {
  console.error(`\n${failed} assertion(s) failed`)
  throw new Error(`${failed} assertion(s) failed`)
}
console.log('\nAll sequence mermaid tests passed')
