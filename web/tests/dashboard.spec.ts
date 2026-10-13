import { test, expect } from '@playwright/test'

test.beforeEach(async ({ request, page }) => {
  await request.post('/api/stop', { data: {} })
  await page.goto('/')
  await expect(
    page.getByText('Controller connected', { exact: true }),
  ).toBeVisible()
  await expect(
    page.getByRole('button', { name: /Move to position/ }),
  ).toBeEnabled()
})

test('simulator target move and live telemetry', async ({ page }) => {
  await expect(page.getByText('SIMULATION MODE', { exact: true })).toBeVisible()
  await page.getByLabel('X target metres').fill('2.500')
  await page.getByLabel('Y target metres').fill('4.850')
  await page.getByRole('button', { name: /Move to position/ }).click()
  await expect(page.getByTestId('x-position')).toHaveText('2.500', {
    timeout: 10000,
  })
  await expect(page.getByTestId('y-position')).toHaveText('4.850', {
    timeout: 10000,
  })
  await expect(
    page.getByRole('button', { name: /Move to position/ }),
  ).toBeEnabled()
  await page.screenshot({
    path: 'test-results/dashboard-desktop.png',
    fullPage: true,
  })
})

test('invalid travel rejected without movement', async ({ page, request }) => {
  await page.getByLabel('X target metres').fill('51')
  await expect(
    page.getByRole('button', { name: /Move to position/ }),
  ).toBeDisabled()
  await expect(page.getByRole('alert')).toContainText('configured travel range')
  const response = await request.post('/api/move', {
    data: { targets: { x: 51 }, speed: 0.1 },
  })
  expect(response.status()).toBe(409)
})

test('incremental jogging moves exactly one selected step', async ({
  page,
}) => {
  const before = Number(await page.getByTestId('x-position').textContent())
  await page.getByRole('tab', { name: 'Manual jog' }).click()
  await page.getByLabel('Jog step').selectOption('0.01')
  await page.getByRole('button', { name: 'Jog X positive' }).click()
  await expect(page.getByTestId('x-position')).toHaveText(
    (before + 0.01).toFixed(3),
  )
})

test('hold survives telemetry updates and release stops motion', async ({
  page,
}) => {
  await page.getByRole('tab', { name: 'Manual jog' }).click()
  await page.getByRole('button', { name: 'Hold to move', exact: true }).click()
  const before = Number(await page.getByTestId('x-position').textContent())
  const button = page.getByRole('button', { name: 'Jog X positive' })
  await expect(button).toBeEnabled()
  const box = (await button.boundingBox())!
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
  await page.mouse.down()
  await expect
    .poll(async () =>
      Number(await page.getByTestId('x-position').textContent()),
    )
    .toBeGreaterThan(before + 0.07)
  await page.mouse.up()
  await expect(button).toBeEnabled()
  await expect
    .poll(
      async () =>
        (await (await page.request.get('/api/state')).json()).axes[0].state,
    )
    .toBe('idle')
  const stopped = await page.getByTestId('x-position').textContent()
  await page.waitForTimeout(600)
  await expect(page.getByTestId('x-position')).toHaveText(stopped!)
})

test('abandoned jog lease expires and cannot be renewed', async ({
  request,
}) => {
  const result = await request.post('/api/jog', {
    data: { axis: 'y', direction: 1, speed: 0.1 },
  })
  expect(result.ok()).toBeTruthy()
  const { lease } = await result.json()
  await expect
    .poll(
      async () =>
        (await (await request.get('/api/state')).json()).axes[1].state,
    )
    .toBe('idle')
  expect(
    (await request.post('/api/jog/renew', { data: { lease } })).status(),
  ).toBe(409)
})

