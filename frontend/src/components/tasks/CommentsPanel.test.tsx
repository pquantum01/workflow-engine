import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import { CommentsPanel } from './CommentsPanel'

vi.mock('../../api/comments', () => ({
  commentsApi: {
    list: vi.fn().mockResolvedValue([
      { id: '1', task_instance_id: 'tid', author: 'alice', content: 'first comment', created_at: '2026-07-03T00:00:00Z' },
    ]),
    add: vi.fn().mockResolvedValue(
      { id: '2', task_instance_id: 'tid', author: 'bob', content: 'new comment', created_at: '2026-07-03T01:00:00Z' }
    ),
  },
}))

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('CommentsPanel', () => {
  it('shows existing comment', async () => {
    wrap(<CommentsPanel taskId="tid" />)
    await waitFor(() => screen.getByText('first comment'))
    expect(screen.getByText('alice')).toBeTruthy()
  })

  it('submit button disabled when fields empty', async () => {
    wrap(<CommentsPanel taskId="tid" />)
    await waitFor(() => screen.getByText('first comment'))
    const btn = screen.getByRole('button', { name: /add comment/i })
    expect(btn).toBeDisabled()
  })
})
