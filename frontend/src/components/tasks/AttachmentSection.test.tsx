import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import { AttachmentSection } from './AttachmentSection'

vi.mock('../../api/attachment', () => ({
  attachmentApi: {
    upload: vi.fn(),
    download: vi.fn(),
    delete: vi.fn(),
  },
}))

function wrap(ui: React.ReactNode) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('AttachmentSection', () => {
  it('shows upload button when no attachment', () => {
    wrap(<AttachmentSection taskId="tid" hasMeta={false} />)
    expect(screen.getByRole('button', { name: /upload/i })).toBeTruthy()
  })

  it('shows download button when attachment exists', () => {
    wrap(<AttachmentSection taskId="tid" hasMeta={true} filename="doc.pdf" />)
    expect(screen.getByRole('button', { name: /download/i })).toBeTruthy()
  })
})
