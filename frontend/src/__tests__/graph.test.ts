import { describe, it, expect } from 'vitest'
import { layoutGraph } from '../lib/graph'
import type { WorkflowDefinition } from '../types'

const linearDef: WorkflowDefinition = {
  name: 'test', version: 1, start_state: 's1', end_state: 's2',
  states: [
    { id: 's1', label: 'S1', tasks: [{ key: 't1', label: 'T1', type: 'manual' }] },
    { id: 's2', label: 'S2', tasks: [] },
  ],
  transitions: [{ from_state: 's1', to_state: 's2' }],
}

describe('layoutGraph', () => {
  it('produces correct node count', () => {
    const { nodes } = layoutGraph(linearDef)
    expect(nodes).toHaveLength(2)
  })

  it('produces correct edge count', () => {
    const { edges } = layoutGraph(linearDef)
    expect(edges).toHaveLength(1)
  })

  it('marks start node', () => {
    const { nodes } = layoutGraph(linearDef)
    const start = nodes.find((n) => n.id === 's1')
    expect(start?.isStart).toBe(true)
  })

  it('marks end node as terminal', () => {
    const { nodes } = layoutGraph(linearDef)
    const end = nodes.find((n) => n.id === 's2')
    expect(end?.isTerminal).toBe(true)
  })

  it('nodes have positive positions after layout', () => {
    const { nodes } = layoutGraph(linearDef)
    for (const n of nodes) {
      expect(n.x).toBeGreaterThanOrEqual(0)
      expect(n.y).toBeGreaterThanOrEqual(0)
    }
  })

  it('edge has trigger label when trigger present', () => {
    const defWithTrigger: WorkflowDefinition = {
      name: 'tb', version: 1, start_state: 'a', end_state: 'b',
      states: [
        { id: 'a', label: 'A', tasks: [{ key: 'approve', label: 'Approve', type: 'manual' }] },
        { id: 'b', label: 'B', tasks: [] },
      ],
      transitions: [{ from_state: 'a', to_state: 'b', trigger: { task_key: 'approve' } }],
    }
    const { edges } = layoutGraph(defWithTrigger)
    expect(edges[0].label).toBe('approve')
    expect(edges[0].isTriggered).toBe(true)
  })
})
