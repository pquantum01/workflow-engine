import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { useWorkflowDefs } from '../hooks/useWorkflowDefs'
import { instancesApi } from '../api/instances'
import { ApiError } from '../api/client'
import { getCompletedBy, setCompletedBy } from '../lib/session'

export function StartInstancePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const preselectedDefId = searchParams.get('def_id') ?? ''

  const { data: defsData } = useWorkflowDefs()
  const defs = defsData?.items ?? []

  const [defId, setDefId] = useState(preselectedDefId)
  const [createdBy, setCreatedByState] = useState(getCompletedBy())
  const [contextRaw, setContextRaw] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!defId) { setError('Select a workflow definition.'); return }
    if (!createdBy.trim()) { setError('Created by is required.'); return }

    let initialContext: Record<string, unknown> | undefined
    if (contextRaw.trim()) {
      try {
        initialContext = JSON.parse(contextRaw)
      } catch {
        setError('Initial context must be valid JSON.')
        return
      }
    }

    setLoading(true)
    try {
      setCompletedBy(createdBy)
      const inst = await instancesApi.start({ workflow_def_id: defId, created_by: createdBy, initial_context: initialContext })
      navigate(`/workflow-instances/${inst.id}`)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to start instance.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-6 max-w-lg">
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => navigate(-1)} className="p-1.5 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-500">
          <ArrowLeft size={16} />
        </button>
        <h1 className="text-2xl font-semibold tracking-tight">Start Instance</h1>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            Workflow Definition <span className="text-red-500">*</span>
          </label>
          <select
            value={defId}
            onChange={(e) => setDefId(e.target.value)}
            className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg px-3 py-2 bg-white dark:bg-zinc-800 focus:ring-2 focus:ring-indigo-500 outline-none"
          >
            <option value="">Select a workflow…</option>
            {defs.map((d) => (
              <option key={d.id} value={d.id}>{d.name} v{d.version}</option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            Created By <span className="text-red-500">*</span>
          </label>
          <input
            value={createdBy}
            onChange={(e) => setCreatedByState(e.target.value)}
            placeholder="your@email.com"
            className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg px-3 py-2 bg-white dark:bg-zinc-800 focus:ring-2 focus:ring-indigo-500 outline-none"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            Instance Context (optional JSON)
          </label>
          <textarea
            value={contextRaw}
            onChange={(e) => setContextRaw(e.target.value)}
            placeholder='{"key": "value"}'
            rows={3}
            className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg px-3 py-2 bg-white dark:bg-zinc-800 focus:ring-2 focus:ring-indigo-500 outline-none font-mono"
          />
        </div>

        {error && <p className="text-sm text-red-600">{error}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full px-4 py-2 text-sm font-medium rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 transition-colors"
        >
          {loading ? 'Starting…' : 'Start Instance'}
        </button>
      </form>
    </div>
  )
}
