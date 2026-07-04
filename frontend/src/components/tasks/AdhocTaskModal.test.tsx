import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import { AdhocTaskModal } from './AdhocTaskModal'

vi.mock('../../api/tasks', () => ({
  tasksExtApi: { createAdhoc: vi.fn().mockResolvedValue({}) },
}))

const states = [{ id: 'review', label: 'Review', tasks: [] }]

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('AdhocTaskModal', () => {
  it('renders state dropdown', () => {
    wrap(<AdhocTaskModal instanceId="iid" states={states} onClose={() => {}} />)
    const selects = screen.getAllByRole('combobox')
    expect(selects.length).toBeGreaterThan(0)
    expect(screen.getByText('Review')).toBeTruthy()
  })

  it('submit button disabled when label empty', () => {
    wrap(<AdhocTaskModal instanceId="iid" states={states} onClose={() => {}} />)
    const btn = screen.getByRole('button', { name: /add task/i })
    expect(btn).toBeDisabled()
  })
})
