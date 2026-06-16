import { test, expect } from '@playwright/test'

test.describe('Admin Reload', () => {
  test('reload-defs shows result', async ({ page }) => {
    await page.goto('/admin/reload')
    await expect(page.locator('h2:has-text("Reload Workflow Definitions")')).toBeVisible()

    await page.locator('button:has-text("Reload Definitions")').click()
    await expect(page.locator('text=Loaded')).toBeVisible({ timeout: 10_000 })
  })
})
