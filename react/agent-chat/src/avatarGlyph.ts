export function avatarGlyph(role: 'user' | 'agent', label?: string): string {
  const trimmed = label?.trim()
  if (trimmed) {
    const chars = [...trimmed]
    return chars[0] ?? (role === 'agent' ? 'AI' : 'U')
  }
  return role === 'agent' ? 'AI' : 'U'
}
