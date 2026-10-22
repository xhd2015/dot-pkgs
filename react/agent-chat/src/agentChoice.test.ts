/**
 * The chat header's runner, model and effort.
 * Run: bun src/agentChoice.test.ts
 */
import {
  agentChoiceLabel,
  defaultAgentChoice,
  modelOptions,
  reconcileAgentChoice,
  selectAgentRunner,
  type AgentModels,
} from './agentChoice'

let failed = 0

function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

function assertEqual(got: unknown, want: unknown, msg: string) {
  assert(JSON.stringify(got) === JSON.stringify(want), `${msg} (got ${JSON.stringify(got)})`)
}

const LUNA: AgentModels = {
  models: ['default', 'gpt-5.6-luna', 'gpt-5.6-terra'],
  reasoning: ['default', 'low', 'high', 'xhigh', 'max', 'ultra'],
  model_reasoning: {
    'gpt-5.6-luna': ['default', 'low', 'high', 'xhigh', 'max'],
  },
}

assertEqual(agentChoiceLabel(defaultAgentChoice()), 'codex · gpt-5.6-terra (high)', 'a new chat names terra at high')

const grok = selectAgentRunner(defaultAgentChoice(), 'grok')
assertEqual(
  [grok.runner, grok.model, grok.reasoning, grok.by_runner.codex],
  ['grok', 'default', 'default', { model: 'gpt-5.6-terra', reasoning: 'high' }],
  'leaving Codex remembers terra / high',
)
const back = selectAgentRunner(
  { ...grok, model: 'ais-kimi', reasoning: 'default' },
  'codex',
)
assertEqual(
  [back.runner, back.model, back.reasoning],
  ['codex', 'gpt-5.6-terra', 'high'],
  'returning to Codex restores its pair',
)

const snapped = reconcileAgentChoice(
  { runner: 'codex', model: 'gpt-5.6-luna', reasoning: 'ultra', by_runner: {} },
  LUNA,
)
assert(snapped.changed && snapped.choice.reasoning === 'high', 'ultra on luna snaps to high')

assertEqual(
  modelOptions(LUNA, 'gpt-5.6-custom'),
  [...LUNA.models, 'gpt-5.6-custom'],
  'a stored model the catalog lacks stays selectable',
)

if (failed) {
  console.error(`${failed} assertion(s) failed`)
  process.exit(1)
}
