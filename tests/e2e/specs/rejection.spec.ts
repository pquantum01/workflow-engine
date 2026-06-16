import { test, expect } from '@playwright/test'
import { reloadDefs, startInstance, getFirstTask, completeTask } from './helpers'

test.describe('Rejection', () => {
  test('reject task → instance cancelled', async ({ request, page }) => {
    await reloadDefs(request)
    const { instId } = await startInstance(request, 'leave-approval')

    const task = await getFirstTask(request, instId)
    await completeTask(request, task.taskId, 'rejected')

    await page.goto(`/workflow-instances/${instId}`)

    // Wait for the page to load and the badge to reflect cancelled status
    // Use a more specific locator scoped to the header area (first badge near instance ID)
    await expect(page.locator('.rounded-full').filter({ hasText: 'cancelled' }).first()).toBeVisible({ timeout: 10_000 })
  })
})
