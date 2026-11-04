export type DrawioFitMode = 'legacy' | 'width' | 'contain'
export type DrawioBounds = { x: number; y: number; width: number; height: number }
export type DrawioTransform = { scale: number; x: number; y: number; height: number }

export function fitDrawio(
  bounds: DrawioBounds,
  width: number,
  height: number,
  mode: 'width' | 'contain',
  padding = 12,
): DrawioTransform | null {
  if (![bounds.x, bounds.y, bounds.width, bounds.height, width, height, padding].every(Number.isFinite)) return null
  if (bounds.width <= 0 || bounds.height <= 0 || width <= padding * 2 || padding < 0) return null
  if (mode === 'contain' && height <= padding * 2) return null
  const scale = Math.min(1, (width - padding * 2) / bounds.width,
    mode === 'contain' ? (height - padding * 2) / bounds.height : Infinity)
  const fittedHeight = mode === 'width' ? Math.ceil(bounds.height * scale + padding * 2) : height
  return {
    scale,
    x: (width / scale - bounds.width) / 2 - bounds.x,
    y: (fittedHeight / scale - bounds.height) / 2 - bounds.y,
    height: fittedHeight,
  }
}

export function zoomDrawio(
  transform: DrawioTransform,
  factor: number,
  from: { x: number; y: number },
  to = from,
): DrawioTransform {
  if (!Number.isFinite(factor) || factor <= 0) return transform
  const scale = Math.max(0.005, Math.min(8, transform.scale * factor))
  return {
    ...transform,
    scale,
    x: to.x / scale - (from.x / transform.scale - transform.x),
    y: to.y / scale - (from.y / transform.scale - transform.y),
  }
}
