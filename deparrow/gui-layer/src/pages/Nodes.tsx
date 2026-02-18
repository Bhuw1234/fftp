import React, { useState, useEffect } from 'react'
import { 
  Server, 
  Monitor, 
  Cpu, 
  HardDrive, 
  Wifi, 
  WifiOff,
  Activity,
  Settings,
  AlertCircle,
  Zap
} from 'lucide-react'
import { nodesAPI } from '../api/client'
import toast from 'react-hot-toast'
import { format } from 'date-fns'

interface Node {
  id: string
  public_key: string
  arch: 'x86_64' | 'arm64'
  status: 'online' | 'offline' | 'maintenance' | 'suspended'
  last_seen: string
  resources: {
    cpu: number
    memory: string
    disk: string
    gpu?: number
  }
  credits_earned: number
  labels: Record<string, string>
  uptime?: string
  location?: string
  version?: string
}

const Nodes: React.FC = () => {
  const [nodes, setNodes] = useState<Node[]>([])
  const [filteredNodes, setFilteredNodes] = useState<Node[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [archFilter, setArchFilter] = useState<string>('all')
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)
  const [showNodeModal, setShowNodeModal] = useState(false)

  useEffect(() => {
    loadNodes()
  }, [])

  useEffect(() => {
    filterNodes()
  }, [nodes, searchTerm, statusFilter, archFilter])

  const loadNodes = async () => {
    try {
      setLoading(true)
      const response = await nodesAPI.list()
      setNodes(response.data.nodes || [])
    } catch (error: any) {
      console.error('Nodes loading error:', error)
      // Load mock data for demonstration
      loadMockNodes()
      toast.error('Using demo data - connect to backend for live data')
    } finally {
      setLoading(false)
    }
  }

  const loadMockNodes = () => {
    const mockNodes: Node[] = [
      {
        id: 'node-001',
        public_key: 'pk_1234567890abcdef...',
        arch: 'x86_64',
        status: 'online',
        last_seen: '2024-01-15T14:30:00Z',
        resources: {
          cpu: 8,
          memory: '16Gi',
          disk: '500Gi',
          gpu: 2
        },
        credits_earned: 125.50,
        labels: {
          region: 'us-east-1',
          type: 'compute',
          cloud: 'aws'
        },
        uptime: '5d 12h 34m',
        location: 'Virginia, US',
        version: '1.0.0'
      },
      {
        id: 'node-002',
        public_key: 'pk_0987654321fedcba...',
        arch: 'arm64',
        status: 'online',
        last_seen: '2024-01-15T14:25:00Z',
        resources: {
          cpu: 4,
          memory: '8Gi',
          disk: '250Gi'
        },
        credits_earned: 89.25,
        labels: {
          region: 'eu-west-1',
          type: 'edge',
          cloud: 'aws'
        },
        uptime: '3d 8h 12m',
        location: 'Ireland, EU',
        version: '1.0.0'
      },
      {
        id: 'node-003',
        public_key: 'pk_1122334455667788...',
        arch: 'x86_64',
        status: 'offline',
        last_seen: '2024-01-15T10:15:00Z',
        resources: {
          cpu: 16,
          memory: '32Gi',
          disk: '1Ti',
          gpu: 4
        },
        credits_earned: 256.75,
        labels: {
          region: 'ap-southeast-1',
          type: 'gpu',
          cloud: 'aws'
        },
        uptime: '2h 45m',
        location: 'Singapore',
        version: '1.0.0'
      },
      {
        id: 'node-004',
        public_key: 'pk_aabbccddeeff0011...',
        arch: 'arm64',
        status: 'maintenance',
        last_seen: '2024-01-15T12:00:00Z',
        resources: {
          cpu: 2,
          memory: '4Gi',
          disk: '100Gi'
        },
        credits_earned: 45.30,
        labels: {
          region: 'us-west-2',
          type: 'small',
          cloud: 'gcp'
        },
        uptime: '1d 6h 20m',
        location: 'California, US',
        version: '0.9.8'
      },
      {
        id: 'node-005',
        public_key: 'pk_9988776655443322...',
        arch: 'x86_64',
        status: 'online',
        last_seen: '2024-01-15T14:32:00Z',
        resources: {
          cpu: 12,
          memory: '24Gi',
          disk: '750Gi',
          gpu: 1
        },
        credits_earned: 178.90,
        labels: {
          region: 'ca-central-1',
          type: 'balanced',
          cloud: 'aws'
        },
        uptime: '4d 2h 18m',
        location: 'Canada Central',
        version: '1.0.0'
      }
    ]
    setNodes(mockNodes)
  }

  const filterNodes = () => {
    let filtered = nodes

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(node =>
        node.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
        node.location?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        Object.values(node.labels).some(label => 
          label.toLowerCase().includes(searchTerm.toLowerCase())
        )
      )
    }

    // Status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(node => node.status === statusFilter)
    }

    // Architecture filter
    if (archFilter !== 'all') {
      filtered = filtered.filter(node => node.arch === archFilter)
    }

    setFilteredNodes(filtered)
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online':
        return <Wifi className="h-5 w-5 text-green-500" />
      case 'offline':
        return <WifiOff className="h-5 w-5 text-red-500" />
      case 'maintenance':
        return <Settings className="h-5 w-5 text-yellow-500" />
      case 'suspended':
        return <AlertCircle className="h-5 w-5 text-gray-500" />
      default:
        return <AlertCircle className="h-5 w-5 text-gray-400" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'bg-green-100 text-green-800'
      case 'offline': return 'bg-red-100 text-red-800'
      case 'maintenance': return 'bg-yellow-100 text-yellow-800'
      case 'suspended': return 'bg-gray-100 text-gray-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getArchIcon = (arch: string) => {
    switch (arch) {
      case 'x86_64':
        return '🏢'
      case 'arm64':
        return '📱'
      default:
        return '💻'
    }
  }

  const getNodeTypeColor = (type: string) => {
    switch (type) {
      case 'gpu': return 'bg-purple-100 text-purple-800'
      case 'compute': return 'bg-blue-100 text-blue-800'
      case 'edge': return 'bg-green-100 text-green-800'
      case 'balanced': return 'bg-indigo-100 text-indigo-800'
      case 'small': return 'bg-gray-100 text-gray-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const formatDate = (dateString: string) => {
    return format(new Date(dateString), 'MMM dd, yyyy HH:mm')
  }

  const formatMemory = (memory: string) => {
    return memory.replace('Gi', ' GB')
  }

  const formatDisk = (disk: string) => {
    if (disk.includes('Ti')) {
      return disk.replace('Ti', ' TB')
    }
    return disk.replace('Gi', ' GB')
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  const onlineNodes = nodes.filter(n => n.status === 'online').length
  const totalCPU = nodes.reduce((sum, n) => sum + n.resources.cpu, 0)
  const totalMemory = nodes.reduce((sum, n) => sum + parseInt(n.resources.memory.replace('Gi', '')), 0)
  const totalGPU = nodes.reduce((sum, n) => sum + (n.resources.gpu || 0), 0)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="sm:flex sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Nodes</h1>
          <p className="text-gray-600">
            Monitor and manage compute nodes in the DEparrow network.
          </p>
        </div>
        <div className="mt-4 sm:mt-0">
          <button
            onClick={() => loadNodes()}
            className="inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            <Activity className="mr-2 h-4 w-4" />
            Refresh
          </button>
        </div>
      </div>

      {/* Network Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Server className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Online Nodes
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {onlineNodes} / {nodes.length}
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Cpu className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Total CPU Cores
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {totalCPU}
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <HardDrive className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Total Memory
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {totalMemory} GB
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Zap className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Total GPUs
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {totalGPU}
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white shadow rounded-lg p-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {/* Search */}
          <div>
            <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-1">
              Search Nodes
            </label>
            <input
              type="text"
              id="search"
              placeholder="Search by ID, location, or label..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
            />
          </div>

          {/* Status Filter */}
          <div>
            <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-1">
              Status
            </label>
            <select
              id="status"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
            >
              <option value="all">All Status</option>
              <option value="online">Online</option>
              <option value="offline">Offline</option>
              <option value="maintenance">Maintenance</option>
              <option value="suspended">Suspended</option>
            </select>
          </div>

          {/* Architecture Filter */}
          <div>
            <label htmlFor="arch" className="block text-sm font-medium text-gray-700 mb-1">
              Architecture
            </label>
            <select
              id="arch"
              value={archFilter}
              onChange={(e) => setArchFilter(e.target.value)}
              className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
            >
              <option value="all">All Architectures</option>
              <option value="x86_64">x86_64</option>
              <option value="arm64">ARM64</option>
            </select>
          </div>

          {/* Summary */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Summary
            </label>
            <div className="text-sm text-gray-600">
              {filteredNodes.length} of {nodes.length} nodes
            </div>
            <div className="text-xs text-gray-500">
              Credits earned: {filteredNodes.reduce((sum, n) => sum + n.credits_earned, 0).toFixed(2)}
            </div>
          </div>
        </div>
      </div>

      {/* Nodes Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredNodes.map((node) => (
          <div key={node.id} className="bg-white shadow rounded-lg p-6 hover:shadow-lg transition-shadow">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-3">
                {getStatusIcon(node.status)}
                <div>
                  <h3 className="text-lg font-medium text-gray-900">{node.id}</h3>
                  <p className="text-sm text-gray-500">{node.location || 'Unknown Location'}</p>
                </div>
              </div>
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(node.status)}`}>
                {node.status}
              </span>
            </div>

            <div className="space-y-3">
              {/* Architecture and Type */}
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <span className="text-lg">{getArchIcon(node.arch)}</span>
                  <span className="text-sm text-gray-600 capitalize">{node.arch.replace('_', '-')}</span>
                </div>
                {node.labels.type && (
                  <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getNodeTypeColor(node.labels.type)}`}>
                    {node.labels.type}
                  </span>
                )}
              </div>

              {/* Resources */}
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div className="flex items-center space-x-2">
                  <Cpu className="h-4 w-4 text-gray-400" />
                  <span className="text-gray-600">{node.resources.cpu} CPU</span>
                </div>
                <div className="flex items-center space-x-2">
                  <HardDrive className="h-4 w-4 text-gray-400" />
                  <span className="text-gray-600">{formatMemory(node.resources.memory)}</span>
                </div>
                <div className="flex items-center space-x-2">
                  <Monitor className="h-4 w-4 text-gray-400" />
                  <span className="text-gray-600">{formatDisk(node.resources.disk)}</span>
                </div>
                {node.resources.gpu && (
                  <div className="flex items-center space-x-2">
                    <Zap className="h-4 w-4 text-gray-400" />
                    <span className="text-gray-600">{node.resources.gpu} GPU</span>
                  </div>
                )}
              </div>

              {/* Credits and Uptime */}
              <div className="border-t pt-3">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-500">Credits Earned:</span>
                  <span className="font-medium text-green-600">{node.credits_earned.toFixed(2)}</span>
                </div>
                {node.uptime && (
                  <div className="flex items-center justify-between text-sm mt-1">
                    <span className="text-gray-500">Uptime:</span>
                    <span className="text-gray-600">{node.uptime}</span>
                  </div>
                )}
                <div className="flex items-center justify-between text-sm mt-1">
                  <span className="text-gray-500">Last Seen:</span>
                  <span className="text-gray-600">{formatDate(node.last_seen)}</span>
                </div>
              </div>

              {/* Labels */}
              {Object.keys(node.labels).length > 0 && (
                <div className="border-t pt-3">
                  <div className="flex flex-wrap gap-1">
                    {Object.entries(node.labels).map(([key, value]) => (
                      <span
                        key={key}
                        className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800"
                      >
                        {key}: {value}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Actions */}
            <div className="mt-4 pt-4 border-t flex justify-end space-x-2">
              <button
                onClick={() => {
                  setSelectedNode(node)
                  setShowNodeModal(true)
                }}
                className="text-blue-600 hover:text-blue-900 text-sm font-medium"
              >
                View Details
              </button>
            </div>
          </div>
        ))}
      </div>

      {filteredNodes.length === 0 && (
        <div className="text-center py-12">
          <Server className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No nodes found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {searchTerm || statusFilter !== 'all' || archFilter !== 'all'
              ? 'Try adjusting your filters or search terms.'
              : 'No nodes are currently registered with the network.'
            }
          </p>
        </div>
      )}

      {/* Node Detail Modal */}
      {showNodeModal && selectedNode && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" onClick={() => setShowNodeModal(false)}></div>
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-2xl sm:w-full">
              <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <div className="sm:flex sm:items-start">
                  <div className="w-full">
                    <div className="flex items-center justify-between mb-4">
                      <h3 className="text-lg leading-6 font-medium text-gray-900">
                        Node Details: {selectedNode.id}
                      </h3>
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(selectedNode.status)}`}>
                        {selectedNode.status}
                      </span>
                    </div>
                    
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      {/* Basic Information */}
                      <div>
                        <h4 className="text-sm font-medium text-gray-900 mb-2">Basic Information</h4>
                        <dl className="space-y-2 text-sm">
                          <div>
                            <dt className="text-gray-500">Node ID</dt>
                            <dd className="text-gray-900 font-mono">{selectedNode.id}</dd>
                          </div>
                          <div>
                            <dt className="text-gray-500">Public Key</dt>
                            <dd className="text-gray-900 font-mono text-xs">{selectedNode.public_key}</dd>
                          </div>
                          <div>
                            <dt className="text-gray-500">Architecture</dt>
                            <dd className="text-gray-900">{selectedNode.arch}</dd>
                          </div>
                          <div>
                            <dt className="text-gray-500">Version</dt>
                            <dd className="text-gray-900">{selectedNode.version || 'Unknown'}</dd>
                          </div>
                          <div>
                            <dt className="text-gray-500">Location</dt>
                            <dd className="text-gray-900">{selectedNode.location || 'Unknown'}</dd>
                          </div>
                          <div>
                            <dt className="text-gray-500">Last Seen</dt>
                            <dd className="text-gray-900">{formatDate(selectedNode.last_seen)}</dd>
                          </div>
                          {selectedNode.uptime && (
                            <div>
                              <dt className="text-gray-500">Uptime</dt>
                              <dd className="text-gray-900">{selectedNode.uptime}</dd>
                            </div>
                          )}
                        </dl>
                      </div>

                      {/* Resources */}
                      <div>
                        <h4 className="text-sm font-medium text-gray-900 mb-2">Resources</h4>
                        <dl className="space-y-2 text-sm">
                          <div>
                            <dt className="text-gray-500">CPU Cores</dt>
                            <dd className="text-gray-900">{selectedNode.resources.cpu}</dd>
                          </div>
                          <div>
                            <dt className="text-gray-500">Memory</dt>
                            <dd className="text-gray-900">{formatMemory(selectedNode.resources.memory)}</dd>
                          </div>
                          <div>
                            <dt className="text-gray-500">Disk</dt>
                            <dd className="text-gray-900">{formatDisk(selectedNode.resources.disk)}</dd>
                          </div>
                          {selectedNode.resources.gpu && (
                            <div>
                              <dt className="text-gray-500">GPU</dt>
                              <dd className="text-gray-900">{selectedNode.resources.gpu}</dd>
                            </div>
                          )}
                          <div>
                            <dt className="text-gray-500">Credits Earned</dt>
                            <dd className="text-green-600 font-medium">{selectedNode.credits_earned.toFixed(2)}</dd>
                          </div>
                        </dl>
                      </div>
                    </div>

                    {/* Labels */}
                    {Object.keys(selectedNode.labels).length > 0 && (
                      <div className="mt-6">
                        <h4 className="text-sm font-medium text-gray-900 mb-2">Labels</h4>
                        <div className="grid grid-cols-2 gap-2">
                          {Object.entries(selectedNode.labels).map(([key, value]) => (
                            <div key={key} className="text-sm">
                              <dt className="text-gray-500">{key}</dt>
                              <dd className="text-gray-900">{value}</dd>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={() => setShowNodeModal(false)}
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Nodes