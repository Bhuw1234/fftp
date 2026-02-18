import React, { useState, useEffect } from 'react'
import { 
  Server, 
  Cpu, 
  HardDrive, 
  Zap,
  Star,
  Search,
  DollarSign,
  Activity,
  Wifi,
  AlertCircle,
  XCircle
} from 'lucide-react'
import { providersAPI } from '../api/deparrow'
import toast from 'react-hot-toast'
import { format } from 'date-fns'

interface Provider {
  id: string
  name: string
  status: 'online' | 'offline' | 'maintenance' | 'busy'
  location: {
    region: string
    country: string
    city: string
    lat?: number
    lon?: number
  }
  resources: {
    cpu_cores: number
    cpu_available: number
    memory_total: string
    memory_available: string
    gpu_count: number
    gpu_available: number
    storage_total: string
    storage_available: string
  }
  pricing: {
    cpu_per_hour: number
    memory_per_gb_hour: number
    gpu_per_hour: number
    storage_per_gb_month: number
  }
  stats: {
    jobs_completed: number
    success_rate: number
    avg_response_time: number
    uptime_30d: number
    total_credits_earned: number
    reputation: number
  }
  labels: Record<string, string>
  last_job?: string
  joined_at: string
}

const Providers: React.FC = () => {
  const [providers, setProviders] = useState<Provider[]>([])
  const [filteredProviders, setFilteredProviders] = useState<Provider[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [regionFilter, setRegionFilter] = useState<string>('all')
  const [gpuFilter, setGpuFilter] = useState<boolean>(false)
  const [sortBy, setSortBy] = useState<string>('reputation')
  const [selectedProvider, setSelectedProvider] = useState<Provider | null>(null)
  const [showProviderModal, setShowProviderModal] = useState(false)

  useEffect(() => {
    loadProviders()
  }, [])

  useEffect(() => {
    filterAndSortProviders()
  }, [providers, searchTerm, statusFilter, regionFilter, gpuFilter, sortBy])

  const loadProviders = async () => {
    try {
      setLoading(true)
      const response = await providersAPI.list()
      setProviders(response.data.providers || [])
    } catch (error: any) {
      console.error('Providers loading error:', error)
      loadMockProviders()
      toast.error('Using demo data - connect to backend for live data')
    } finally {
      setLoading(false)
    }
  }

  const loadMockProviders = () => {
    const mockProviders: Provider[] = [
      {
        id: 'provider-001',
        name: 'US-East GPU Cluster',
        status: 'online',
        location: {
          region: 'us-east-1',
          country: 'USA',
          city: 'Virginia'
        },
        resources: {
          cpu_cores: 64,
          cpu_available: 42,
          memory_total: '256Gi',
          memory_available: '128Gi',
          gpu_count: 8,
          gpu_available: 5,
          storage_total: '4Ti',
          storage_available: '2.5Ti'
        },
        pricing: {
          cpu_per_hour: 0.05,
          memory_per_gb_hour: 0.01,
          gpu_per_hour: 0.50,
          storage_per_gb_month: 0.02
        },
        stats: {
          jobs_completed: 1245,
          success_rate: 99.2,
          avg_response_time: 1.2,
          uptime_30d: 99.9,
          total_credits_earned: 4567.89,
          reputation: 4.8
        },
        labels: {
          tier: 'premium',
          specialization: 'gpu',
          network: '10gbps'
        },
        joined_at: '2024-01-01T00:00:00Z'
      },
      {
        id: 'provider-002',
        name: 'EU-West Compute Farm',
        status: 'online',
        location: {
          region: 'eu-west-1',
          country: 'Ireland',
          city: 'Dublin'
        },
        resources: {
          cpu_cores: 128,
          cpu_available: 96,
          memory_total: '512Gi',
          memory_available: '384Gi',
          gpu_count: 0,
          gpu_available: 0,
          storage_total: '8Ti',
          storage_available: '6Ti'
        },
        pricing: {
          cpu_per_hour: 0.03,
          memory_per_gb_hour: 0.008,
          gpu_per_hour: 0,
          storage_per_gb_month: 0.015
        },
        stats: {
          jobs_completed: 3421,
          success_rate: 98.7,
          avg_response_time: 0.8,
          uptime_30d: 99.5,
          total_credits_earned: 8934.12,
          reputation: 4.9
        },
        labels: {
          tier: 'standard',
          specialization: 'cpu',
          network: '10gbps'
        },
        joined_at: '2023-11-15T00:00:00Z'
      },
      {
        id: 'provider-003',
        name: 'Asia-Pacific Edge Node',
        status: 'busy',
        location: {
          region: 'ap-southeast-1',
          country: 'Singapore',
          city: 'Singapore'
        },
        resources: {
          cpu_cores: 32,
          cpu_available: 4,
          memory_total: '128Gi',
          memory_available: '16Gi',
          gpu_count: 4,
          gpu_available: 0,
          storage_total: '2Ti',
          storage_available: '500Gi'
        },
        pricing: {
          cpu_per_hour: 0.04,
          memory_per_gb_hour: 0.012,
          gpu_per_hour: 0.45,
          storage_per_gb_month: 0.018
        },
        stats: {
          jobs_completed: 876,
          success_rate: 97.5,
          avg_response_time: 2.1,
          uptime_30d: 98.2,
          total_credits_earned: 2345.67,
          reputation: 4.6
        },
        labels: {
          tier: 'standard',
          specialization: 'edge',
          network: '1gbps'
        },
        joined_at: '2024-02-01T00:00:00Z'
      },
      {
        id: 'provider-004',
        name: 'US-West High Memory',
        status: 'online',
        location: {
          region: 'us-west-2',
          country: 'USA',
          city: 'Oregon'
        },
        resources: {
          cpu_cores: 48,
          cpu_available: 36,
          memory_total: '768Gi',
          memory_available: '512Gi',
          gpu_count: 0,
          gpu_available: 0,
          storage_total: '2Ti',
          storage_available: '1.5Ti'
        },
        pricing: {
          cpu_per_hour: 0.04,
          memory_per_gb_hour: 0.006,
          gpu_per_hour: 0,
          storage_per_gb_month: 0.02
        },
        stats: {
          jobs_completed: 1567,
          success_rate: 99.1,
          avg_response_time: 1.5,
          uptime_30d: 99.7,
          total_credits_earned: 5678.90,
          reputation: 4.7
        },
        labels: {
          tier: 'premium',
          specialization: 'memory',
          network: '10gbps'
        },
        joined_at: '2023-12-01T00:00:00Z'
      },
      {
        id: 'provider-005',
        name: 'EU-Central Budget',
        status: 'maintenance',
        location: {
          region: 'eu-central-1',
          country: 'Germany',
          city: 'Frankfurt'
        },
        resources: {
          cpu_cores: 16,
          cpu_available: 0,
          memory_total: '64Gi',
          memory_available: '0Gi',
          gpu_count: 2,
          gpu_available: 0,
          storage_total: '1Ti',
          storage_available: '0Gi'
        },
        pricing: {
          cpu_per_hour: 0.02,
          memory_per_gb_hour: 0.005,
          gpu_per_hour: 0.30,
          storage_per_gb_month: 0.01
        },
        stats: {
          jobs_completed: 523,
          success_rate: 96.8,
          avg_response_time: 3.2,
          uptime_30d: 95.5,
          total_credits_earned: 1234.56,
          reputation: 4.2
        },
        labels: {
          tier: 'budget',
          specialization: 'general',
          network: '1gbps'
        },
        joined_at: '2024-01-15T00:00:00Z'
      }
    ]
    setProviders(mockProviders)
  }

  const filterAndSortProviders = () => {
    let filtered = providers

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(p =>
        p.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        p.location.city.toLowerCase().includes(searchTerm.toLowerCase()) ||
        p.location.country.toLowerCase().includes(searchTerm.toLowerCase()) ||
        p.id.toLowerCase().includes(searchTerm.toLowerCase())
      )
    }

    // Status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(p => p.status === statusFilter)
    }

    // Region filter
    if (regionFilter !== 'all') {
      filtered = filtered.filter(p => p.location.region.startsWith(regionFilter))
    }

    // GPU filter
    if (gpuFilter) {
      filtered = filtered.filter(p => p.resources.gpu_count > 0)
    }

    // Sort
    switch (sortBy) {
      case 'reputation':
        filtered.sort((a, b) => b.stats.reputation - a.stats.reputation)
        break
      case 'price':
        filtered.sort((a, b) => a.pricing.cpu_per_hour - b.pricing.cpu_per_hour)
        break
      case 'uptime':
        filtered.sort((a, b) => b.stats.uptime_30d - a.stats.uptime_30d)
        break
      case 'jobs':
        filtered.sort((a, b) => b.stats.jobs_completed - a.stats.jobs_completed)
        break
    }

    setFilteredProviders(filtered)
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online':
        return <Wifi className="h-5 w-5 text-green-500" />
      case 'busy':
        return <Activity className="h-5 w-5 text-yellow-500" />
      case 'maintenance':
        return <AlertCircle className="h-5 w-5 text-orange-500" />
      case 'offline':
        return <XCircle className="h-5 w-5 text-red-500" />
      default:
        return <AlertCircle className="h-5 w-5 text-gray-400" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'bg-green-100 text-green-800'
      case 'busy': return 'bg-yellow-100 text-yellow-800'
      case 'maintenance': return 'bg-orange-100 text-orange-800'
      case 'offline': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getRegionFlag = (region: string) => {
    if (region.startsWith('us')) return '🇺🇸'
    if (region.startsWith('eu')) return '🇪🇺'
    if (region.startsWith('ap')) return '🌏'
    return '🌐'
  }

  const formatDate = (dateString: string) => {
    return format(new Date(dateString), 'MMM dd, yyyy')
  }

  const calculateEstimatedCost = (provider: Provider, cpuHours: number = 1, memoryGb: number = 4, gpuHours: number = 0) => {
    const cpuCost = cpuHours * provider.pricing.cpu_per_hour
    const memoryCost = memoryGb * provider.pricing.memory_per_gb_hour * cpuHours
    const gpuCost = gpuHours * provider.pricing.gpu_per_hour
    return (cpuCost + memoryCost + gpuCost).toFixed(4)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  const onlineProviders = providers.filter(p => p.status === 'online').length
  const totalCPU = providers.reduce((sum, p) => sum + p.resources.cpu_cores, 0)
  const totalGPU = providers.reduce((sum, p) => sum + p.resources.gpu_count, 0)
  const avgReputation = providers.length > 0 
    ? (providers.reduce((sum, p) => sum + p.stats.reputation, 0) / providers.length).toFixed(1)
    : '0.0'

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="sm:flex sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Compute Providers</h1>
          <p className="text-gray-600">
            Discover and compare compute providers in the DEparrow network.
          </p>
        </div>
        <div className="mt-4 sm:mt-0">
          <button
            onClick={() => loadProviders()}
            className="inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
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
                    Online Providers
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {onlineProviders} / {providers.length}
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
                    {totalCPU.toLocaleString()}
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

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Star className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Avg Reputation
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {avgReputation} ⭐
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white shadow rounded-lg p-6">
        <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          {/* Search */}
          <div className="md:col-span-2">
            <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-1">
              Search Providers
            </label>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                type="text"
                id="search"
                placeholder="Search by name, location..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
              />
            </div>
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
              <option value="busy">Busy</option>
              <option value="maintenance">Maintenance</option>
              <option value="offline">Offline</option>
            </select>
          </div>

          {/* Region Filter */}
          <div>
            <label htmlFor="region" className="block text-sm font-medium text-gray-700 mb-1">
              Region
            </label>
            <select
              id="region"
              value={regionFilter}
              onChange={(e) => setRegionFilter(e.target.value)}
              className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
            >
              <option value="all">All Regions</option>
              <option value="us">Americas</option>
              <option value="eu">Europe</option>
              <option value="ap">Asia-Pacific</option>
            </select>
          </div>

          {/* Sort */}
          <div>
            <label htmlFor="sort" className="block text-sm font-medium text-gray-700 mb-1">
              Sort By
            </label>
            <select
              id="sort"
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              className="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
            >
              <option value="reputation">Reputation</option>
              <option value="price">Price (Low)</option>
              <option value="uptime">Uptime</option>
              <option value="jobs">Jobs Completed</option>
            </select>
          </div>
        </div>

        <div className="mt-4 flex items-center space-x-4">
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={gpuFilter}
              onChange={(e) => setGpuFilter(e.target.checked)}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
            <span className="text-sm text-gray-600">GPU Available Only</span>
          </label>
          <span className="text-sm text-gray-500">
            {filteredProviders.length} of {providers.length} providers
          </span>
        </div>
      </div>

      {/* Providers List */}
      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Provider
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Location
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Resources
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Pricing
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Stats
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredProviders.map((provider) => (
                <tr key={provider.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      {getStatusIcon(provider.status)}
                      <div className="ml-3">
                        <div className="text-sm font-medium text-gray-900">{provider.name}</div>
                        <div className="text-xs text-gray-500">{provider.id}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <span className="mr-2 text-lg">{getRegionFlag(provider.location.region)}</span>
                      <div>
                        <div className="text-sm text-gray-900">{provider.location.city}</div>
                        <div className="text-xs text-gray-500">{provider.location.country}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">
                      <div className="flex items-center space-x-3">
                        <span className="flex items-center">
                          <Cpu className="h-4 w-4 text-gray-400 mr-1" />
                          {provider.resources.cpu_available}/{provider.resources.cpu_cores}
                        </span>
                        <span className="flex items-center">
                          <HardDrive className="h-4 w-4 text-gray-400 mr-1" />
                          {provider.resources.memory_available}
                        </span>
                        {provider.resources.gpu_count > 0 && (
                          <span className="flex items-center">
                            <Zap className="h-4 w-4 text-purple-400 mr-1" />
                            {provider.resources.gpu_available}/{provider.resources.gpu_count}
                          </span>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">
                      <div className="flex items-center">
                        <DollarSign className="h-4 w-4 text-gray-400" />
                        <span className="text-green-600 font-medium">
                          ${provider.pricing.cpu_per_hour.toFixed(2)}/CPU/hr
                        </span>
                      </div>
                      {provider.resources.gpu_count > 0 && (
                        <div className="text-xs text-gray-500">
                          GPU: ${provider.pricing.gpu_per_hour.toFixed(2)}/hr
                        </div>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">
                      <div className="flex items-center space-x-1">
                        <Star className="h-4 w-4 text-yellow-400 fill-current" />
                        <span className="font-medium">{provider.stats.reputation}</span>
                      </div>
                      <div className="text-xs text-gray-500">
                        {provider.stats.jobs_completed.toLocaleString()} jobs • {provider.stats.uptime_30d}% uptime
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button
                      onClick={() => {
                        setSelectedProvider(provider)
                        setShowProviderModal(true)
                      }}
                      className="text-blue-600 hover:text-blue-900 mr-3"
                    >
                      Details
                    </button>
                    <button className="text-green-600 hover:text-green-900">
                      Submit Job
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {filteredProviders.length === 0 && (
        <div className="text-center py-12">
          <Server className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No providers found</h3>
          <p className="mt-1 text-sm text-gray-500">
            Try adjusting your filters or search terms.
          </p>
        </div>
      )}

      {/* Provider Detail Modal */}
      {showProviderModal && selectedProvider && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" onClick={() => setShowProviderModal(false)}></div>
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-3xl sm:w-full">
              <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-3">
                    {getStatusIcon(selectedProvider.status)}
                    <div>
                      <h3 className="text-lg font-medium text-gray-900">{selectedProvider.name}</h3>
                      <p className="text-sm text-gray-500">{selectedProvider.id}</p>
                    </div>
                  </div>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(selectedProvider.status)}`}>
                    {selectedProvider.status}
                  </span>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {/* Location & Info */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-900 mb-3">Location</h4>
                    <div className="bg-gray-50 rounded-lg p-4 space-y-2">
                      <div className="flex items-center space-x-2">
                        <span className="text-2xl">{getRegionFlag(selectedProvider.location.region)}</span>
                        <div>
                          <p className="text-sm font-medium text-gray-900">{selectedProvider.location.city}, {selectedProvider.location.country}</p>
                          <p className="text-xs text-gray-500">{selectedProvider.location.region}</p>
                        </div>
                      </div>
                      <div className="text-xs text-gray-500">
                        Joined: {formatDate(selectedProvider.joined_at)}
                      </div>
                    </div>

                    <h4 className="text-sm font-medium text-gray-900 mt-4 mb-3">Resources Available</h4>
                    <div className="bg-gray-50 rounded-lg p-4 space-y-2">
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-500">CPU</span>
                        <span className="font-medium">{selectedProvider.resources.cpu_available} / {selectedProvider.resources.cpu_cores} cores</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-500">Memory</span>
                        <span className="font-medium">{selectedProvider.resources.memory_available} / {selectedProvider.resources.memory_total}</span>
                      </div>
                      {selectedProvider.resources.gpu_count > 0 && (
                        <div className="flex justify-between text-sm">
                          <span className="text-gray-500">GPU</span>
                          <span className="font-medium">{selectedProvider.resources.gpu_available} / {selectedProvider.resources.gpu_count}</span>
                        </div>
                      )}
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-500">Storage</span>
                        <span className="font-medium">{selectedProvider.resources.storage_available} / {selectedProvider.resources.storage_total}</span>
                      </div>
                    </div>
                  </div>

                  {/* Pricing & Stats */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-900 mb-3">Pricing</h4>
                    <div className="bg-green-50 rounded-lg p-4 space-y-2">
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-600">CPU</span>
                        <span className="font-medium text-green-700">${selectedProvider.pricing.cpu_per_hour.toFixed(3)}/hour</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-600">Memory</span>
                        <span className="font-medium text-green-700">${selectedProvider.pricing.memory_per_gb_hour.toFixed(4)}/GB/hour</span>
                      </div>
                      {selectedProvider.resources.gpu_count > 0 && (
                        <div className="flex justify-between text-sm">
                          <span className="text-gray-600">GPU</span>
                          <span className="font-medium text-green-700">${selectedProvider.pricing.gpu_per_hour.toFixed(2)}/hour</span>
                        </div>
                      )}
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-600">Storage</span>
                        <span className="font-medium text-green-700">${selectedProvider.pricing.storage_per_gb_month.toFixed(3)}/GB/month</span>
                      </div>
                      <div className="border-t border-green-200 pt-2 mt-2">
                        <div className="text-xs text-gray-600 mb-1">Est. cost for 1 CPU, 4GB RAM, 1hr:</div>
                        <div className="text-lg font-bold text-green-700">${calculateEstimatedCost(selectedProvider)} credits</div>
                      </div>
                    </div>

                    <h4 className="text-sm font-medium text-gray-900 mt-4 mb-3">Statistics</h4>
                    <div className="bg-gray-50 rounded-lg p-4 space-y-2">
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-500">Jobs Completed</span>
                        <span className="font-medium">{selectedProvider.stats.jobs_completed.toLocaleString()}</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-500">Success Rate</span>
                        <span className="font-medium text-green-600">{selectedProvider.stats.success_rate}%</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-500">Avg Response</span>
                        <span className="font-medium">{selectedProvider.stats.avg_response_time}s</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-500">30d Uptime</span>
                        <span className="font-medium">{selectedProvider.stats.uptime_30d}%</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-500">Reputation</span>
                        <span className="font-medium flex items-center">
                          <Star className="h-4 w-4 text-yellow-400 fill-current mr-1" />
                          {selectedProvider.stats.reputation}
                        </span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-500">Credits Earned</span>
                        <span className="font-medium text-green-600">{selectedProvider.stats.total_credits_earned.toFixed(2)}</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Labels */}
                {Object.keys(selectedProvider.labels).length > 0 && (
                  <div className="mt-4">
                    <h4 className="text-sm font-medium text-gray-900 mb-2">Labels</h4>
                    <div className="flex flex-wrap gap-2">
                      {Object.entries(selectedProvider.labels).map(([key, value]) => (
                        <span
                          key={key}
                          className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800"
                        >
                          {key}: {value}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
                >
                  Submit Job to Provider
                </button>
                <button
                  type="button"
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={() => setShowProviderModal(false)}
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

export default Providers
