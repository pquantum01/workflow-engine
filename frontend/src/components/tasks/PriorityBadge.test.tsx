import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { PriorityBadge } from './PriorityBadge'

describe('PriorityBadge', () => {
  it('renders high priority in red', () => {
    render(<PriorityBadge priority="high" />)
    const el = screen.getByText('High')
    expect(el).toBeTruthy()
    expect(el.className).toContain('red')
  })

  it('renders medium priority in yellow', () => {
    render(<PriorityBadge priority="medium" />)
    expect(screen.getByText('Medium').className).toContain('yellow')
  })

  it('renders low priority in zinc', () => {
    render(<PriorityBadge priority="low" />)
    expect(screen.getByText('Low').className).toContain('zinc')
  })
})
