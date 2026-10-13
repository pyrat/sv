import { test, expect } from '@playwright/test'

test.beforeEach(async ({ page, request }) => {
  const state = await (await request.get('/api/state')).json()
  expect(state.simulation).toBe(true)
  await request.post('/api/stop', { data: {} })
  await page.goto('/')
  await expect(
    page.getByRole('button', { name: /Move to position/ }),
  ).toBeEnabled()
})

test('history includes older events beyond the live preview and pages without duplicates', async ({
  page,
  request,
}) => {
  expect(
    (await request.post('/api/home/clear', { data: {} })).ok(),
  ).toBeTruthy()
  for (let i = 0; i < 65; i++) {
    expect((await request.post('/api/stop', { data: {} })).ok()).toBeTruthy()
  }
  await expect
    .poll(
      async () =>
        (await (await request.get('/api/state')).json()).activity.queued,
    )
    .toBe(0)
  const state = await (await request.get('/api/state')).json()
  expect(state.events.length).toBe(12)
  expect(state.activity.retained).toBeGreaterThan(65)
  await page.getByRole('button', { name: 'View history', exact: true }).click()
  const dialog = page.getByRole('dialog', { name: 'Activity history' })
  await expect(dialog.locator('.history-event')).toHaveCount(50)
  await expect(
    dialog.getByText('Home position cleared', { exact: true }),
  ).toHaveCount(0)
  await dialog.getByRole('button', { name: 'Load older activity' }).click()
  await expect(
    dialog.getByText('Home position cleared', { exact: true }),
  ).toBeVisible()
  const ids = (
    await dialog.locator('.history-event-meta span').allTextContents()
  ).filter((text) => text.startsWith('#'))
  expect(new Set(ids).size).toBe(ids.length)
  expect(ids.length).toBeGreaterThan(50)
  await dialog.getByRole('button', { name: 'Refresh history' }).click()
  await expect(dialog.locator('.history-event')).toHaveCount(50)
  await page.screenshot({ path: 'test-results/activity-history-desktop.png' })
})

test('history reports API errors and can retry', async ({ page }) => {
  await page.route('**/api/activity?*', (route) =>
    route.fulfill({
      status: 503,
      contentType: 'application/json',
      body: '{"error":"Archive temporarily offline"}',
    }),
  )
  await page.getByRole('button', { name: 'View history', exact: true }).click()
  const dialog = page.getByRole('dialog', { name: 'Activity history' })
  await expect(dialog.getByRole('alert')).toContainText(
    'Archive temporarily offline',
  )
  await page.unroute('**/api/activity?*')
  await dialog.getByRole('button', { name: 'Refresh history' }).click()
  await expect(dialog.getByRole('alert')).toHaveCount(0)
  await expect(dialog.locator('.history-event').first()).toBeVisible()
})

test('storage errors are visible instead of claiming persistence', async ({
  page,
  request,
}) => {
  const state = await (await request.get('/api/state')).json()
  state.activity = {
    ...state.activity,
    persistent: true,
    queued: 3,
    error: 'disk full',
  }
  await page.route('**/api/state', (route) =>
    route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify(state),
    }),
  )
  await expect(page.getByText('STORAGE ISSUE', { exact: true })).toBeVisible()
  await expect(page.getByRole('status')).toContainText('disk full')
  await expect(page.getByText('PERSISTED', { exact: true })).toHaveCount(0)
})

test('history fits mobile and closes with Escape', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page
    .getByRole('navigation', { name: 'Phone views' })
    .getByRole('button', { name: 'history' })
    .click()
  await page.getByRole('button', { name: 'View history', exact: true }).click()
  const dialog = page.getByRole('dialog', { name: 'Activity history' })
  await expect(dialog.locator('.history-event').first()).toBeVisible()
  const bounds = (await dialog.boundingBox())!
  expect(bounds.x).toBeGreaterThanOrEqual(0)
  expect(bounds.x + bounds.width).toBeLessThanOrEqual(390)
  await page.screenshot({ path: 'test-results/activity-history-mobile.png' })
  await page.keyboard.press('Escape')
  await expect(dialog).toHaveCount(0)
})
