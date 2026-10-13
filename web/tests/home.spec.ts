import { test, expect } from '@playwright/test'

test.beforeEach(async ({ page, request }) => {
  await request.post('/api/stop', { data: {} })
  await page.goto('/')
  await expect(
    page.getByRole('button', { name: /Move to position/ }),
  ).toBeEnabled()
  expect(
    (await request.post('/api/home/clear', { data: {} })).ok(),
  ).toBeTruthy()
  await expect(page.getByTestId('home-status')).toHaveText('NOT SET')
})

test('capture home, go out X then Y, and return Y then X', async ({
  page,
  request,
}) => {
  const state = async () => await (await request.get('/api/state')).json()
  const initial = await state()
  const x = initial.axes[0].position,
    y = initial.axes[1].position
  await page.getByRole('button', { name: 'Set current as home' }).click()
  await expect(page.getByTestId('home-status')).toHaveText('AT HOME')
  const saved = (await state()).home.saved
  expect(saved.axes.x.feedbackCounter).toBe(initial.axes[0].feedbackCounter)
  expect((await state()).axes[0].profileCounter).toBe(
    initial.axes[0].profileCounter,
  )
  const targetX = (x + 0.15).toFixed(3),
    targetY = (y + 0.15).toFixed(3)
  await page.getByLabel('X target metres').fill(targetX)
  await page.getByLabel('Y target metres').fill(targetY)
  await page.getByRole('button', { name: /Move to position/ }).click()
  await expect.poll(async () => (await state()).motion.activeAxis).toBe('x')
  expect((await state()).axes[1].position).toBe(y)
  expect((await state()).axes[1].state).toBe('queued')
  await expect.poll(async () => (await state()).motion.activeAxis).toBe('y')
  expect((await state()).axes[0].position).toBe(Number(targetX))
  await expect.poll(async () => (await state()).motion.status).toBe('completed')
  await expect(page.getByRole('button', { name: /Return home/ })).toBeEnabled()
  await page.getByRole('button', { name: /Return home/ }).click()
  await expect.poll(async () => (await state()).motion.activeAxis).toBe('y')
  expect((await state()).axes[0].position).toBe(Number(targetX))
  await expect(page.getByTestId('motion-status')).toContainText(
    'RETURNING HOME',
  )
  await expect.poll(async () => (await state()).motion.activeAxis).toBe('x')
  expect((await state()).axes[1].position).toBe(y)
  await expect(page.getByTestId('home-status')).toHaveText('AT HOME')
  await expect(page.getByTestId('x-position')).toHaveText(x.toFixed(3))
  await expect(page.getByTestId('y-position')).toHaveText(y.toFixed(3))
  await page.screenshot({
    path: 'test-results/home-desktop.png',
    fullPage: true,
  })
})

test('home editor accepts command counters without changing live position', async ({
  page,
  request,
}) => {
  const before = await (await request.get('/api/state')).json()
  await page.getByRole('button', { name: 'Edit home position' }).click()
  await page
    .getByRole('button', { name: 'Command counters', exact: true })
    .click()
  await page.getByLabel('X home command counter').fill('106000')
  await page.getByLabel('Y home command counter').fill('212000')
  await page.getByRole('button', { name: 'Save home', exact: true }).click()
  await expect(page.getByTestId('home-x')).toContainText('2.000')
  await expect(page.getByTestId('home-y')).toContainText('4.000')
  const after = await (await request.get('/api/state')).json()
  expect(after.axes[0].position).toBe(before.axes[0].position)
  expect(after.axes[1].feedbackCounter).toBe(before.axes[1].feedbackCounter)
  await page.reload()
  await expect(page.getByTestId('home-x')).toContainText('2.000')
  await page.getByRole('button', { name: 'Clear home', exact: true }).click()
  await expect(page.getByTestId('home-status')).toHaveText('NOT SET')
  await expect(page.getByRole('button', { name: /Return home/ })).toBeDisabled()
})

test('stopping an outbound move discards the queued Y stage', async ({
  page,
  request,
}) => {
  const initial = await (await request.get('/api/state')).json()
  await page
    .getByLabel('X target metres')
    .fill((initial.axes[0].position + 0.5).toFixed(3))
  await page
    .getByLabel('Y target metres')
    .fill((initial.axes[1].position + 0.5).toFixed(3))
  await page.getByRole('button', { name: /Move to position/ }).click()
  await expect(page.getByTestId('motion-status')).toContainText('OUTBOUND MOVE')
  await page.getByRole('button', { name: /Stop all motion/ }).click()
  await expect(page.getByTestId('motion-status')).toContainText(
    'SEQUENCE STOPPED',
  )
  await expect
    .poll(
      async () =>
        (await (await request.get('/api/state')).json()).axes[0].state,
    )
    .toBe('idle')
  expect(
    (await (await request.get('/api/state')).json()).axes[1].position,
  ).toBe(initial.axes[1].position)
})
