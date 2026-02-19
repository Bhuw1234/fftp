import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import TruncatedTextWithTooltip from './TruncatedTextWithTooltip'

// Mock the tooltip component to avoid radix-ui portal issues in tests
vi.mock('@/components/ui/tooltip', () => ({
  Tooltip: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TooltipTrigger: ({ children }: { children: React.ReactNode }) => <span>{children}</span>,
  TooltipContent: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="tooltip-content">{children}</div>
  ),
  TooltipProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

describe('TruncatedTextWithTooltip', () => {
  describe('rendering text', () => {
    it('returns null when text is undefined', () => {
      const { container } = render(<TruncatedTextWithTooltip text={undefined} />)

      expect(container.firstChild).toBeNull()
    })

    it('returns null when text is empty string', () => {
      const { container } = render(<TruncatedTextWithTooltip text="" />)

      expect(container.firstChild).toBeNull()
    })

    it('renders short text without truncation', () => {
      render(<TruncatedTextWithTooltip text="short text" />)

      expect(screen.getByText('short text')).toBeInTheDocument()
    })

    it('renders long text with truncation', () => {
      const longText = 'This is a very long text that should be truncated'
      render(<TruncatedTextWithTooltip text={longText} maxLength={20} />)

      // Should show truncated text with ellipsis
      expect(screen.getByText('This is a very long ...')).toBeInTheDocument()
    })
  })

  describe('maxLength parameter', () => {
    it('uses default maxLength of 50 when not specified', () => {
      const textExactly50 = 'a'.repeat(50)
      render(<TruncatedTextWithTooltip text={textExactly50} />)

      // Should not truncate at exactly 50 characters
      expect(screen.getByText(textExactly50)).toBeInTheDocument()
    })

    it('truncates at 51 characters with default maxLength', () => {
      const text51 = 'a'.repeat(51)
      render(<TruncatedTextWithTooltip text={text51} />)

      // Should truncate
      expect(screen.getByText('a'.repeat(50) + '...')).toBeInTheDocument()
    })

    it('respects custom maxLength', () => {
      const text = 'This is a test text'
      render(<TruncatedTextWithTooltip text={text} maxLength={10} />)

      expect(screen.getByText('This is a ...')).toBeInTheDocument()
    })
  })

  describe('tooltip behavior', () => {
    it('does not show tooltip for short text', () => {
      render(<TruncatedTextWithTooltip text="short" maxLength={20} />)

      // No tooltip content should be rendered
      expect(screen.queryByTestId('tooltip-content')).not.toBeInTheDocument()
    })

    it('shows tooltip for truncated text', async () => {
      const user = userEvent.setup()
      const longText = 'This is a very long text that needs truncation'
      render(<TruncatedTextWithTooltip text={longText} maxLength={20} />)

      // Tooltip content should be present in the DOM
      expect(screen.getByTestId('tooltip-content')).toBeInTheDocument()
      expect(screen.getByText(longText)).toBeInTheDocument()
    })
  })

  describe('text display', () => {
    it('displays text in a span element', () => {
      const { container } = render(<TruncatedTextWithTooltip text="test" />)

      const span = container.querySelector('span')
      expect(span).toHaveTextContent('test')
    })

    it('preserves the original text in tooltip', () => {
      const longText = 'This is the complete text that should appear in tooltip'
      render(<TruncatedTextWithTooltip text={longText} maxLength={20} />)

      // Tooltip should contain the full text
      expect(screen.getByText(longText)).toBeInTheDocument()
    })
  })
})
