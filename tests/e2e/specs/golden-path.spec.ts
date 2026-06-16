import { test, expect } from '@playwright/test'
import { reloadDefs, startInstance, getFirstTask, completeTask } from './helpers'

test.describe('Golden Path', () => {
  test('start instance → complete all tasks → instance completed', async ({ request, page }) => {
    await reloadDefs(request)
    const { instId } = await startInstance(request, 'leave-approval')

    // Navigate to instance detail
    await page.goto(`/workflow-instances/${instId}`)
    await expect(page.locator('text=active')).toBeVisible({ timeout: 10_000 })

    // Complete submitted task
    const task1 = await getFirstTask(request, instId)
    await completeTask(request, task1.taskId)

    // Reload page, expect state moved to hr_review
    await page.reload()
    await expect(page.locator('text=hr_review')).toBeVisible({ timeout: 10_000 })

    // Complete hr_review task
    const task2 = await getFirstTask(request, instId)
    await completeTask(request, task2.taskId)

    // Expect completed
    await page.reload()
    await expect(page.locator('text=completed')).toBeVisible({ timeout: 10_000 })
  })
})
