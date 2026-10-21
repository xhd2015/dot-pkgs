import { useLayoutEffect, useRef, useState } from 'react'
import type { CSSProperties, ReactNode } from 'react'
import { activationProps, type CardActivation } from './cardActivation'
import { DIAMOND_COPY_WIDTH, diamondBoxForCopy, type DiamondBox } from './diamondBox'
import './ProcessDiamond.css'

export type ProcessDiamondProps = CardActivation & {
  label: string
  summary?: string
  /**
   * A fixed box, for a surface that decides the size itself. A surface that
   * leaves these out gets a rhombus sized from its own question.
   */
  width?: number
  height?: number
  /**
   * The card box around the rhombus, when the shape and its nested detail live
   * in one measured box. The rhombus then fills the box's inscribed square.
   */
  wrapperClassName?: string
  /** Skin class names for the rhombus and the copy inside its slants. */
  className?: string
  copyClassName?: string
  titleClassName?: string
  subClassName?: string
  /** The fold control, which reads as part of the question. */
  fold?: ReactNode
  /** Nested detail, drawn beside the rhombus inside the wrapper. */
  children?: ReactNode
  /** Tooltip for the card box, e.g. why a change was made. */
  title?: string
  /** Tooltip for the question itself, e.g. its summary. */
  copyTitle?: string
}

/**
 * A flowchart decision: a rhombus inscribed in its box, so a connector that
 * meets the box edge lands on the shape the reader can see. The outline is an
 * SVG polygon rather than a rotated square because width and height are
 * independent here — a short question is a short rhombus, not a small square.
 *
 * Left to itself the rhombus measures its own copy and sizes itself, which is
 * what keeps a one-line question from reserving a square: the surface around it
 * measures the rendered card, so whatever this draws is what the layout uses.
 */
export function ProcessDiamond({
  label,
  summary,
  width,
  height,
  wrapperClassName,
  className,
  copyClassName,
  titleClassName,
  subClassName,
  fold,
  children,
  title,
  copyTitle,
  onActivate,
}: ProcessDiamondProps) {
  const fixed = width != null || height != null
  const copyRef = useRef<HTMLDivElement>(null)
  const [measured, setMeasured] = useState<DiamondBox | null>(null)

  // The copy sets its own width from the text, never from the box, so measuring
  // it cannot feed back into the size it just produced.
  useLayoutEffect(() => {
    if (fixed) return
    const copy = copyRef.current
    if (!copy) return
    const next = diamondBoxForCopy({ width: copy.offsetWidth, height: copy.offsetHeight })
    setMeasured((prev) =>
      prev && Math.abs(prev.width - next.width) <= 1 && Math.abs(prev.height - next.height) <= 1 ? prev : next,
    )
  })

  const box: CSSProperties | undefined = fixed
    ? { width, height }
    : measured
      ? { width: measured.width, height: measured.height }
      : undefined
  const copyStyle: CSSProperties | undefined = fixed
    ? undefined
    : { width: 'max-content', maxWidth: DIAMOND_COPY_WIDTH, flexShrink: 0 }

  const rhombus = (
    <div className={['pd-branch', className].filter(Boolean).join(' ')} style={box} title={copyTitle}>
      <svg className="pd-branch-outline" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden>
        <polygon points="50,1.5 98.5,50 50,98.5 1.5,50" />
      </svg>
      <div
        ref={copyRef}
        className={['pd-branch-copy', copyClassName].filter(Boolean).join(' ')}
        style={copyStyle}
        {...activationProps({ onActivate })}
      >
        <div className={['pd-branch-title', titleClassName].filter(Boolean).join(' ')}>{label}</div>
        {summary ? <div className={['pd-branch-sub', subClassName].filter(Boolean).join(' ')}>{summary}</div> : null}
        {fold}
      </div>
    </div>
  )
  if (!wrapperClassName) return rhombus
  return (
    <div className={wrapperClassName} title={title}>
      {rhombus}
      {children}
    </div>
  )
}
