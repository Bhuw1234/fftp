import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { NodesTable } from './NodesTable'
import { models_NodeState, models_NodeType } from '@/lib/api/generated'

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

vi.mock('@/components/nodes/NodeStatus', () => ({
  ConnectionStatus: ({ node }: { node: models_NodeState }) => (
    <span data-testid="connection-status">{node.Connection ? 'Connected' : 'Unknown'}</span>
  ),
  MembershipStatus: ({ node }: { node: models_NodeState }) => (
    <span data-testid="membership-status">{node.Membership ? 'Active' : 'Pending'}</span>
  ),
}))

vi.mock('@/components/nodes/NodeResources', () => ({
  NodeResources: () => <span data-testid="node-resources">4 CPU, 16GB RAM</span>,
}))

vi.mock('@/components/Labels', () => ({
  default: ({ labels }: { labels?: Record<string, string> }) => (
    <span data-testid="labels">{JSON.stringify(labels)}</span>
  ),
}))

vi.mock('@/lib/api/utils', () => ({
  getNodeType: (node: models_NodeState) => node.Info?.NodeType === models_NodeType.NodeTypeCompute ? 'Compute' : 'Orchestrator',
}))

// Create mock node data
const createMockNode = (overrides: Partial<models_NodeState> = {}): models_NodeState => ({
  Info: {
    NodeID: 'n-test-node-id-123',
    NodeType: models_NodeType.NodeTypeCompute,
    Labels: { env: 'test', region: 'us-west' },
  },
  Connection: { connection: 1 },
  Membership: { membership: 1 },
  ...overrides,
})

describe('NodesTable', () => {
  const defaultProps = {
    nodes: [],
    pageSize: 10,
    setPageSize: vi.fn(),
    pageIndex: 0,
    onPreviousPage: vi.fn(),
    onNextPage: vi.fn(),
    hasNextPage: false,
  }

  it('renders empty table when no nodes', () => {
    render(<NodesTable {...defaultProps} />)

    expect(screen.getByText('Node ID')).toBeInTheDocument()
    expect(screen.getByText('Node Type')).toBeInTheDocument()
    expect(screen.getByText('Membership')).toBeInTheDocument()
    expect(screen.getByText('Connection')).toBeInTheDocument()
  })

  it('renders nodes with correct data', () => {
    const nodes = [
      createMockNode({
        Info: {
          NodeID: 'n-abc123',
          NodeType: models_NodeType.NodeTypeCompute,
          Labels: { env: 'prod' },
        },
      }),
    ]

    render(<NodesTable {...defaultProps} nodes={nodes} />)

    expect(screen.getByTestId('connection-status')).toBeInTheDocument()
    expect(screen.getByTestId('membership-status')).toBeInTheDocument()
    expect(screen.getByTestId('node-resources')).toBeInTheDocument()
  })

  it('renders multiple nodes', () => {
    const nodes = [
      createMockNode({ Info: { NodeID: 'n-1', NodeType: models_NodeType.NodeTypeCompute } }),
      createMockNode({ Info: { NodeID: 'n-2', NodeType: models_NodeType.NodeTypeCompute } }),
      createMockNode({ Info: { NodeID: 'n-3', NodeType: models_NodeType.NodeTypeCompute } }),
    ]

    render(<NodesTable {...defaultProps} nodes={nodes} />)

    // Check that we have 3 rows of connection statuses
    const connectionStatuses = screen.getAllByTestId('connection-status')
    expect(connectionStatuses).toHaveLength(3)
  })

  it('renders node links correctly', () => {
    const nodes = [
      createMockNode({
        Info: {
          NodeID: 'n-abc123',
          NodeType: models_NodeType.NodeTypeCompute,
        },
      }),
    ]

    render(<NodesTable {...defaultProps} nodes={nodes} />)

    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/nodes?id=n-abc123')
  })

  it('displays correct node type for Compute node', () => {
    const nodes = [
      createMockNode({
        Info: {
          NodeID: 'n-1',
          NodeType: models_NodeType.NodeTypeCompute,
        },
      }),
    ]

    render(<NodesTable {...defaultProps} nodes={nodes} />)

    expect(screen.getByText('Compute')).toBeInTheDocument()
  })

  it('displays correct node type for Requester (Orchestrator) node', () => {
    const nodes = [
      createMockNode({
        Info: {
          NodeID: 'n-1',
          NodeType: models_NodeType.NodeTypeRequester,
        },
      }),
    ]

    render(<NodesTable {...defaultProps} nodes={nodes} />)

    expect(screen.getByText('Orchestrator')).toBeInTheDocument()
  })

  it('displays labels for each node', () => {
    const nodes = [
      createMockNode({
        Info: {
          NodeID: 'n-1',
          NodeType: models_NodeType.NodeTypeCompute,
          Labels: { env: 'production', region: 'us-east' },
        },
      }),
    ]

    render(<NodesTable {...defaultProps} nodes={nodes} />)

    expect(screen.getByTestId('labels')).toHaveTextContent('env')
    expect(screen.getByTestId('labels')).toHaveTextContent('production')
  })

  it('handles nodes without labels', () => {
    const nodes = [
      createMockNode({
        Info: {
          NodeID: 'n-1',
          NodeType: models_NodeType.NodeTypeCompute,
          Labels: undefined,
        },
      }),
    ]

    render(<NodesTable {...defaultProps} nodes={nodes} />)

    expect(screen.getByTestId('labels')).toBeInTheDocument()
  })

  it('renders with default empty array when nodes is undefined', () => {
    render(<NodesTable {...defaultProps} nodes={undefined as unknown as models_NodeState[]} />)

    expect(screen.getByText('Node ID')).toBeInTheDocument()
  })

  it('displays node resources', () => {
    const nodes = [createMockNode()]

    render(<NodesTable {...defaultProps} nodes={nodes} />)

    expect(screen.getByText('4 CPU, 16GB RAM')).toBeInTheDocument()
  })
})
