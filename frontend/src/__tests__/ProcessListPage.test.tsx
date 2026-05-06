import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ProcessListPage from '../pages/ProcessListPage';
import * as processApi from '../api/processes';

vi.mock('../api/processes');

const mockProcesses = [
  { id: '1', name: 'Bug Triage', description: 'desc', version: 1, status: 'draft', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
  { id: '2', name: 'Onboarding', description: '', version: 1, status: 'active', created_at: '2026-01-02T00:00:00Z', updated_at: '2026-01-02T00:00:00Z' },
];

function renderPage() {
  return render(
    <MemoryRouter>
      <ProcessListPage />
    </MemoryRouter>
  );
}

describe('ProcessListPage', () => {
  beforeEach(() => {
    vi.mocked(processApi.listProcesses).mockResolvedValue(mockProcesses);
    vi.mocked(processApi.createProcess).mockResolvedValue(mockProcesses[0]);
    vi.mocked(processApi.deleteProcess).mockResolvedValue(undefined);
  });

  it('renders the process table', async () => {
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('Bug Triage')).toBeInTheDocument();
      expect(screen.getByText('Onboarding')).toBeInTheDocument();
    });
  });

  it('shows the New Process button', async () => {
    renderPage();
    expect(screen.getByText('New Process')).toBeInTheDocument();
  });

  it('opens create dialog on button click', async () => {
    const user = userEvent.setup();
    renderPage();
    await user.click(screen.getByText('New Process'));
    expect(screen.getByText('New Process', { selector: 'h2' })).toBeInTheDocument();
  });

  it('shows empty state when no processes', async () => {
    vi.mocked(processApi.listProcesses).mockResolvedValue([]);
    renderPage();
    await waitFor(() => {
      expect(screen.getByText(/No processes yet/)).toBeInTheDocument();
    });
  });
});