test('mocked API errors are visible and offline controls are disabled', async ({
  page,
}) => {
  await page.route('**/api/move', (route) =>
    route.fulfill({
      status: 409,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'Encoder fault on X axis' }),
    }),
  )
  await page.getByRole('button', { name: /Move to position/ }).click()
  await expect(page.getByRole('alert')).toContainText('Encoder fault')
  await page.route('**/api/state', (route) => route.abort())
  await expect(
    page.getByText(
      'Connection lost. Positions are stale; movement controls are unavailable.',
    ),
  ).toBeVisible()
  await expect(
    page.getByRole('button', { name: /Move to position/ }),
  ).toBeDisabled()
})

test('mobile layout fits and exposes all motion controls', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await expect(
    page.getByRole('heading', { name: 'Control station.' }),
  ).toBeVisible()
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBeTruthy()
  await page.getByRole('tab', { name: 'Manual jog' }).click()
  await expect(
    page.getByRole('button', { name: 'Jog Y positive' }),
  ).toBeVisible()
  await page.screenshot({
    path: 'test-results/dashboard-mobile.png',
    fullPage: true,
  })
  const nav = page.getByRole('navigation', { name: 'Phone views' })
  await nav.getByRole('button', { name: 'home', exact: true }).click()
  await expect(
    page.getByRole('region', { name: 'Home position' }),
  ).toBeVisible()
  await expect(page.locator('.controls')).toBeHidden()
  const stop = page.getByRole('button', {
    name: 'Stop all motion',
    exact: true,
  })
  await expect(stop).toBeInViewport()
  await nav.getByRole('button', { name: 'history', exact: true }).click()
  await expect(
    page.getByRole('button', { name: 'View history', exact: true }),
  ).toBeVisible()
  await expect(stop).toBeInViewport()
})

test('full API fixture renders one commissioned axis and one unavailable axis', async ({
  page,
}) => {
  await page.route('**/api/state', (route) =>
    route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify({
        simulation: false,
        time: new Date().toISOString(),
        events: [],
        axes: [
          {
            id: 'x',
            position: 1.234,
            target: 1.234,
            min: 0,
            max: 5,
            maxSpeed: 0.1,
            state: 'idle',
            error: '',
            available: true,
            updated: new Date().toISOString(),
          },
          {
            id: 'y',
            position: 0,
            target: 0,
            min: 0,
            max: 0,
            maxSpeed: 0,
            state: 'unconfigured',
            error: 'Axis calibration required',
            available: false,
            updated: new Date().toISOString(),
          },
        ],
      }),
    }),
  )
  await page.reload()
  await expect(page.getByText('HARDWARE MODE', { exact: true })).toBeVisible()
  await expect(page.getByTestId('x-position')).toHaveText('1.234')
  await expect(page.getByLabel('Y target metres')).toBeDisabled()
  await page.getByRole('tab', { name: 'Manual jog' }).click()
  await expect(
    page.getByRole('button', { name: 'Jog Y positive' }),
  ).toBeDisabled()
})

test('window blur and Escape release an active jog', async ({
  page,
  request,
}) => {
  await page.getByRole('tab', { name: 'Manual jog' }).click()
  await page.getByRole('button', { name: 'Hold to move', exact: true }).click()
  const button = page.getByRole('button', { name: 'Jog Y negative' })
  await expect(button).toBeEnabled()
  await button.focus()
  await page.keyboard.down('Space')
  await expect
    .poll(
      async () =>
        (await (await request.get('/api/state')).json()).axes[1].state,
    )
    .toBe('jogging')
  await page.evaluate(() => window.dispatchEvent(new Event('blur')))
  await expect
    .poll(
      async () =>
        (await (await request.get('/api/state')).json()).axes[1].state,
    )
    .toBe('idle')
  await page.keyboard.up('Space')
  await expect(button).toBeEnabled()
  await button.focus()
  await page.keyboard.down('Space')
  await expect
    .poll(
      async () =>
        (await (await request.get('/api/state')).json()).axes[1].state,
    )
    .toBe('jogging')
  await page.keyboard.press('Escape')
  await expect
    .poll(
      async () =>
        (await (await request.get('/api/state')).json()).axes[1].state,
    )
    .toBe('idle')
  await page.keyboard.up('Space')
})
