import type { CSSProperties, ReactNode } from 'react'
import { activationProps, type CardActivation } from './cardActivation'

export type ProcessStepProps = CardActivation & {
  label: string
  summary?: string
  /**
   * The box the layout measured. Omitted when the skin sizes the card from its
   * own CSS instead, which is what the product pages do.
   */
  width?: number
  height?: number
  /** Skin class names: the box, the row inside it, and the two lines of copy. */
  className?: string
  bodyClassName?: string
  titleClassName?: string
  subClassName?: string
  /** A kind glyph before the copy, and the chips after it. */
  leading?: ReactNode
  trailing?: ReactNode
  /** Nested detail, drawn under the row. */
  children?: ReactNode
  title?: string
}

/**
 * A flowchart step: a rounded box, its label, an optional caption, and whatever
 * a surface hangs off it — a kind glyph, change chips, nested detail. The caller
 * chooses the box (or lets its own CSS size the card) and supplies the skin.
 */
export function ProcessStep({
  label,
  summary,
  width,
  height,
  className,
  bodyClassName,
  titleClassName,
  subClassName,
  leading,
  trailing,
  children,
  title,
  onActivate,
}: ProcessStepProps) {
  const box: CSSProperties | undefined = width != null || height != null ? { width, height } : undefined
  return (
    <article className={['pd-step', className].filter(Boolean).join(' ')} style={box} title={title}>
      <div className={['pd-step-body', bodyClassName].filter(Boolean).join(' ')} {...activationProps({ onActivate })}>
        {leading}
        <div className="pd-step-copy">
          <div className={['pd-step-title', titleClassName].filter(Boolean).join(' ')}>{label}</div>
          {summary ? <div className={['pd-step-sub', subClassName].filter(Boolean).join(' ')}>{summary}</div> : null}
        </div>
        {trailing}
      </div>
      {children}
    </article>
  )
}
