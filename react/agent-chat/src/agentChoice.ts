/** The chat's runner, model and reasoning effort. */
export type AgentChoice = {
  runner: 'codex' | 'grok'
  model: string
  reasoning: string
  by_runner: Partial<Record<'codex' | 'grok', { model: string; reasoning: string }>>
}

export type AgentModels = {
  models: string[]
  reasoning: string[]
  model_reasoning?: Record<string, string[]>
}

const CODEX_MODEL = 'gpt-5.6-terra'
const CODEX_EFFORT = 'high'

/** What a new install shows before the preference file exists. */
export function defaultAgentChoice(): AgentChoice {
  return { runner: 'codex', model: CODEX_MODEL, reasoning: CODEX_EFFORT, by_runner: {} }
}

/** The closed header label: runner, model, effort. */
export function agentChoiceLabel(choice: Pick<AgentChoice, 'runner' | 'model' | 'reasoning'>): string {
  return `${choice.runner} · ${choice.model} (${choice.reasoning})`
}

/** Switch runner, remembering the pair being left and restoring the pair being entered. */
export function selectAgentRunner(choice: AgentChoice, runner: AgentChoice['runner']): AgentChoice {
  const byRunner = {
    ...choice.by_runner,
    [choice.runner]: { model: choice.model, reasoning: choice.reasoning },
  }
  const saved = byRunner[runner]
  const fallback = runner === 'grok'
    ? { model: 'default', reasoning: 'default' }
    : { model: CODEX_MODEL, reasoning: CODEX_EFFORT }
  return {
    runner,
    model: saved?.model || fallback.model,
    reasoning: saved?.reasoning || fallback.reasoning,
    by_runner: byRunner,
  }
}

/** Efforts the panel offers for this model. An unknown model keeps the full list. */
export function effortsFor(catalog: AgentModels, model: string): string[] {
  const specific = catalog.model_reasoning?.[model]
  if (specific && specific.length) return specific
  return catalog.reasoning
}

/**
 * Keep a stored effort that this model allows. An effort the model does not
 * list snaps to high for Codex and default for Grok.
 */
export function reconcileAgentChoice(choice: AgentChoice, catalog: AgentModels): { choice: AgentChoice; changed: boolean } {
  const efforts = effortsFor(catalog, choice.model)
  if (!efforts.length || efforts.includes(choice.reasoning)) {
    return { choice, changed: false }
  }
  const reasoning = choice.runner === 'grok' ? 'default' : efforts.includes(CODEX_EFFORT) ? CODEX_EFFORT : efforts[0]
  const next: AgentChoice = {
    ...choice,
    reasoning,
    by_runner: { ...choice.by_runner, [choice.runner]: { model: choice.model, reasoning } },
  }
  return { choice: next, changed: true }
}

/** Model ids to show, with a stored id that the catalog omitted kept at the end. */
export function modelOptions(catalog: AgentModels, model: string): string[] {
  if (!model || catalog.models.includes(model)) return catalog.models
  return [...catalog.models, model]
}
