import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { commentsApi } from '../../api/comments'
import type { TaskComment } from '../../types'

interface Props {
  taskId: string
}

export function CommentsPanel({ taskId }: Props) {
  const qc = useQueryClient()
  const [author, setAuthor] = useState('')
  const [content, setContent] = useState('')

  const { data: comments = [] } = useQuery<TaskComment[]>({
    queryKey: ['task-comments', taskId],
    queryFn: () => commentsApi.list(taskId),
    staleTime: 30_000,
  })

  const addMutation = useMutation({
    mutationFn: () => commentsApi.add(taskId, { author, content }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task-comments', taskId] })
      setAuthor('')
      setContent('')
    },
  })

  const canSubmit = author.trim().length > 0 && content.trim().length > 0

  return (
    <div className="space-y-3">
      <h3 className="text-xs font-medium text-zinc-500 uppercase tracking-wide">Comments</h3>

      <div className="space-y-2">
        {comments.map((c) => (
          <div key={c.id} className="rounded-lg border border-zinc-200 dark:border-zinc-700 p-3 bg-zinc-50 dark:bg-zinc-800/50">
            <div className="flex items-center justify-between mb-1">
              <span className="text-xs font-medium text-zinc-700 dark:text-zinc-300">{c.author}</span>
              <span className="text-xs text-zinc-400">{new Date(c.created_at).toLocaleString()}</span>
            </div>
            <p className="text-sm text-zinc-600 dark:text-zinc-400 whitespace-pre-wrap leading-relaxed">{c.content}</p>
          </div>
        ))}
        {comments.length === 0 && (
          <p className="text-xs text-zinc-400 py-1">No comments yet.</p>
        )}
      </div>

      <div className="space-y-2 border-t border-zinc-100 dark:border-zinc-800 pt-3">
        <input
          value={author}
          onChange={(e) => setAuthor(e.target.value)}
          placeholder="Your name"
          className="w-full rounded-lg border border-zinc-300 dark:border-zinc-600 px-3 py-1.5 text-sm bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 placeholder-zinc-400 focus:ring-2 focus:ring-indigo-500 outline-none"
        />
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Add a comment…"
          rows={3}
          maxLength={2048}
          className="w-full rounded-lg border border-zinc-300 dark:border-zinc-600 px-3 py-1.5 text-sm bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 placeholder-zinc-400 focus:ring-2 focus:ring-indigo-500 outline-none resize-none"
        />
        <div className="flex items-center gap-2">
          <button
            disabled={!canSubmit || addMutation.isPending}
            onClick={() => addMutation.mutate()}
            className="rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-indigo-700 disabled:opacity-50 transition-colors"
          >
            {addMutation.isPending ? 'Adding…' : 'Add Comment'}
          </button>
          {addMutation.isError && (
            <p className="text-xs text-red-500">Failed to add comment.</p>
          )}
        </div>
      </div>
    </div>
  )
}
