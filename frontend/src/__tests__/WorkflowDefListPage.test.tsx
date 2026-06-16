import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { WorkflowDefListPage } from '../pages/WorkflowDefListPage'
import * as hooks from '../hooks/useWorkflowDefs'

vi.mock('../hooks/useWorkflowDefs')

const mockDef = {
  id: 'abc123',
  name: 'leave-approval',
  version: 1,
  description: 'Test',
  definition: { name: 'leave-approval', version: 1, start_state: 's1', end_state: 's2', states: [{ id: 's1', label: 'S1', tasks: [] }, { id: 's2', label: 'S2', tasks: [] }], transitions: [] },
  created_at: '2026-06-01T00:00:00Z',
}

function wrap(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>
  )
}

describe('WorkflowDefListPage', () => {
  beforeEach(() => { vi.clearAllMocks() })

  it('shows loading skeletons', () => {
    vi.mocked(hooks.useWorkflowDefs).mockReturnValue({ data: undefined, isLoading: true, error: null } as any)
    const { container } = wrap(<WorkflowDefListPage />)
    expect(container.querySelectorAll('.animate-pulse').length).toBeGreaterThan(0)
  })

  it('renders workflow def rows', () => {
    vi.mocked(hooks.useWorkflowDefs).mockReturnValue({
      data: { items: [mockDef], total: 1 },
      isLoading: false,
      error: null,
    } as any)
    wrap(<WorkflowDefListPage />)
    expect(screen.getByText('leave-approval')).toBeInTheDocument()
    expect(screen.getByText('v1')).toBeInTheDocument()
  })

  it('shows empty state when no defs', () => {
    vi.mocked(hooks.useWorkflowDefs).mockReturnValue({
      data: { items: [], total: 0 },
      isLoading: false,
      error: null,
    } as any)
    wrap(<WorkflowDefListPage />)
    expect(screen.getByText('No workflow definitions')).toBeInTheDocument()
  })

  it('shows error message on failure', () => {
    vi.mocked(hooks.useWorkflowDefs).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('network error'),
    } as any)
    wrap(<WorkflowDefListPage />)
    expect(screen.getByText(/Failed to load/)).toBeInTheDocument()
  })
})
