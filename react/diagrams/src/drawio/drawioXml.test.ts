/**
 * draw.io mxfile unwrap + inflate.
 * Run: bun src/drawio/drawioXml.test.ts
 */
import { deflateSync } from 'fflate'
import { extractMxGraphModel, inflateDrawioPayload } from './drawioXml'

let failed = 0

function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

const inner = '<mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel>'
const packed = btoa(String.fromCharCode(...deflateSync(new TextEncoder().encode(inner))))
assert(inflateDrawioPayload(packed).includes('<mxGraphModel'), 'inflates draw.io deflate+base64')
assert(
  extractMxGraphModel(`<mxfile><diagram id="x">${packed}</diagram></mxfile>`).includes('<mxCell id="0"'),
  'compressed diagram element inflates',
)

assert(extractMxGraphModel(inner) === inner, 'bare mxGraphModel passes through')
assert(
  extractMxGraphModel(`<mxfile><diagram>${inner}</diagram></mxfile>`).startsWith('<mxGraphModel'),
  'plain mxfile unwraps',
)
assert(!extractMxGraphModel(`<mxfile><diagram>${inner}</diagram></mxfile>`).includes('<mxfile'), 'strips mxfile wrapper')

try {
  extractMxGraphModel('not a diagram')
  assert(false, 'junk xml should throw')
} catch {
  assert(true, 'junk xml throws')
}

try {
  extractMxGraphModel('')
  assert(false, 'empty xml should throw')
} catch {
  assert(true, 'empty xml throws')
}

if (failed > 0) {
  console.error(`\n${failed} assertion(s) failed`)
  process.exit(1)
}
console.log('\ndrawioXml: all assertions passed')
