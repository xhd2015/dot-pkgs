import { useEffect, useRef, useState, type ReactNode } from 'react'
import { AgentChatComposer } from './AgentChatComposer'
import { avatarGlyph } from './avatarGlyph'
import { pendingStillWaiting, type AgentChatPending } from './pending'
import './AgentChat.css'

export type AgentChatMessage = { role: 'user' | 'agent'; text: string }

export type { AgentChatPending }

export type AgentChatLayout = 'overlay' | 'page'

export type AgentChatProps = {
  messages: AgentChatMessage[]
  pending?: AgentChatPending[]
  thinking?: boolean
  layout?: AgentChatLayout
  header?: ReactNode
  composer?: ReactNode
  placeholder?: string
  sendLabel?: string
  sendingLabel?: string
  retryLabel?: string
  deliveryFailedLabel?: string
  agentLabel?: string
  userLabel?: string
  renderAvatar?: (role: 'user' | 'agent') => ReactNode
  disabled?: boolean
  className?: string
  renderMessage?: (message: AgentChatMessage) => ReactNode
  afterMessage?: (index: number) => ReactNode
  logFooter?: ReactNode
  onSend?: (text: string) => void
  onRetry?: (id: string) => void
}

export function AgentChat(props: AgentChatProps) {
  const layout = props.layout ?? 'overlay'
  const logRef = useRef<HTMLDivElement>(null)
  const prevCount = useRef(0)
  const [draft, setDraft] = useState('')
  const waiting = pendingStillWaiting(
    props.messages.map((item) => item.text),
    props.pending ?? [],
  )
  const count = props.messages.length + waiting.length + (props.thinking ? 1 : 0)

  useEffect(() => {
    if (count > prevCount.current) {
      const log = logRef.current
      if (log) log.scrollTop = log.scrollHeight
    }
    prevCount.current = count
  }, [count])

  function send() {
    const text = draft.trim()
    if (!text || props.disabled) return
    setDraft('')
    props.onSend?.(text)
  }

  function body(message: AgentChatMessage): ReactNode {
    return props.renderMessage ? props.renderMessage(message) : message.text
  }

  function avatar(role: 'user' | 'agent'): ReactNode {
    if (props.renderAvatar) return props.renderAvatar(role)
    const label = role === 'agent' ? props.agentLabel : props.userLabel
    return <div className={`agent-chat-avatar agent-chat-avatar-${role}`}>{avatarGlyph(role, label)}</div>
  }

  function turn(role: 'user' | 'agent', content: ReactNode, extra?: ReactNode) {
    const name = role === 'agent' ? props.agentLabel : props.userLabel
    return (
      <div className={`agent-chat-turn agent-chat-turn-${role}`}>
        {role === 'agent' ? avatar(role) : null}
        <div className="agent-chat-turn-main">
          {name ? <div className="agent-chat-name">{name}</div> : null}
          <div className={`agent-chat-bubble agent-chat-bubble-${role}`}>{content}</div>
          {extra}
        </div>
        {role === 'user' ? avatar(role) : null}
      </div>
    )
  }

  const rootClass = ['agent-chat', `agent-chat-${layout}`, props.className].filter(Boolean).join(' ')

  return (
    <div className={rootClass}>
      <div className="agent-chat-box">
        {props.header ? <div className="agent-chat-head">{props.header}</div> : null}
        <div className="agent-chat-log" ref={logRef}>
          {props.messages.map((item, index) => (
            <div key={index}>
              {turn(item.role, body(item))}
              {props.afterMessage?.(index)}
            </div>
          ))}
          {props.logFooter}
          {waiting.map((item) => (
            <div key={item.id}>
              {turn(
                'user',
                body({ role: 'user', text: item.text }),
                item.status === 'sending' ? (
                  <div className="agent-chat-status">{props.sendingLabel ?? 'Sending…'}</div>
                ) : item.status === 'failed' ? (
                  <div className="agent-chat-status">
                    {item.error || props.deliveryFailedLabel || 'Delivery failed'}{' '}
                    <button type="button" onClick={() => props.onRetry?.(item.id)}>
                      {props.retryLabel ?? 'Retry'}
                    </button>
                  </div>
                ) : null,
              )}
            </div>
          ))}
          {props.thinking ? <div>{turn('agent', '···')}</div> : null}
        </div>
        <div className="agent-chat-compose">
          {props.composer ?? (
            <>
              <AgentChatComposer
                value={draft}
                placeholder={props.placeholder ?? 'Message'}
                disabled={props.disabled}
                onChange={setDraft}
                onSubmit={send}
              />
              <button
                type="button"
                className="agent-chat-send"
                disabled={props.disabled || draft.trim() === ''}
                onClick={send}
              >
                {props.sendLabel ?? 'Send'}
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
