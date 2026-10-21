/**
 * The rhombus a decision needs for the copy it holds. Sizing lives with the
 * component that measures the copy (ProcessDiamond); this module is only the
 * arithmetic, so it can be checked without a DOM.
 */

/** The widest a question wraps before the rhombus grows wider instead. */
export const DIAMOND_COPY_WIDTH = 150
/** The narrowest and shortest a rhombus gets, however short the question. */
export const DIAMOND_MIN_WIDTH = 112
export const DIAMOND_MIN_HEIGHT = 96
/**
 * Room for the outline: the polygon is inset 1.5% of the box and stroked 2px, so
 * the copy needs a little clearance from the slants it sits between.
 */
const COPY_PAD = 14

export type DiamondBox = { width: number; height: number }

/**
 * The rhombus a copy of this size needs. The copy is centred, so the rhombus
 * must satisfy `copyWidth/W + copyHeight/H ≤ 1`; giving the copy half the box on
 * each axis makes that sum ≈ 0.96, which is why the box is twice the copy. A
 * flatter rule — the copy plus a constant — fails as soon as the question wraps
 * tall, because the slants close in faster than the text grows.
 */
export function diamondBoxForCopy(copy: DiamondBox): DiamondBox {
  return {
    width: Math.max(DIAMOND_MIN_WIDTH, Math.ceil(2 * copy.width + COPY_PAD)),
    height: Math.max(DIAMOND_MIN_HEIGHT, Math.ceil(2 * copy.height + COPY_PAD)),
  }
}
