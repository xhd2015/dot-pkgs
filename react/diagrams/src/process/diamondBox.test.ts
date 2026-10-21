/**
 * The box a decision rhombus derives from its own copy. Run:
 * bun src/process/diamondBox.test.ts
 */
import {
  DIAMOND_COPY_WIDTH,
  DIAMOND_MIN_HEIGHT,
  DIAMOND_MIN_WIDTH,
  diamondBoxForCopy,
} from './diamondBox'

let failed = 0
function assert(cond: unknown, msg: string) {
  if (!cond) {
    failed++
    console.error('FAIL:', msg)
  } else {
    console.log('ok:', msg)
  }
}

/** A centred copy of w×h sits inside a rhombus of W×H only when w/W + h/H ≤ 1. */
function fit(copy: { width: number; height: number }, box: { width: number; height: number }): number {
  return copy.width / box.width + copy.height / box.height
}

// --- floors ---------------------------------------------------------------
{
  const tiny = diamondBoxForCopy({ width: 20, height: 16 })
  assert(tiny.width === DIAMOND_MIN_WIDTH, `a very short question keeps the minimum width (${tiny.width})`)
  assert(tiny.height === DIAMOND_MIN_HEIGHT, `a very short question keeps the minimum height (${tiny.height})`)
  const oneLine = diamondBoxForCopy({ width: 150, height: 16 })
  assert(
    oneLine.height === DIAMOND_MIN_HEIGHT,
    `a one-line question is ${DIAMOND_MIN_HEIGHT} tall, not a square (${oneLine.height})`,
  )
  assert(oneLine.width === 2 * 150 + 14, `a full-width line sets the box from the copy (${oneLine.width})`)
}

// --- the wrap width is what the box is built from -------------------------
{
  const box = diamondBoxForCopy({ width: DIAMOND_COPY_WIDTH, height: 33 })
  assert(
    box.width === 2 * DIAMOND_COPY_WIDTH + 14,
    `the widest copy gives the widest rhombus (${box.width})`,
  )
}

// --- the slant invariant, including the tall cases the old rule missed ----
{
  const copies = [
    { width: 30, height: 16 },
    { width: 112, height: 16 },
    { width: 150, height: 16 },
    { width: 150, height: 33 },
    { width: 150, height: 49 },
    { width: 150, height: 65 },
    { width: 150, height: 76 },
    { width: 150, height: 98 },
    { width: 150, height: 200 },
    { width: 140, height: 400 },
  ]
  const worst = copies.map((copy) => ({ copy, ratio: fit(copy, diamondBoxForCopy(copy)) }))
  const escaped = worst.filter((item) => item.ratio > 1)
  assert(
    escaped.length === 0,
    `every question stays inside its rhombus${escaped.length ? ' — ' + escaped.map((i) => `${i.copy.width}x${i.copy.height}@${i.ratio.toFixed(2)}`).join(', ') : ''}`,
  )
  const max = Math.max(...worst.map((item) => item.ratio))
  assert(max < 0.99, `the fit keeps a margin from the outline (worst ${max.toFixed(3)})`)
}

// --- the old rule, for contrast: it let tall text through the slants ------
{
  const copy = { width: 150, height: 76 }
  const old = { width: 300, height: copy.height + 64 }
  assert(fit(copy, old) > 1, `the copy-plus-64 rule fails once the text wraps tall (${fit(copy, old).toFixed(2)})`)
  assert(fit(copy, diamondBoxForCopy(copy)) <= 1, 'the derived rule holds for the same question')
}

// --- a square is only ever a coincidence ---------------------------------
{
  const box = diamondBoxForCopy({ width: 150, height: 16 })
  assert(box.height !== box.width, `a one-line question is not drawn as a square (${box.width}x${box.height})`)
}

if (failed) {
  console.error(`${failed} assertion(s) failed`)
  throw new Error(`${failed} assertion(s) failed`)
}
console.log('diamondBox: all assertions passed')
