import { test, expect } from '@playwright/test'
import { reloadDefs, startInstance } from './helpers'

test.describe('Cancel Instance', () => {
  test('cancel instance → status cancelled', async ({ request, page }) => {
    await reloadDefs(request)
    const { instId } = await startInstance(request, 'leave-approval')

    await page.goto(`/workflow-instances/${instId}`)
    await expect(page.locator('text=active')).toBeVisible({ timeout: 10_000 })

    // Click Cancel Instance
    await page.locator('button:has-text("Cancel Instance")').click()
    await page.locator('button:has-text("Confirm")').click()

    await expect(page.locator('text=cancelled')).toBeVisible({ timeout: 10_000 })
  })
})
