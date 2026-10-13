import { test, expect, type Page } from '@playwright/test'

test('down increases Y in the keyboard controls, jog pad and plot', async ({
  page,
  request,
}) => {
  const state = await (await request.get('/api/state')).json()
  const y = state.axes[1].position
  const marker = page.getByRole('button', {
    name: 'Drag winch position to set X and Y targets',
  })
  await marker.focus()
  await page.keyboard.press('ArrowDown')
  await expect(page.getByLabel('Y target metres')).toHaveValue(
    (y + 0.01).toFixed(3),
  )
  await page.keyboard.press('ArrowUp')
  await expect(page.getByLabel('Y target metres')).toHaveValue(y.toFixed(3))
  const dot = page.getByTestId('position-drag-handle')
  const before =
    (await dot.boundingBox())!.y -
    (await page.locator('.position-plot').boundingBox())!.y
  await page.getByRole('tab', { name: 'Manual jog' }).click()
  const down = page.locator('.jog-pad .down button')
  await expect(down).toHaveAccessibleName('Jog Y positive')
  await expect(page.locator('.jog-pad .up button')).toHaveAccessibleName(
    'Jog Y negative',
  )
  await down.click()
  await expect(page.getByTestId('y-position')).toHaveText((y + 0.01).toFixed(3))
  await expect
    .poll(
      async () =>
        (await dot.boundingBox())!.y -
        (await page.locator('.position-plot').boundingBox())!.y,
    )
    .toBeGreaterThan(before)
})

test('touch drag previews and submits a target on release', async ({
  page,
  request,
}) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.locator('.position-plot').scrollIntoViewIfNeeded()
  const state = await (await request.get('/api/state')).json()
  const start = await plotPoint(
    page,
    state.axes[0].position,
    state.axes[1].position,
  )
  const end = await plotPoint(
    page,
    state.axes[0].position + 0.5,
    state.axes[1].position + 0.5,
  )
  let moves = 0
  await page.route('**/api/move', async (route) => {
    moves++
    await route.fulfill({
      contentType: 'application/json',
      body: '{"ok":true}',
    })
  })
  const session = await page.context().newCDPSession(page)
  await session.send('Emulation.setTouchEmulationEnabled', { enabled: true })
  await session.send('Input.dispatchTouchEvent', {
    type: 'touchStart',
    touchPoints: [{ x: start.x, y: start.y }],
  })
  await session.send('Input.dispatchTouchEvent', {
    type: 'touchMove',
    touchPoints: [{ x: end.x, y: end.y }],
  })
  await expect(page.getByText(/Release to move both winches/)).toBeVisible()
  expect(moves).toBe(0)
  await session.send('Input.dispatchTouchEvent', {
    type: 'touchEnd',
    touchPoints: [],
  })
  await expect.poll(() => moves).toBe(1)
})

test('lost telemetry cancels a drag before pointer release', async ({
  page,
  request,
}) => {
  const state = await (await request.get('/api/state')).json()
  const start = await plotPoint(
    page,
    state.axes[0].position,
    state.axes[1].position,
  )
  let moves = 0
  page.on('request', (r) => {
    if (r.url().endsWith('/api/move') && r.method() === 'POST') moves++
  })
  await page.mouse.move(start.x, start.y)
  await page.mouse.down()
  await page.mouse.move(start.x + 30, start.y - 20)
  await expect(page.getByText(/Release to move both winches/)).toBeVisible()
  await page.route('**/api/state', (route) => route.abort())
  await expect(page.getByRole('alert')).toContainText('Connection lost')
  await page.mouse.up()
  await expect(page.locator('.live-marker')).not.toHaveClass(/dragging/)
  expect(moves).toBe(0)
})

async function plotPoint(page: Page, x: number, y: number) {
  return page.locator('.position-plot').evaluate(
    (svg, point) => {
      const matrix = (svg as SVGSVGElement).getScreenCTM()!
      const screen = new DOMPoint(
        64 +
          (point.x / 50) * ((svg as SVGSVGElement).viewBox.baseVal.width - 128),
        64 + (point.y / 30) * 280,
      ).matrixTransform(matrix)
      return { x: screen.x, y: screen.y }
    },
    { x, y },
  )
}

