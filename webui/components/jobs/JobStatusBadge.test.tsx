import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import JobStatusBadge from './JobStatusBadge'
import { models_JobStateType } from '@/lib/api/generated'

describe('JobStatusBadge', () => {
  describe('rendering different job states', () => {
    it('renders Undefined state with gray styling', () => {
      render(<JobStatusBadge status={models_JobStateType.JobStateTypeUndefined} />)

      const badge = screen.getByText('Undefined')
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-gray-100')
      expect(badge).toHaveClass('text-gray-800')
    })

    it('renders Pending state with yellow styling', () => {
      render(<JobStatusBadge status={models_JobStateType.JobStateTypePending} />)

      const badge = screen.getByText('Pending')
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-yellow-100')
      expect(badge).toHaveClass('text-yellow-800')
    })

    it('renders Queued state with blue styling', () => {
      render(<JobStatusBadge status={models_JobStateType.JobStateTypeQueued} />)

      const badge = screen.getByText('Queued')
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-blue-100')
      expect(badge).toHaveClass('text-blue-800')
    })

    it('renders Running state with purple styling', () => {
      render(<JobStatusBadge status={models_JobStateType.JobStateTypeRunning} />)

      const badge = screen.getByText('Running')
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-purple-100')
      expect(badge).toHaveClass('text-purple-800')
    })

    it('renders Completed state with green styling', () => {
      render(<JobStatusBadge status={models_JobStateType.JobStateTypeCompleted} />)

      const badge = screen.getByText('Completed')
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-green-100')
      expect(badge).toHaveClass('text-green-800')
    })

    it('renders Failed state with red styling', () => {
      render(<JobStatusBadge status={models_JobStateType.JobStateTypeFailed} />)

      const badge = screen.getByText('Failed')
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-red-100')
      expect(badge).toHaveClass('text-red-800')
    })

    it('renders Stopped state with orange styling', () => {
      render(<JobStatusBadge status={models_JobStateType.JobStateTypeStopped} />)

      const badge = screen.getByText('Stopped')
      expect(badge).toBeInTheDocument()
      expect(badge).toHaveClass('bg-orange-100')
      expect(badge).toHaveClass('text-orange-800')
    })
  })

  describe('handling different input types', () => {
    it('handles string status input', () => {
      render(<JobStatusBadge status="Running" />)

      const badge = screen.getByText('Running')
      expect(badge).toBeInTheDocument()
    })

    it('handles number status input', () => {
      // JobStateTypeRunning = 3
      render(<JobStatusBadge status={3} />)

      const badge = screen.getByText('Running')
      expect(badge).toBeInTheDocument()
    })

    it('handles undefined status gracefully', () => {
      render(<JobStatusBadge status={undefined} />)

      const badge = screen.getByText('Unknown')
      expect(badge).toBeInTheDocument()
    })
  })

  describe('badge styling', () => {
    it('applies text-xs class for small text', () => {
      render(<JobStatusBadge status={models_JobStateType.JobStateTypeRunning} />)

      const badge = screen.getByText('Running')
      expect(badge).toHaveClass('text-xs')
    })
  })
})
