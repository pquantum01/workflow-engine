import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('Instance API clients', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve([]),
    }));
  });

  it('listProcessInstances calls correct URL', async () => {
    const { listProcessInstances } = await import('../api/processInstances');
    await listProcessInstances();
    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/process-instances?offset=0&limit=20',
      expect.objectContaining({ headers: expect.objectContaining({ 'Content-Type': 'application/json' }) }),
    );
  });

  it('startProcessInstance sends POST with body', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ id: '1', status: 'running' }),
    }));
    const { startProcessInstance } = await import('../api/processInstances');
    await startProcessInstance({ process_id: 'p1' });
    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/process-instances',
      expect.objectContaining({ method: 'POST', body: JSON.stringify({ process_id: 'p1' }) }),
    );
  });

  it('listWorkflowInstances calls correct URL', async () => {
    const { listWorkflowInstances } = await import('../api/workflowInstances');
    await listWorkflowInstances('pi-1');
    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/process-instances/pi-1/workflows',
      expect.objectContaining({ headers: expect.objectContaining({ 'Content-Type': 'application/json' }) }),
    );
  });

  it('listTaskInstances calls correct URL', async () => {
    const { listTaskInstances } = await import('../api/taskInstances');
    await listTaskInstances('wi-1');
    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/workflow-instances/wi-1/tasks',
      expect.objectContaining({ headers: expect.objectContaining({ 'Content-Type': 'application/json' }) }),
    );
  });

  it('listStateTransitions calls correct URL', async () => {
    const { listStateTransitions } = await import('../api/stateTransitions');
    await listStateTransitions('wi-1');
    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/workflow-instances/wi-1/transitions',
      expect.objectContaining({ headers: expect.objectContaining({ 'Content-Type': 'application/json' }) }),
    );
  });
});