test.beforeEach(async ({ request, page }) => {
  await request.post('/api/stop', { data: {} })
  await page.goto('/')
  await expect(
    page.getByRole('button', { name: /Move to position/ }),
  ).toBeEnabled()
})

test('sea-cage context and dragging preview then move both axes on release', async ({
  page,
  request,
}) => {
  await expect(
    page.getByRole('region', { name: 'Sea-cage installation context' }),
  ).toBeVisible()
  const state = await (await request.get('/api/state')).json()
  const x = state.axes[0].position,
    y = state.axes[1].position
  const destination = {
    x: Number((x + 0.5).toFixed(3)),
    y: Number((y + 0.5).toFixed(3)),
  }
  const start = await plotPoint(page, x, y)
  const end = await plotPoint(page, destination.x, destination.y)
  let moves = 0
  page.on('request', (r) => {
    if (r.url().endsWith('/api/move') && r.method() === 'POST') moves++
  })
  await page.mouse.move(start.x, start.y)
  await page.mouse.down()
  await page.mouse.move(end.x, end.y, { steps: 5 })
  await expect(page.getByText(/Release to move both winches/)).toBeVisible()
  await expect(page.getByLabel('X target metres')).toHaveValue(
    destination.x.toFixed(3),
  )
  expect(moves).toBe(0)
  expect(
    (await (await request.get('/api/state')).json()).axes[0].position,
  ).toBe(x)
  await page.mouse.up()
  await expect.poll(() => moves).toBe(1)
  await expect(page.getByTestId('x-position')).toHaveText(
    destination.x.toFixed(3),
    { timeout: 10000 },
  )
  await expect(page.getByTestId('y-position')).toHaveText(
    destination.y.toFixed(3),
    { timeout: 10000 },
  )
  await expect(
    page.getByRole('button', { name: /Move to position/ }),
  ).toBeEnabled()
  await page.screenshot({
    path: 'test-results/sea-cage-desktop.png',
    fullPage: true,
  })
})

test('Escape cancels a drag without issuing a move', async ({
  page,
  request,
}) => {
  const state = await (await request.get('/api/state')).json()
  const start = await plotPoint(
    page,
    state.axes[0].position,
    state.axes[1].position,
  )
  let moves = 0
  page.on('request', (r) => {
    if (r.url().endsWith('/api/move') && r.method() === 'POST') moves++
  })
  await page.mouse.move(start.x, start.y)
  await page.mouse.down()
  await page.mouse.move(start.x + 30, start.y - 20)
  await expect(page.getByText(/Release to move both winches/)).toBeVisible()
  await page.keyboard.press('Escape')
  await page.mouse.up()
  await expect(page.locator('.live-marker')).not.toHaveClass(/dragging/)
  expect(moves).toBe(0)
})

test('mobile drag uses SVG coordinates and clamps to configured travel', async ({
  page,
  request,
}) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.locator('.position-plot').scrollIntoViewIfNeeded()
  const state = await (await request.get('/api/state')).json()
  const start = await plotPoint(
    page,
    state.axes[0].position,
    state.axes[1].position,
  )
  const end = await plotPoint(page, 60, 36)
  let target: unknown
  await page.route('**/api/move', async (route) => {
    target = route.request().postDataJSON().targets
    await route.fulfill({
      contentType: 'application/json',
      body: '{"ok":true}',
    })
  })
  await page.mouse.move(start.x, start.y)
  await page.mouse.down()
  await page.mouse.move(end.x, end.y, { steps: 5 })
  await expect(page.getByLabel('X target metres')).toHaveValue('50.000')
  await expect(page.getByLabel('Y target metres')).toHaveValue('30.000')
  await page.mouse.up()
  await expect.poll(() => target).toEqual({ x: 50, y: 30 })
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBeTruthy()
  await page.screenshot({
    path: 'test-results/sea-cage-mobile.png',
    fullPage: true,
  })
})
