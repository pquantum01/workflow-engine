import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Layout } from './components/layout/Layout'
import { WorkflowDefListPage } from './pages/WorkflowDefListPage'
import { WorkflowDefDetailPage } from './pages/WorkflowDefDetailPage'
import { InstanceListPage } from './pages/InstanceListPage'
import { StartInstancePage } from './pages/StartInstancePage'
import { InstanceDetailPage } from './pages/InstanceDetailPage'
import { TaskDetailPage } from './pages/TaskDetailPage'
import { AdminReloadPage } from './pages/AdminReloadPage'
import { NotFoundPage } from './pages/NotFoundPage'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: 1 },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Navigate to="/workflow-defs" replace />} />
            <Route path="workflow-defs" element={<WorkflowDefListPage />} />
            <Route path="workflow-defs/:id" element={<WorkflowDefDetailPage />} />
            <Route path="workflow-instances" element={<InstanceListPage />} />
            <Route path="workflow-instances/new" element={<StartInstancePage />} />
            <Route path="workflow-instances/:id" element={<InstanceDetailPage />} />
            <Route path="task-instances/:id" element={<TaskDetailPage />} />
            <Route path="admin/reload" element={<AdminReloadPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
