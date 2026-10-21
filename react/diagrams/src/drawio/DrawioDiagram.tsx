import { useEffect, useRef, useState } from 'react'
import { Graph, InternalEvent, ModelXmlSerializer, type FitPlugin } from '@maxgraph/core'
import '@maxgraph/core/css/common.css'
import { extractMxGraphModel } from './drawioXml'
import './DrawioDiagram.css'

export type DrawioDiagramProps = {
  xml: string
}

export function DrawioDiagram({ xml }: DrawioDiagramProps) {
  const host = useRef<HTMLDivElement>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    const el = host.current
    if (!el) return
    let graph: Graph | undefined
    setError('')
    el.replaceChildren()
    el.style.minWidth = ''
    const trimmed = xml.trim()
    if (!trimmed) return
    try {
      const modelXml = extractMxGraphModel(trimmed)
      InternalEvent.disableContextMenu(el)
      graph = new Graph(el)
      graph.setEnabled(false)
      graph.setHtmlLabels(true)
      new ModelXmlSerializer(graph.getDataModel()).import(modelXml)
      const bounds = graph.getGraphBounds()
      const scale = graph.getView().getScale() || 1
      el.style.minWidth = `${Math.ceil(bounds.width / scale + 24)}px`
      graph.getPlugin<FitPlugin>('fit')?.fitCenter({ margin: 12 })
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err))
    }
    return () => {
      graph?.destroy()
    }
  }, [xml])

  return (
    <div className="dpd-drawio-shell">
      {error ? <div className="dpd-drawio-error">{error}</div> : null}
      <div className="dpd-drawio-scroll">
        <div ref={host} className="dpd-drawio" hidden={Boolean(error)} />
      </div>
    </div>
  )
}
