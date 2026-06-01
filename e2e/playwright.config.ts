import { defineConfig, devices } from '@playwright/test'

// Targets production by default. Override with TEST_BASE_URL for local stack.
const baseURL = process.env.TEST_BASE_URL ?? 'https://migrator.170.84.141.15.sslip.io'

export default defineConfig({
  testDir: './tests',
  fullyParallel: false, // serial: tests share a Supabase project, avoid races
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: process.env.CI ? [['github'], ['html', { open: 'never' }]] : 'list',
  timeout: 5 * 60 * 1000, // 5 min — migration flow can take ~90s
  expect: { timeout: 10_000 },
  use: {
    baseURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    viewport: { width: 1440, height: 900 },
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
