// Run against a separately started depthpilot-rope simulation only.
import { chromium } from '@playwright/test';
import assert from 'node:assert/strict';

const base = process.env.ROPE_TEST_URL || 'http://127.0.0.1:8088';
const browser = await chromium.launch({ headless: true });
try {
  const page = await browser.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  const initial = await (await page.request.get(`${base}/api/state`)).json();
  assert.equal(initial.mode, 'simulation', 'Refusing to run motion smoke tests on hardware');
  await page.goto(base);
  await page.waitForFunction(() => document.querySelector('#connection').textContent === 'Connected');
  await page.getByRole('button', { name: 'Capture current positions as virtual origin' }).click();
  await page.waitForFunction(() => document.querySelector('#phase').textContent === 'idle');
  for (const [name, target] of [
    ['Right', [0.1, 0]], ['Left', [0, 0]], ['Lower', [0, 0.1]],
    ['Raise', [0, 0]], ['Right + raise', [0.1, -0.1]],
  ]) {
    await page.getByRole('button', { name, exact: true }).click();
    await page.waitForFunction(() => ['preparing', 'moving'].includes(document.querySelector('#phase').textContent));
    await page.waitForFunction(() => document.querySelector('#phase').textContent === 'idle', null, { timeout: 20000 });
    const s = await (await page.request.get(`${base}/api/state`)).json();
    assert.equal(s.error, '');
    assert.equal(s.lastMove.outcome, 'completed');
    const both = name === 'Right' || name === 'Left';
    assert.equal(s.lastMove.startMode, initial.startMode === 'broadcast' && both ? 'broadcast' : 'individual');
    assert.ok(s.lastMove.points.length > 5, 'Missing aligned path samples');
    assert.ok(s.lastMove.finalErrorM < 0.0001, 'Unexpected endpoint error');
    assert.ok(s.lastMove.maxCrossTrackM < 0.001, 'Unexpected simulated path deviation');
    const csv = await page.request.get(`${base}/api/trace.csv`);
    assert.equal(csv.status(), 200);
    assert.ok((await csv.text()).includes('interpolation_bracket_ms'));
    assert.ok(Math.abs(s.position.x - target[0]) < 0.00001);
    assert.ok(Math.abs(s.position.z - target[1]) < 0.00001);
    assert.equal(s.drums[0].enabled, false, 'Automatic brake enable must be restored');
    if (name === 'Right + raise') assert.equal(s.plan.profiles[1].delta, 0);
    console.log(`${name}: passed`);
  }
  await page.setViewportSize({ width: 390, height: 844 });
  await page.waitForFunction(() => document.querySelector('#tracePlot path') !== null);
  assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
  await page.getByRole('button', { name: 'STOP BOTH DRIVES' }).click();
  await page.waitForFunction(() => document.querySelector('#phase').textContent === 'unreferenced', null, { timeout: 10000 });
  assert.equal(await page.getByRole('button', { name: 'Move both drums to target' }).isDisabled(), true);
  assert.deepEqual(errors, []);
  console.log('Mobile layout and paired stop: passed');
} finally {
  await browser.close();
}
