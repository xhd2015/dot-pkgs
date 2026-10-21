import type { ReactNode } from 'react'
import type { APIFlowchartNode } from './model'
import { markLabel, markOf } from './marks'
import { ProcessDiamond } from './ProcessDiamond'
import { ProcessStep } from './ProcessStep'
// The product skin for both shapes: whichever surface draws this card gets it.
import './ProcessNodeCard.css'

/** The skin the product pages draw: antd greys and blues. */
const SKIN = {
  step: { body: 'pf-head', title: 'pf-ttl', sub: 'pf-sub' },
  diamond: { copy: undefined, title: 'pf-ttl', sub: 'pf-sub' },
}

export type ProcessNodeCardProps = {
  node: APIFlowchartNode
  /** Whether this node holds nested detail, i.e. whether it can be expanded. */
  hasDetail?: boolean
  expanded?: boolean
  /** Called when the reader clicks the card or presses Enter/Space on it. */
  onActivate?: () => void
  /** Nested detail, drawn under a step's row or inside a diamond's copy. */
  children?: ReactNode
}

/**
 * One card of the process diagram the product pages draw: a step box or a
 * decision rhombus, with the node's kind, its change mark and its fold control.
 * The two shapes themselves come from ProcessStep and ProcessDiamond, so the
 * fixture gallery draws the same geometry.
 */
export function ProcessNodeCard({ node, hasDetail = false, expanded = false, onActivate, children }: ProcessNodeCardProps) {
  const kind = node.kind || 'step'
  const mark = markOf(node.change)
  const boxClass = [
    'pf-node',
    hasDetail ? 'pf-has' : '',
    kind === 'response' ? 'pf-response' : '',
    kind === 'branch' ? 'pf-decision' : '',
    expanded ? 'pf-expanded' : '',
    mark.className,
  ]
    .filter(Boolean)
    .join(' ')

  if (kind === 'branch') {
    return (
      <ProcessDiamond
        wrapperClassName={boxClass}
        titleClassName={SKIN.diamond.title}
        subClassName={SKIN.diamond.sub}
        label={node.label}
        summary={node.summary}
        title={mark.title || undefined}
        copyTitle={node.summary || undefined}
        onActivate={hasDetail ? onActivate : undefined}
        fold={hasDetail ? <div className="pf-fold">{expanded ? '▾ hide detail' : '⊕ detail'}</div> : null}
      >
        {children}
      </ProcessDiamond>
    )
  }

  return (
    <ProcessStep
      className={boxClass}
      bodyClassName={SKIN.step.body}
      titleClassName={SKIN.step.title}
      subClassName={SKIN.step.sub}
      label={node.label}
      summary={node.summary}
      title={mark.title || undefined}
      onActivate={hasDetail ? onActivate : undefined}
      leading={
        <div className="pf-icon">{kind === 'response' ? '✓' : hasDetail ? (expanded ? '▾' : '⊕') : '•'}</div>
      }
      trailing={
        kind !== 'step' || mark.kind ? (
          <div className="pf-pills">
            {kind !== 'step' ? <div className="pf-pill">{kind}</div> : null}
            {mark.kind ? <div className={`pf-pill pf-pill-mark ${mark.className}`}>{markLabel(mark.kind)}</div> : null}
          </div>
        ) : null
      }
    >
      {children}
    </ProcessStep>
  )
}
