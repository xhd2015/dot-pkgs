import type { Sequence } from './model'
import { isRemoved } from './sequenceHighlights'

function slugID(raw: string): string {
  const s = raw
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '_')
    .replace(/^_+|_+$/g, '')
  if (!s) return 'p'
  if (/^\d/.test(s)) return `p_${s}`
  return s
}

/**
 * Sanitize a participant alias or message for Mermaid sequenceDiagram source.
 * Semicolon is a statement terminator (even inside quotes), which is what broke
 * story 269's Typical Path, so it becomes a fullwidth semicolon.
 */
function escapeLabel(s: string): string {
  return s.replace(/\n/g, ' ').replace(/"/g, "'").replace(/;/g, '；')
}

/** API name and description on one line. The description is coloured after render. */
function messageLabel(api: string | undefined, description: string | undefined, text: string | undefined): string {
  const method = escapeLabel(api || '')
  const detail = escapeLabel(description || '')
  if (method && detail) return `${method} ${detail}`
  if (method) return method
  if (detail) return detail
  return escapeLabel(text || '')
}

/**
 * Compile sequence JSON → Mermaid sequenceDiagram (render-only).
 *
 * A removed step is the one change the source has to carry, as a crossed arrow.
 * Every other mark is colour, which the diagram pass applies to the drawn elements
 * after rendering.
 */
export function toSequenceMermaid(seq: Sequence): string {
  const lines: string[] = ['sequenceDiagram']
  if (seq.autonumber) lines.push('  autonumber')
  for (const p of seq.participants || []) {
    const id = slugID(p.id)
    const label = escapeLabel(p.label)
    if (p.kind === 'actor') lines.push(`  actor ${id} as ${label}`)
    else lines.push(`  participant ${id} as ${label}`)
  }
  for (const m of seq.messages || []) {
    if (m.fragment === 'alt') {
      lines.push(`  alt ${escapeLabel(m.guard || '')}`)
      continue
    }
    if (m.fragment === 'else') {
      lines.push(`  else ${escapeLabel(m.guard || '')}`)
      continue
    }
    if (m.fragment === 'end') {
      lines.push('  end')
      continue
    }
    const from = slugID(m.from)
    const to = slugID(m.to)
    const text = messageLabel(m.api, m.description, m.text)
    // A self-message with no method name is an annotation. One that names an API
    // is a call on that lifeline, drawn as an arrow and numbered with the others.
    if (from === to && !(m.api || '').trim()) {
      lines.push(`  Note over ${from}: ${text}`)
      continue
    }
    const dashed = m.arrow === 'dashed'
    let arrow = dashed ? '-->>' : '->>'
    if (isRemoved(m)) arrow = dashed ? '--x' : '-x'
    lines.push(`  ${from}${arrow}${to}: ${text}`)
  }
  return lines.join('\n') + '\n'
}
