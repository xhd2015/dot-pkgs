import type { Graph } from '@maxgraph/core'
import { fitDrawio, zoomDrawio, type DrawioTransform } from './drawioViewport'

export type DrawioDiagramHandle = {
  fit: () => void
  zoomIn: () => void
  zoomOut: () => void
}

type Point = { x: number; y: number }

export function attachDrawioViewport(
  graph: Graph,
  el: HTMLDivElement,
  mode: 'width' | 'contain',
  interactive: boolean,
): DrawioDiagramHandle & { destroy: () => void } {
  const viewport = el.parentElement!
  let frame = 0
  let destroyed = false
  let size = ''
  let transform: DrawioTransform = { scale: 1, x: 0, y: 0, height: 0 }
  const pointers = new Map<number, Point>()

  function apply(next: DrawioTransform) {
    transform = next
    graph.getView().scaleAndTranslate(next.scale, next.x, next.y)
    el.dataset.drawioScale = String(next.scale)
    el.dataset.drawioX = String(next.x)
    el.dataset.drawioY = String(next.y)
  }

  function fit() {
    if (destroyed) return
    const bounds = graph.getGraphBounds()
    const view = graph.getView()
    const currentScale = view.getScale() || 1
    const next = fitDrawio({
      x: bounds.x / currentScale - view.translate.x,
      y: bounds.y / currentScale - view.translate.y,
      width: bounds.width / currentScale,
      height: bounds.height / currentScale,
    }, viewport.clientWidth, viewport.clientHeight, mode)
    if (!next) return
    if (mode === 'width') el.style.height = `${next.height}px`
    apply(next)
    el.dataset.drawioFitScale = String(next.scale)
  }

  function schedule() {
    if (destroyed) return
    const nextSize = `${viewport.clientWidth}:${mode === 'contain' ? viewport.clientHeight : ''}`
    if (nextSize === size) return
    size = nextSize
    cancelAnimationFrame(frame)
    frame = requestAnimationFrame(fit)
  }

  function center(): Point {
    return { x: viewport.clientWidth / 2, y: viewport.clientHeight / 2 }
  }

  function zoom(factor: number, from = center(), to = from) {
    if (!destroyed) apply(zoomDrawio(transform, factor, from, to))
  }

  function local(point: Point): Point {
    const rect = el.getBoundingClientRect()
    return { x: point.x - rect.left, y: point.y - rect.top }
  }

  function down(event: PointerEvent) {
    if (event.pointerType === 'mouse' && event.button !== 0) return
    pointers.set(event.pointerId, { x: event.clientX, y: event.clientY })
    try { el.setPointerCapture(event.pointerId) } catch { /* pointer may already have ended */ }
    el.classList.add('is-dragging')
    event.preventDefault()
  }

  function move(event: PointerEvent) {
    if (!pointers.has(event.pointerId)) return
    const before = [...pointers.values()]
    const old = pointers.get(event.pointerId)!
    const point = { x: event.clientX, y: event.clientY }
    pointers.set(event.pointerId, point)
    const after = [...pointers.values()]
    if (before.length >= 2) {
      const midpoint = (a: Point, b: Point): Point => ({ x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 })
      const distance = (a: Point, b: Point) => Math.hypot(a.x - b.x, a.y - b.y)
      const previousDistance = distance(before[0], before[1])
      if (previousDistance > 0) {
        zoom(distance(after[0], after[1]) / previousDistance,
          local(midpoint(before[0], before[1])), local(midpoint(after[0], after[1])))
      }
    } else {
      apply({ ...transform,
        x: transform.x + (point.x - old.x) / transform.scale,
        y: transform.y + (point.y - old.y) / transform.scale,
      })
    }
    event.preventDefault()
  }

  function up(event: PointerEvent) {
    pointers.delete(event.pointerId)
    if (pointers.size === 0) el.classList.remove('is-dragging')
  }

  function wheel(event: WheelEvent) {
    event.preventDefault()
    const delta = event.deltaY * (event.deltaMode === 1 ? 16 : event.deltaMode === 2 ? viewport.clientHeight : 1)
    zoom(Math.exp(-Math.max(-500, Math.min(500, delta)) * 0.002), local({ x: event.clientX, y: event.clientY }))
  }

  const observer = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(schedule)
  observer?.observe(viewport)
  window.addEventListener('resize', schedule)
  if (interactive) {
    el.addEventListener('pointerdown', down)
    el.addEventListener('pointermove', move)
    el.addEventListener('pointerup', up)
    el.addEventListener('pointercancel', up)
    el.addEventListener('lostpointercapture', up)
    el.addEventListener('wheel', wheel, { passive: false })
  }
  schedule()

  return {
    fit,
    zoomIn: () => zoom(1.25),
    zoomOut: () => zoom(0.8),
    destroy() {
      destroyed = true
      cancelAnimationFrame(frame)
      observer?.disconnect()
      window.removeEventListener('resize', schedule)
      el.removeEventListener('pointerdown', down)
      el.removeEventListener('pointermove', move)
      el.removeEventListener('pointerup', up)
      el.removeEventListener('pointercancel', up)
      el.removeEventListener('lostpointercapture', up)
      el.removeEventListener('wheel', wheel)
      pointers.clear()
    },
  }
}
