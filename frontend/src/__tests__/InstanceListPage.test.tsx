import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import InstanceListPage from '../pages/InstanceListPage';
import * as processApi from '../api/processes';
import * as processInstanceApi from '../api/processInstances';

vi.mock('../api/processes');
vi.mock('../api/processInstances');

const mockProcesses = [
  { id: 'p1', name: 'Bug Triage', description: '', version: 1, status: 'active', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
];

const mockInstances = [
  { id: 'i1', process_id: 'p1', parent_id: null, status: 'running', started_at: '2026-04-01T10:00:00Z', completed_at: null, created_at: '2026-04-01T10:00:00Z', updated_at: '2026-04-01T10:00:00Z' },
  { id: 'i2', process_id: 'p1', parent_id: null, status: 'completed', started_at: '2026-03-15T08:00:00Z', completed_at: '2026-03-16T12:00:00Z', created_at: '2026-03-15T08:00:00Z', updated_at: '2026-03-16T12:00:00Z' },
];

function renderPage() {
  return render(
    <MemoryRouter>
      <InstanceListPage />
    </MemoryRouter>
  );
}

describe('InstanceListPage', () => {
  beforeEach(() => {
    vi.mocked(processApi.listProcesses).mockResolvedValue(mockProcesses);
    vi.mocked(processInstanceApi.listProcessInstances).mockResolvedValue(mockInstances);
  });

  it('renders the instance table', async () => {
    renderPage();
    await waitFor(() => {
      expect(screen.getAllByText('Bug Triage')).toHaveLength(2);
      expect(screen.getByText('running')).toBeInTheDocument();
      expect(screen.getByText('completed')).toBeInTheDocument();
    });
  });

  it('shows the Start Process button', () => {
    renderPage();
    expect(screen.getByText('Start Process')).toBeInTheDocument();
  });

  it('opens start dialog on button click', async () => {
    const user = userEvent.setup();
    renderPage();
    await user.click(screen.getByText('Start Process'));
    expect(screen.getByText('Start Process', { selector: 'h2' })).toBeInTheDocument();
  });

  it('shows empty state when no instances', async () => {
    vi.mocked(processInstanceApi.listProcessInstances).mockResolvedValue([]);
    renderPage();
    await waitFor(() => {
      expect(screen.getByText(/No instances yet/)).toBeInTheDocument();
    });
  });
});
