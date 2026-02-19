import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button, buttonVariants } from './button'

describe('Button', () => {
  describe('rendering', () => {
    it('renders a button element', () => {
      render(<Button>Click me</Button>)

      const button = screen.getByRole('button', { name: /click me/i })
      expect(button).toBeInTheDocument()
    })

    it('renders children correctly', () => {
      render(
        <Button>
          <span>Icon</span>
          <span>Text</span>
        </Button>
      )

      expect(screen.getByText('Icon')).toBeInTheDocument()
      expect(screen.getByText('Text')).toBeInTheDocument()
    })

    it('forwards ref correctly', () => {
      const ref = { current: null as HTMLButtonElement | null }
      render(<Button ref={ref}>Button</Button>)

      expect(ref.current).toBeInstanceOf(HTMLButtonElement)
    })
  })

  describe('variants', () => {
    it('applies default variant styles', () => {
      render(<Button>Default</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('bg-primary')
    })

    it('applies destructive variant styles', () => {
      render(<Button variant="destructive">Destructive</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('bg-destructive')
    })

    it('applies outline variant styles', () => {
      render(<Button variant="outline">Outline</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('border')
      expect(button).toHaveClass('border-input')
    })

    it('applies secondary variant styles', () => {
      render(<Button variant="secondary">Secondary</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('bg-secondary')
    })

    it('applies ghost variant styles', () => {
      render(<Button variant="ghost">Ghost</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('hover:bg-accent')
    })

    it('applies link variant styles', () => {
      render(<Button variant="link">Link</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('text-primary')
      expect(button).toHaveClass('underline-offset-4')
    })
  })

  describe('sizes', () => {
    it('applies default size', () => {
      render(<Button>Default Size</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('h-9')
      expect(button).toHaveClass('px-4')
    })

    it('applies small size', () => {
      render(<Button size="sm">Small</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('h-8')
      expect(button).toHaveClass('px-3')
    })

    it('applies large size', () => {
      render(<Button size="lg">Large</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('h-10')
      expect(button).toHaveClass('px-8')
    })

    it('applies icon size', () => {
      render(<Button size="icon">Icon</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('h-9')
      expect(button).toHaveClass('w-9')
    })
  })

  describe('interactions', () => {
    it('handles click events', async () => {
      let clicked = false
      const handleClick = () => {
        clicked = true
      }

      const user = userEvent.setup()
      render(<Button onClick={handleClick}>Click me</Button>)

      await user.click(screen.getByRole('button'))

      expect(clicked).toBe(true)
    })

    it('can be disabled', () => {
      render(<Button disabled>Disabled</Button>)

      const button = screen.getByRole('button')
      expect(button).toBeDisabled()
    })

    it('applies disabled styles', () => {
      render(<Button disabled>Disabled</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('disabled:pointer-events-none')
      expect(button).toHaveClass('disabled:opacity-50')
    })
  })

  describe('asChild prop', () => {
    it('renders as button by default', () => {
      render(<Button>Button</Button>)

      const button = screen.getByRole('button')
      expect(button.tagName).toBe('BUTTON')
    })

    it('renders as child component when asChild is true', () => {
      render(
        <Button asChild>
          <a href="/test">Link Button</a>
        </Button>
      )

      const link = screen.getByRole('link', { name: /link button/i })
      expect(link).toBeInTheDocument()
      expect(link.tagName).toBe('A')
    })
  })

  describe('custom className', () => {
    it('merges custom className with default styles', () => {
      render(<Button className="custom-class">Custom</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveClass('custom-class')
      expect(button).toHaveClass('bg-primary') // Still has default styles
    })
  })

  describe('buttonVariants function', () => {
    it('returns variant classes without rendering a component', () => {
      const classes = buttonVariants({ variant: 'destructive', size: 'sm' })

      expect(classes).toContain('bg-destructive')
      expect(classes).toContain('h-8')
    })

    it('returns default variant classes when no options provided', () => {
      const classes = buttonVariants()

      expect(classes).toContain('bg-primary')
      expect(classes).toContain('h-9')
    })
  })

  describe('accessibility', () => {
    it('can have aria-label', () => {
      render(<Button aria-label="Close dialog">×</Button>)

      const button = screen.getByRole('button', { name: /close dialog/i })
      expect(button).toBeInTheDocument()
    })

    it('supports type attribute', () => {
      render(<Button type="submit">Submit</Button>)

      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('type', 'submit')
    })
  })
})
