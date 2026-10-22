/** Enter sends. Shift+Enter is a newline. IME composition must not send. */
export function composerSubmitKey(event: {
  key: string
  shiftKey: boolean
  keyCode?: number
  nativeEvent?: { isComposing?: boolean }
}): boolean {
  if (event.key !== 'Enter' || event.shiftKey) return false
  if (event.nativeEvent?.isComposing || event.keyCode === 229) return false
  return true
}
