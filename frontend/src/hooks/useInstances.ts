import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { instancesApi, tasksApi, type InstanceFilter } from '../api/instances'

export function useInstances(filter: InstanceFilter = {}) {
  return useQuery({
    queryKey: ['instances', filter],
    queryFn: () => instancesApi.list(filter),
    refetchInterval: 15_000,
    staleTime: 5_000,
  })
}

export function useInstance(id: string) {
  const query = useQuery({
    queryKey: ['instances', id],
    queryFn: () => instancesApi.getById(id),
    enabled: !!id,
    refetchInterval: (query) => {
      const data = query.state.data
      if (data?.status === 'completed' || data?.status === 'cancelled') return false
      return 5_000
    },
  })
  return query
}

export function useCancelInstance() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => instancesApi.cancel(id),
    onSuccess: (data) => {
      qc.setQueryData(['instances', data.id], data)
      qc.invalidateQueries({ queryKey: ['instances'] })
    },
  })
}

export function useCompleteTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      taskId,
      status,
      nextState,
      completedBy,
    }: {
      taskId: string
      status: 'done' | 'rejected'
      nextState?: string
      completedBy?: string
    }) => tasksApi.complete(taskId, { status, next_state: nextState, completed_by: completedBy }),
    onSuccess: (data) => {
      qc.setQueryData(['instances', data.id], data)
      qc.invalidateQueries({ queryKey: ['instances'] })
    },
  })
}
