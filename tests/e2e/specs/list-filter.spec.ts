import { test, expect } from '@playwright/test'
import { reloadDefs, startInstance } from './helpers'

test.describe('Instance List Filters', () => {
  test('filter by status shows correct subset', async ({ request, page }) => {
    await reloadDefs(request)
    await startInstance(request, 'leave-approval')
    await startInstance(request, 'leave-approval')

    await page.goto('/workflow-instances')
    await expect(page.locator('table')).toBeVisible({ timeout: 10_000 })

    // Select active filter
    await page.locator('select').first().selectOption('active')
    await page.waitForTimeout(500)

    // Should show rows
    const rows = page.locator('tbody tr')
    await expect(rows).not.toHaveCount(0)
  })
})
