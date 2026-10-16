import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  workers: 1,
  use: {
    baseURL: 'http://127.0.0.1:18080',
    viewport: { width: 1440, height: 1100 },
    trace: 'retain-on-failure',
  },
  webServer: {
    command:
      'go run ./cmd/depthpilot-web -listen 127.0.0.1:18080 -simulate -state - -web ../web/dist',
    cwd: '../controller',
    url: 'http://127.0.0.1:18080/api/state',
    reuseExistingServer: process.env.PW_REUSE_SERVER === '1',
    timeout: 120000,
  },
})
