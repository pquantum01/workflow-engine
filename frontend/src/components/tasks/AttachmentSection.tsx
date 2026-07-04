import { useRef } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Paperclip, Download, Trash2, Upload } from 'lucide-react'
import { attachmentApi } from '../../api/attachment'

const MAX_BYTES = 2 * 1024 * 1024

interface Props {
  taskId: string
  hasMeta: boolean
  filename?: string
}

export function AttachmentSection({ taskId, hasMeta, filename }: Props) {
  const qc = useQueryClient()
  const inputRef = useRef<HTMLInputElement>(null)

  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      if (file.size > MAX_BYTES) throw new Error('FILE_TOO_LARGE')
      const buf = await file.arrayBuffer()
      const b64 = btoa(String.fromCharCode(...new Uint8Array(buf)))
      return attachmentApi.upload(taskId, {
        filename: file.name,
        content_type: file.type || 'application/octet-stream',
        data: b64,
      })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['task', taskId] }),
  })

  const deleteMutation = useMutation({
    mutationFn: () => attachmentApi.delete(taskId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['task', taskId] }),
  })

  const handleDownload = async () => {
    const blob = await attachmentApi.download(taskId)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename ?? 'attachment'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-3">
      <h3 className="text-xs font-medium text-zinc-500 uppercase tracking-wide">Attachment</h3>

      {hasMeta && filename && (
        <div className="flex items-center gap-2 p-3 rounded-lg border border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-800/50">
          <Paperclip size={14} className="text-zinc-400 shrink-0" />
          <span className="text-sm text-zinc-700 dark:text-zinc-300 flex-1 truncate">{filename}</span>
          <button
            onClick={handleDownload}
            className="flex items-center gap-1 px-2 py-1 rounded text-xs font-medium text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 transition-colors"
          >
            <Download size={12} />
            Download
          </button>
          <button
            onClick={() => deleteMutation.mutate()}
            disabled={deleteMutation.isPending}
            className="flex items-center gap-1 px-2 py-1 rounded text-xs font-medium text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 transition-colors"
          >
            <Trash2 size={12} />
            {deleteMutation.isPending ? 'Removing…' : 'Remove'}
          </button>
        </div>
      )}

      <div>
        <input
          ref={inputRef}
          type="file"
          className="hidden"
          onChange={(e) => {
            const file = e.target.files?.[0]
            if (file) uploadMutation.mutate(file)
            if (inputRef.current) inputRef.current.value = ''
          }}
        />
        <button
          onClick={() => inputRef.current?.click()}
          disabled={uploadMutation.isPending}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border border-zinc-300 dark:border-zinc-600 text-zinc-600 dark:text-zinc-400 hover:bg-zinc-50 dark:hover:bg-zinc-800 disabled:opacity-50 transition-colors"
        >
          <Upload size={12} />
          {uploadMutation.isPending ? 'Uploading…' : hasMeta ? 'Replace File' : 'Upload File'}
        </button>
        {uploadMutation.isError && (
          <p className="text-xs text-red-500 mt-1.5">
            {(uploadMutation.error as Error).message === 'FILE_TOO_LARGE'
              ? 'File must be under 2 MB.'
              : 'Upload failed.'}
          </p>
        )}
      </div>
    </div>
  )
}
