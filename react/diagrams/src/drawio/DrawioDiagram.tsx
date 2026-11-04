import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react'
import { Graph, InternalEvent, ModelXmlSerializer, type FitPlugin } from '@maxgraph/core'
import '@maxgraph/core/css/common.css'
import { extractMxGraphModel } from './drawioXml'
import type { DrawioFitMode } from './drawioViewport'
import { attachDrawioViewport, type DrawioDiagramHandle } from './drawioViewportController'
import './DrawioDiagram.css'

export type DrawioDiagramProps = {
  xml: string
  fitMode?: DrawioFitMode
  interactive?: boolean
  ariaLabel?: string
  onReady?: (ready: boolean) => void
}

export const DrawioDiagram = forwardRef<DrawioDiagramHandle, DrawioDiagramProps>(function DrawioDiagram(
  { xml, fitMode = 'legacy', interactive = false, ariaLabel = 'Drawio 图表', onReady }, ref,
) {
  const host = useRef<HTMLDivElement>(null)
  const controls = useRef<DrawioDiagramHandle | null>(null)
  const [error, setError] = useState('')

  useImperativeHandle(ref, () => ({
    fit: () => controls.current?.fit(),
    zoomIn: () => controls.current?.zoomIn(),
    zoomOut: () => controls.current?.zoomOut(),
  }), [])

  useEffect(() => {
    const el = host.current
    if (!el) return
    let graph: Graph | undefined
    let viewport: ReturnType<typeof attachDrawioViewport> | undefined
    onReady?.(false)
    setError('')
    el.replaceChildren()
    el.style.minWidth = ''
    el.style.height = ''
    delete el.dataset.drawioScale
    const trimmed = xml.trim()
    if (!trimmed) return
    try {
      const modelXml = extractMxGraphModel(trimmed)
      InternalEvent.disableContextMenu(el)
      graph = new Graph(el)
      graph.setEnabled(false)
      graph.setHtmlLabels(true)
      new ModelXmlSerializer(graph.getDataModel()).import(modelXml)
      if (fitMode === 'legacy') {
        const bounds = graph.getGraphBounds()
        const scale = graph.getView().getScale() || 1
        el.style.minWidth = `${Math.ceil(bounds.width / scale + 24)}px`
        graph.getPlugin<FitPlugin>('fit')?.fitCenter({ margin: 12 })
      } else {
        viewport = attachDrawioViewport(graph, el, fitMode, interactive)
        controls.current = viewport
      }
      onReady?.(true)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err))
    }
    return () => {
      controls.current = null
      viewport?.destroy()
      graph?.destroy()
    }
  }, [xml, fitMode, interactive, onReady])

  return (
    <div className="dpd-drawio-shell" data-fit={fitMode}>
      {error ? <div className="dpd-drawio-error" role="alert">{error}</div> : null}
      <div className="dpd-drawio-scroll">
        <div ref={host} className={interactive ? 'dpd-drawio is-interactive' : 'dpd-drawio'}
          hidden={Boolean(error)} role="img" aria-label={ariaLabel} />
      </div>
    </div>
  )
})
