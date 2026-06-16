import type { APIRequestContext } from '@playwright/test'

const API = (process.env.BASE_URL ?? 'http://localhost:8080') + '/api/v1'

export async function reloadDefs(request: APIRequestContext) {
  await request.post(`${API}/admin/reload-defs`)
}

export async function startInstance(request: APIRequestContext, defName: string): Promise<{ instId: string; defId: string }> {
  // get def id
  const defsResp = await request.get(`${API}/workflow-defs`)
  const defs = await defsResp.json()
  const def = defs.items.find((d: any) => d.name === defName)
  if (!def) throw new Error(`Def "${defName}" not found`)

  const instResp = await request.post(`${API}/workflow-instances`, {
    data: { workflow_def_id: def.id, created_by: 'e2e-test' },
  })
  const inst = await instResp.json()
  return { instId: inst.id, defId: def.id }
}

export async function getFirstTask(request: APIRequestContext, instId: string): Promise<{ taskId: string; taskKey: string }> {
  const resp = await request.get(`${API}/workflow-instances/${instId}/tasks`)
  const tasks = await resp.json()
  const task = tasks.find((t: any) => t.status === 'new' || t.status === 'in_progress')
  if (!task) throw new Error('No open task found')
  return { taskId: task.id, taskKey: task.task_key }
}

export async function completeTask(request: APIRequestContext, taskId: string, status = 'done', nextState?: string) {
  const body: Record<string, string> = { status, completed_by: 'e2e-test' }
  if (nextState) body.next_state = nextState
  return request.post(`${API}/task-instances/${taskId}/complete`, { data: body })
}
