/**
 * Pull an mxGraphModel out of a draw.io mxfile (plain or deflate+base64).
 * Run: bun src/drawio/drawioXml.test.ts
 */
import { inflateSync } from 'fflate'

const MODEL_OPEN = '<mxGraphModel'
const MODEL_CLOSE = '</mxGraphModel>'

export function extractMxGraphModel(xml: string): string {
  const trimmed = (xml || '').trim()
  if (!trimmed) throw new Error('empty diagram')
  const fromTag = sliceModel(trimmed)
  if (fromTag) return fromTag
  const inner = diagramPayload(trimmed)
  if (!inner) throw new Error('no mxGraphModel in draw.io file')
  if (inner.startsWith('<')) {
    const nested = sliceModel(inner) || (inner.includes(MODEL_OPEN) ? inner : '')
    if (nested) return nested
  }
  const inflated = inflateDrawioPayload(inner)
  const model = sliceModel(inflated) || inflated
  if (!model.includes(MODEL_OPEN)) throw new Error('inflated diagram is not an mxGraphModel')
  return model
}

function sliceModel(xml: string): string {
  const start = xml.indexOf(MODEL_OPEN)
  if (start < 0) return ''
  const end = xml.lastIndexOf(MODEL_CLOSE)
  if (end < start) return ''
  return xml.slice(start, end + MODEL_CLOSE.length)
}

function diagramPayload(xml: string): string {
  const match = /<diagram\b[^>]*>([\s\S]*?)<\/diagram>/i.exec(xml)
  return match ? match[1].trim() : ''
}

export function inflateDrawioPayload(payload: string): string {
  const compact = payload.replace(/\s+/g, '')
  const binary = Uint8Array.from(atob(compact), (c) => c.charCodeAt(0))
  let text = new TextDecoder().decode(inflateSync(binary))
  if (text.includes('%3C') || text.includes('%3c')) {
    try {
      text = decodeURIComponent(text)
    } catch {
      /* keep raw inflate */
    }
  }
  return text.trim()
}
