import assert from 'node:assert/strict'
import { fitDrawio, zoomDrawio } from './drawioViewport'

function near(actual: number, expected: number) {
  assert.ok(Math.abs(actual - expected) < 0.000001, `${actual} != ${expected}`)
}

const wide = { x: 25, y: 90, width: 1000, height: 500 }
for (const width of [288, 320, 366, 390, 430, 800, 1440]) {
  const fit = fitDrawio(wide, width, 0, 'width')!
  assert.ok(fit)
  assert.ok(fit.scale <= 1, 'inline never automatically enlarges small graphs')
  assert.ok(wide.width * fit.scale <= width - 24 + 0.000001)
  assert.equal(fit.height, Math.ceil(wide.height * fit.scale + 24))
  near((wide.x + fit.x) * fit.scale, (width - wide.width * fit.scale) / 2)
  assert.ok((wide.y + fit.y) * fit.scale >= 12 - 0.000001)
}

const small = fitDrawio({ x: -30, y: -20, width: 100, height: 40 }, 390, 0, 'width')!
assert.equal(small.scale, 1)
assert.equal(small.height, 64)
near((-30 + small.x) * small.scale, 145)
near((-20 + small.y) * small.scale, 12)

for (const bounds of [wide, { x: -500, y: -300, width: 200, height: 4000 }]) {
  for (const [width, height] of [[366, 600], [800, 300], [1160, 700]]) {
    const fit = fitDrawio(bounds, width, height, 'contain')!
    assert.ok(bounds.width * fit.scale <= width - 24 + 0.000001)
    assert.ok(bounds.height * fit.scale <= height - 24 + 0.000001)
    assert.equal(fit.height, height)
    near((bounds.x + fit.x) * fit.scale, (width - bounds.width * fit.scale) / 2)
    near((bounds.y + fit.y) * fit.scale, (height - bounds.height * fit.scale) / 2)
  }
}

for (const bounds of [
  { ...wide, width: 0 }, { ...wide, height: -1 }, { ...wide, x: NaN }, { ...wide, width: Infinity },
]) assert.equal(fitDrawio(bounds, 390, 600, 'contain'), null)
assert.equal(fitDrawio(wide, 0, 600, 'width'), null, 'hidden disclosure waits for a real width')
assert.equal(fitDrawio(wide, 24, 600, 'width'), null)
assert.equal(fitDrawio(wide, 390, 0, 'contain'), null, 'unmeasured modal waits for its layout')
assert.equal(fitDrawio(wide, 390, 600, 'width', -1), null)

const original = fitDrawio(wide, 390, 600, 'contain')!
const point = { x: 120, y: 220 }
const zoomed = zoomDrawio(original, 1.25, point)
assert.ok(zoomed.scale > original.scale)
near(point.x / zoomed.scale - zoomed.x, point.x / original.scale - original.x)
near(point.y / zoomed.scale - zoomed.y, point.y / original.scale - original.y)
const restored = zoomDrawio(zoomed, 0.8, point)
near(restored.scale, original.scale)
near(restored.x, original.x)
near(restored.y, original.y)

const movedCenter = { x: 150, y: 260 }
const pinched = zoomDrawio(original, 1.5, point, movedCenter)
near(movedCenter.x / pinched.scale - pinched.x, point.x / original.scale - original.x)
near(movedCenter.y / pinched.scale - pinched.y, point.y / original.scale - original.y)
assert.equal(zoomDrawio(original, 1e9, point).scale, 8)
assert.equal(zoomDrawio(original, 1e-9, point).scale, 0.005)
assert.equal(zoomDrawio(original, NaN, point), original)
assert.equal(zoomDrawio(original, -1, point), original)

console.log('Drawio viewport: width/contain fit, small/large/hidden graphs, normalized origins and zoom/pinch anchors passed.')
