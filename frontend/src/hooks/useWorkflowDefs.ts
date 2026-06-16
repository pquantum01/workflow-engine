import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { workflowDefsApi, adminApi } from '../api/workflowDefs'

export function useWorkflowDefs(includeDeleted = false) {
  return useQuery({
    queryKey: ['workflow-defs', { includeDeleted }],
    queryFn: () => workflowDefsApi.list(includeDeleted),
    staleTime: 30_000,
  })
}

export function useWorkflowDef(id: string) {
  return useQuery({
    queryKey: ['workflow-defs', id],
    queryFn: () => workflowDefsApi.getById(id),
    enabled: !!id,
    staleTime: 30_000,
  })
}

export function useReloadDefs() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: adminApi.reloadDefs,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workflow-defs'] })
    },
  })
}
