import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Header } from './Header'

// Mock MobileNav since it has complex dependencies
vi.mock('./MobileNav', () => ({
  MobileNav: () => <div data-testid="mobile-nav">MobileNav</div>,
}))

// Mock ThemeToggle
vi.mock('./ThemeToggle', () => ({
  ThemeToggle: () => <div data-testid="theme-toggle">ThemeToggle</div>,
}))

describe('Header', () => {
  it('renders the header component', () => {
    render(<Header />)

    expect(screen.getByRole('banner')).toBeInTheDocument()
  })

  it('renders the help button', () => {
    render(<Header />)

    const helpButton = screen.getByRole('button', { name: /help menu/i })
    expect(helpButton).toBeInTheDocument()
  })

  it('renders the HelpCircle icon in the button', () => {
    render(<Header />)

    // Check for the help button which contains the icon
    const helpButton = screen.getByRole('button', { name: /help menu/i })
    expect(helpButton).toBeInTheDocument()
  })

  it('opens dropdown menu when help button is clicked', async () => {
    const user = userEvent.setup()
    render(<Header />)

    // Click the help button
    const helpButton = screen.getByRole('button', { name: /help menu/i })
    await user.click(helpButton)

    // Check that documentation link appears
    expect(screen.getByText('Documentation')).toBeInTheDocument()
  })

  it('has correct link to documentation', async () => {
    const user = userEvent.setup()
    render(<Header />)

    // Open the dropdown
    const helpButton = screen.getByRole('button', { name: /help menu/i })
    await user.click(helpButton)

    // Check the link
    const docLink = screen.getByRole('link', { name: /documentation/i })
    expect(docLink).toHaveAttribute('href', 'https://docs.bacalhau.org')
    expect(docLink).toHaveAttribute('target', '_blank')
    expect(docLink).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('renders MobileNav component', () => {
    render(<Header />)

    expect(screen.getByTestId('mobile-nav')).toBeInTheDocument()
  })
})
