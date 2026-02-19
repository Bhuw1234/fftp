import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import Labels from './Labels'

describe('Labels', () => {
  describe('rendering empty or undefined labels', () => {
    it('returns null when labels is undefined', () => {
      const { container } = render(<Labels labels={undefined} />)

      expect(container.firstChild).toBeNull()
    })

    it('returns null when labels is empty array', () => {
      const { container } = render(<Labels labels={[]} />)

      expect(container.firstChild).toBeNull()
    })

    it('returns null when labels is empty object', () => {
      const { container } = render(<Labels labels={{}} />)

      expect(container.firstChild).toBeNull()
    })
  })

  describe('rendering array labels', () => {
    it('renders string labels', () => {
      render(<Labels labels={['label1', 'label2', 'label3']} />)

      expect(screen.getByText('label1')).toBeInTheDocument()
      expect(screen.getByText('label2')).toBeInTheDocument()
      expect(screen.getByText('label3')).toBeInTheDocument()
    })

    it('renders tuple labels with key-value pairs', () => {
      render(<Labels labels={[['env', 'production'], ['region', 'us-west']]} />)

      expect(screen.getByText('env: production')).toBeInTheDocument()
      expect(screen.getByText('region: us-west')).toBeInTheDocument()
    })

    it('renders mixed string and tuple labels', () => {
      render(<Labels labels={['simple-label', ['key', 'value']]} />)

      expect(screen.getByText('simple-label')).toBeInTheDocument()
      expect(screen.getByText('key: value')).toBeInTheDocument()
    })
  })

  describe('rendering object labels', () => {
    it('renders object labels as key-value pairs', () => {
      render(<Labels labels={{ env: 'staging', team: 'platform', version: '1.0' }} />)

      expect(screen.getByText('env: staging')).toBeInTheDocument()
      expect(screen.getByText('team: platform')).toBeInTheDocument()
      expect(screen.getByText('version: 1.0')).toBeInTheDocument()
    })
  })

  describe('custom colors', () => {
    it('applies custom color when provided', () => {
      render(<Labels labels={['test']} color="bg-red-500 text-white" />)

      const badge = screen.getByText('test')
      expect(badge).toHaveClass('bg-red-500')
      expect(badge).toHaveClass('text-white')
    })

    it('uses dynamic color when custom color is not provided', () => {
      render(<Labels labels={['test']} />)

      const badge = screen.getByText('test')
      // Should have some color class (exact color depends on hash)
      expect(badge.className).toMatch(/bg-(blue|green|yellow|purple|pink|indigo)-100/)
    })
  })

  describe('badge styling', () => {
    it('applies text-xs class for small text', () => {
      render(<Labels labels={['test']} />)

      const badge = screen.getByText('test')
      expect(badge).toHaveClass('text-xs')
    })
  })

  describe('flex layout', () => {
    it('renders badges in flex container with gap', () => {
      const { container } = render(<Labels labels={['a', 'b', 'c']} />)

      const wrapper = container.firstChild
      expect(wrapper).toHaveClass('flex')
      expect(wrapper).toHaveClass('flex-wrap')
      expect(wrapper).toHaveClass('gap-2')
    })
  })
})
