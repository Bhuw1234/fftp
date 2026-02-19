import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { JobsTable } from './JobsTable'
import { models_Job, models_JobStateType } from '@/lib/api/generated'

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}))

// Mock child components
vi.mock('@/components/TruncatedTextWithTooltip', () => ({
  default: ({ text }: { text?: string }) => <span>{text || 'N/A'}</span>,
}))

vi.mock('@/components/jobs/JobStatusBadge', () => ({
  default: ({ status }: { status: number }) => (
    <span data-testid="status-badge">{status}</span>
  ),
}))

vi.mock('@/components/jobs/JobEngine', () => ({
  default: () => <span data-testid="job-engine">Docker</span>,
}))

vi.mock('@/lib/time', () => ({
  formatTimestamp: (timestamp: number) => `formatted-${timestamp}`,
}))

vi.mock('@/lib/api/utils', () => ({
  getJobRunTime: () => '5m 30s',
}))

// Create mock job data
const createMockJob = (overrides: Partial<models_Job> = {}): models_Job => ({
  ID: 'j-test-job-id-123',
  Name: 'test-job',
  CreateTime: 1700000000000,
  ModifyTime: 1700000100000,
  State: {
    StateType: models_JobStateType.JobStateTypeRunning,
    Message: 'Job is running smoothly',
  },
  Type: 'batch',
  ...overrides,
})

describe('JobsTable', () => {
  const defaultProps = {
    jobs: [],
    pageSize: 10,
    setPageSize: vi.fn(),
    pageIndex: 0,
    onPreviousPage: vi.fn(),
    onNextPage: vi.fn(),
    hasNextPage: false,
  }

  it('renders empty table when no jobs', () => {
    render(<JobsTable {...defaultProps} />)

    expect(screen.getByText('ID')).toBeInTheDocument()
    expect(screen.getByText('Status')).toBeInTheDocument()
    expect(screen.getByText('Created At')).toBeInTheDocument()
  })

  it('renders jobs with correct data', () => {
    const jobs = [
      createMockJob({
        ID: 'j-abc123',
        Name: 'my-test-job',
        Type: 'batch',
      }),
    ]

    render(<JobsTable {...defaultProps} jobs={jobs} />)

    expect(screen.getByText('my-test-job')).toBeInTheDocument()
    expect(screen.getByText('batch')).toBeInTheDocument()
  })

  it('renders multiple jobs', () => {
    const jobs = [
      createMockJob({ ID: 'j-1', Name: 'job-1' }),
      createMockJob({ ID: 'j-2', Name: 'job-2' }),
      createMockJob({ ID: 'j-3', Name: 'job-3' }),
    ]

    render(<JobsTable {...defaultProps} jobs={jobs} />)

    expect(screen.getByText('job-1')).toBeInTheDocument()
    expect(screen.getByText('job-2')).toBeInTheDocument()
    expect(screen.getByText('job-3')).toBeInTheDocument()
  })

  it('renders job links correctly', () => {
    const jobs = [createMockJob({ ID: 'j-abc123', Name: 'test-job' })]

    render(<JobsTable {...defaultProps} jobs={jobs} />)

    const link = screen.getByRole('link', { name: 'test-job' })
    expect(link).toHaveAttribute('href', '/jobs?id=j-abc123')
  })

  it('calls setPageSize when page size select changes', async () => {
    const user = userEvent.setup()
    const setPageSize = vi.fn()

    render(<JobsTable {...defaultProps} setPageSize={setPageSize} />)

    // Open the select
    const selectTrigger = screen.getByRole('combobox')
    await user.click(selectTrigger)

    // Click on 20
    const option20 = screen.getByRole('option', { name: '20' })
    await user.click(option20)

    expect(setPageSize).toHaveBeenCalledWith(20)
  })

  it('disables Previous button on first page', () => {
    render(<JobsTable {...defaultProps} pageIndex={0} />)

    const prevButton = screen.getByRole('button', { name: /previous/i })
    expect(prevButton).toBeDisabled()
  })

  it('enables Previous button when not on first page', () => {
    render(<JobsTable {...defaultProps} pageIndex={1} />)

    const prevButton = screen.getByRole('button', { name: /previous/i })
    expect(prevButton).not.toBeDisabled()
  })

  it('disables Next button when no next page', () => {
    render(<JobsTable {...defaultProps} hasNextPage={false} />)

    const nextButton = screen.getByRole('button', { name: /next/i })
    expect(nextButton).toBeDisabled()
  })

  it('enables Next button when hasNextPage is true', () => {
    render(<JobsTable {...defaultProps} hasNextPage={true} />)

    const nextButton = screen.getByRole('button', { name: /next/i })
    expect(nextButton).not.toBeDisabled()
  })

  it('calls onPreviousPage when Previous button is clicked', async () => {
    const user = userEvent.setup()
    const onPreviousPage = vi.fn()

    render(<JobsTable {...defaultProps} pageIndex={1} onPreviousPage={onPreviousPage} />)

    const prevButton = screen.getByRole('button', { name: /previous/i })
    await user.click(prevButton)

    expect(onPreviousPage).toHaveBeenCalled()
  })

  it('calls onNextPage when Next button is clicked', async () => {
    const user = userEvent.setup()
    const onNextPage = vi.fn()

    render(<JobsTable {...defaultProps} hasNextPage={true} onNextPage={onNextPage} />)

    const nextButton = screen.getByRole('button', { name: /next/i })
    await user.click(nextButton)

    expect(onNextPage).toHaveBeenCalled()
  })

  it('displays correct page size options', async () => {
    const user = userEvent.setup()
    render(<JobsTable {...defaultProps} />)

    const selectTrigger = screen.getByRole('combobox')
    await user.click(selectTrigger)

    expect(screen.getByRole('option', { name: '10' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: '20' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: '30' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: '40' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: '50' })).toBeInTheDocument()
  })

  it('displays run time for each job', () => {
    const jobs = [createMockJob()]
    render(<JobsTable {...defaultProps} jobs={jobs} />)

    expect(screen.getByText('5m 30s')).toBeInTheDocument()
  })
})
