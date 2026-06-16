import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './specs',
  timeout: 30_000,
  use: {
    baseURL: process.env.BASE_URL ?? 'http://localhost:8080',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  retries: 1,
})
