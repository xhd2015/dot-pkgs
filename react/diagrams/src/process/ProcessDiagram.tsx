import { useState } from 'react'
import { ProcessGraph } from './ProcessGraph'
import type { APIFlowchart } from './model'
import { crumbsFor, revealPath } from './nesting'
import './ProcessDiagram.css'

export type ProcessDiagramProps = {
  /** The process model to draw, including any nested breakdowns. */
  chart: APIFlowchart
  /** Card mode: no breadcrumb and no hint, for a diagram inside a card. */
  embedded?: boolean
  /**
   * Open on the first chart that carries a change mark. On by default because
   * that is what a reader of a technical design wants to see first.
   */
  revealChanges?: boolean
  /**
   * Which nested node is open. Controlled when given; the diagram keeps its own
   * state otherwise.
   */
  expandPath?: string[]
  onExpandPath?: (next: string[]) => void
}

/**
 * A process model drawn as a diagram. This is the entry point: it owns the
 * breadcrumb, the scroll frame and which branch is open, and hands the drawing
 * itself to ProcessGraph.
 */
export function ProcessDiagram({
  chart,
  embedded = false,
  revealChanges = true,
  expandPath,
  onExpandPath,
}: ProcessDiagramProps) {
  const [ownPath, setOwnPath] = useState<string[]>(() => (revealChanges ? (revealPath(chart) ?? []) : []))
  const open = expandPath ?? ownPath
  const setOpen = onExpandPath ?? setOwnPath

  if (!(chart.nodes || []).length) {
    return <div className="pf-empty">No process flowchart yet</div>
  }

  const crumbs = crumbsFor(chart, open)

  return (
    <div className={embedded ? 'pf-shell pf-embed' : 'pf-shell'}>
      {embedded ? null : (
        <div className="pf-crumbs">
          {crumbs.map((label, i) => (
            <span key={`${label}-${i}`} className="pf-crumb-item">
              {i > 0 ? <span className="pf-sep">/</span> : null}
              <button
                type="button"
                className={i === crumbs.length - 1 ? 'pf-crumb current' : 'pf-crumb'}
                onClick={() => setOpen(open.slice(0, i))}
              >
                {label}
              </button>
            </span>
          ))}
        </div>
      )}
      <div className="pf-canvas">
        <ProcessGraph chart={chart} pathPrefix={[]} expandPath={open} onExpandPath={setOpen} />
      </div>
      {embedded ? null : (
        <div className="pf-hint">
          {open.length
            ? `Expanded · ${crumbs.join(' / ')}`
            : 'Click ⊕ on a node with nested detail to expand in place'}
        </div>
      )}
    </div>
  )
}
