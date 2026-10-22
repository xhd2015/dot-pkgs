export type AgentChatPending = {
  id: string
  text: string
  status: 'sending' | 'failed'
  error?: string
  seen?: number
}

export type PendingMessage = AgentChatPending

// pendingStillWaiting drops a sending item once the server has more copies of
// its text than it did when the send started. A failed item stays for Retry.
export function pendingStillWaiting<T extends { text: string; status: string; seen?: number }>(
  serverTexts: string[],
  pending: T[],
): T[] {
  const counts = new Map<string, number>()
  for (const text of serverTexts) counts.set(text, (counts.get(text) || 0) + 1)
  return pending.filter((item) => {
    if (item.status !== 'sending') return true
    return (counts.get(item.text) || 0) <= (item.seen || 0)
  })
}
