import { useEffect, useId, useMemo, useRef, useState } from 'react'
import mermaid from 'mermaid'
import type { SequenceDiagramModel } from './model'
import { toSequenceMermaid } from './sequenceMermaid'
import { stepMarkPlan } from './sequenceHighlights'
import { colorSequenceDescriptions, colorSequenceMarks } from './sequenceDiagramMarks'
import './SequenceDiagram.css'

mermaid.initialize({
  startOnLoad: false,
  theme: 'default',
  securityLevel: 'loose',
  fontFamily: 'system-ui, -apple-system, sans-serif',
})

export type SequenceDiagramProps = {
  seq: SequenceDiagramModel | null | undefined
  showSource?: boolean
  emptyDescription?: string
}

export function SequenceDiagram({
  seq,
  showSource = false,
  emptyDescription = 'Add participants to render',
}: SequenceDiagramProps) {
  const reactId = useId().replace(/:/g, '')
  const [svg, setSvg] = useState('')
  const [renderErr, setRenderErr] = useState('')
  const [rendering, setRendering] = useState(false)
  const canvasRef = useRef<HTMLDivElement>(null)

  const source = useMemo(() => (seq ? toSequenceMermaid(seq) : ''), [seq])
  const hasParticipants = (seq?.participants || []).length > 0

  useEffect(() => {
    let cancelled = false
    async function render() {
      if (!source.trim() || !seq || !hasParticipants) {
        setSvg('')
        setRenderErr('')
        setRendering(false)
        return
      }
      setRendering(true)
      try {
        const id = `dpd-seq-${reactId}-${Date.now()}`
        const { svg: out } = await mermaid.render(id, source)
        if (!cancelled) {
          setSvg(out)
          setRenderErr('')
        }
      } catch (e) {
        if (!cancelled) {
          setSvg('')
          setRenderErr(e instanceof Error ? e.message : String(e))
        }
      } finally {
        if (!cancelled) setRendering(false)
      }
    }
    void render()
    return () => {
      cancelled = true
    }
  }, [source, seq, hasParticipants, reactId])

  const markCount = useMemo(() => {
    const plan = stepMarkPlan(seq)
    return plan.nodes.length + plan.arrows.filter((s) => s.kind).length + plan.notes.filter((s) => s.kind).length
  }, [seq])
  useEffect(() => {
    const root = canvasRef.current
    if (!root || !svg) return
    colorSequenceDescriptions(root, seq)
    if (markCount > 0) colorSequenceMarks(root, seq)
  }, [svg, seq, markCount])

  if (!hasParticipants) {
    return <div className="dpd-seq-empty">{emptyDescription}</div>
  }

  return (
    <div className="dpd-seq">
      {rendering && !svg ? <div className="dpd-seq-pending">Rendering…</div> : null}
      {renderErr ? <div className="dpd-seq-error">{renderErr}</div> : null}
      {svg ? (
        <div className="dpd-seq-canvas" ref={canvasRef}>
          <div className="dpd-seq-svg" dangerouslySetInnerHTML={{ __html: svg }} />
        </div>
      ) : null}
      {showSource ? (
        <details className="dpd-seq-source">
          <summary>Mermaid source (read-only)</summary>
          <pre>{source || '(empty)'}</pre>
        </details>
      ) : null}
    </div>
  )
}
