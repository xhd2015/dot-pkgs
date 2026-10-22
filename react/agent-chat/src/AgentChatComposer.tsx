import { composerSubmitKey } from './composerKey'

export function AgentChatComposer(props: {
  value: string
  placeholder?: string
  disabled?: boolean
  onChange: (value: string) => void
  onSubmit: () => void
}) {
  return (
    <textarea
      className="agent-chat-input"
      value={props.value}
      placeholder={props.placeholder}
      disabled={props.disabled}
      rows={3}
      onChange={(event) => props.onChange(event.target.value)}
      onKeyDown={(event) => {
        if (!composerSubmitKey(event)) return
        event.preventDefault()
        props.onSubmit()
      }}
    />
  )
}
