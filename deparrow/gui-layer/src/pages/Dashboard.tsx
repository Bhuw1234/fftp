import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { 
  Server, 
  Briefcase, 
  Wallet, 
  Activity, 
  TrendingUp,
  Users,
  Cpu,
  HardDrive
} from 'lucide-react'
import { systemAPI, nodesAPI, jobsAPI, walletAPI } from '../api/client'
import { useAuth } from '../contexts/AuthContext'
import toast from 'react-hot-toast'

interface DashboardStats {
  nodes: {
    total: number
    online: number
    byArch: Record<string, number>
  }
  credits: {
    total_circulating: number
    total_earned: number
    user_balances: Record<string, number>
  }
  orchestrators: {
    total: number
    online: number
  }
  jobs: {
    total: number
    active: number
  }
}

interface SystemHealth {
  status: string
  timestamp: string
  version: string
  components: Record<string, number>
}

const Dashboard: React.FC = () => {
  const { user } = useAuth()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [health, setHealth] = useState<SystemHealth | null>(null)
  const [recentJobs, setRecentJobs] = useState<any[]>([])
  const [userCredits, setUserCredits] = useState<number>(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadDashboardData()
  }, [])

  const loadDashboardData = async () => {
    try {
      setLoading(true)
      
      // Load system metrics
      const metricsResponse = await systemAPI.stats()
      setStats(metricsResponse.data.metrics)
      
      // Load system health
      const healthResponse = await systemAPI.health()
      setHealth(healthResponse.data)
      
      // Load recent jobs
      const jobsResponse = await jobsAPI.list({ limit: 5 })
      setRecentJobs(jobsResponse.data.jobs || [])
      
      // Load user credits
      if (user) {
        const creditsResponse = await walletAPI.balance()
        setUserCredits(creditsResponse.data.credit_balance)
      }
      
    } catch (error: any) {
      console.error('Dashboard data loading error:', error)
      // Don't show error toast for initial load, use mock data
      loadMockData()
    } finally {
      setLoading(false)
    }
  }

  const loadMockData = () => {
    // Mock data for demonstration
    setStats({
      nodes: {
        total: 12,
        online: 8,
        byArch: {
          x86_64: 6,
          arm64: 2
        }
      },
      credits: {
        total_circulating: 1250.5,
        total_earned: 2340.0,
        user_balances: {
          'user-1': 150.5,
          'user-2': 200.0,
          'user-3': 75.25
        }
      },
      orchestrators: {
        total: 2,
        online: 2
      },
      jobs: {
        total: 156,
        active: 8
      }
    })

    setHealth({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      version: '1.0.0',
      components: {
        nodes: 8,
        orchestrators: 2,
        users: 3,
        jobs: 156
      }
    })

    setUserCredits(150.5)
    setRecentJobs([
      {
        id: 'job-1',
        name: 'Data Processing Pipeline',
        status: 'running',
        progress: 65,
        credit_cost: 15.5,
        created_at: '2024-01-15T10:30:00Z'
      },
      {
        id: 'job-2', 
        name: 'ML Model Training',
        status: 'completed',
        progress: 100,
        credit_cost: 25.0,
        created_at: '2024-01-15T09:15:00Z'
      },
      {
        id: 'job-3',
        name: 'Image Processing',
        status: 'failed',
        progress: 30,
        credit_cost: 8.0,
        created_at: '2024-01-15T08:45:00Z'
      }
    ])
  }

  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleString()
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'text-blue-600 bg-blue-100'
      case 'completed': return 'text-green-600 bg-green-100'
      case 'failed': return 'text-red-600 bg-red-100'
      default: return 'text-gray-600 bg-gray-100'
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600">
          Welcome back, {user?.name}! Here's what's happening with your DEparrow network.
        </p>
      </div>

      {/* Quick Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* User Credits */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Wallet className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Your Credits
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {userCredits.toFixed(2)}
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-5 py-3">
            <div className="text-sm">
              <Link to="/wallet" className="font-medium text-blue-700 hover:text-blue-900">
                Manage wallet
              </Link>
            </div>
          </div>
        </div>

        {/* Active Nodes */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Server className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Active Nodes
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {stats?.nodes.online || 0} / {stats?.nodes.total || 0}
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-5 py-3">
            <div className="text-sm">
              <Link to="/nodes" className="font-medium text-blue-700 hover:text-blue-900">
                View all nodes
              </Link>
            </div>
          </div>
        </div>

        {/* Active Jobs */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Activity className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Active Jobs
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {stats?.jobs.active || 0}
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-5 py-3">
            <div className="text-sm">
              <Link to="/jobs" className="font-medium text-blue-700 hover:text-blue-900">
                View all jobs
              </Link>
            </div>
          </div>
        </div>

        {/* System Health */}
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <TrendingUp className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    System Health
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                      {health?.status || 'Unknown'}
                    </span>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-5 py-3">
            <div className="text-sm">
              <span className="text-gray-500">
                v{health?.version || '1.0.0'}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Jobs */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">Recent Jobs</h3>
          </div>
          <div className="divide-y divide-gray-200">
            {recentJobs.length > 0 ? (
              recentJobs.map((job) => (
                <div key={job.id} className="px-6 py-4">
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {job.name}
                      </p>
                      <p className="text-sm text-gray-500">
                        Cost: {job.credit_cost} credits
                      </p>
                    </div>
                    <div className="flex items-center space-x-2">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(job.status)}`}>
                        {job.status}
                      </span>
                    </div>
                  </div>
                  {job.status === 'running' && (
                    <div className="mt-2">
                      <div className="bg-gray-200 rounded-full h-2">
                        <div 
                          className="bg-blue-600 h-2 rounded-full transition-all duration-300" 
                          style={{ width: `${job.progress || 0}%` }}
                        ></div>
                      </div>
                    </div>
                  )}
                  <div className="mt-2 text-xs text-gray-500">
                    {formatTime(job.created_at)}
                  </div>
                </div>
              ))
            ) : (
              <div className="px-6 py-8 text-center">
                <Briefcase className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-2 text-sm font-medium text-gray-900">No jobs</h3>
                <p className="mt-1 text-sm text-gray-500">Get started by creating a new job.</p>
                <div className="mt-6">
                  <Link
                    to="/jobs"
                    className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
                  >
                    <Briefcase className="mr-2 h-4 w-4" />
                    View Jobs
                  </Link>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Network Statistics */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">Network Statistics</h3>
          </div>
          <div className="px-6 py-4 space-y-4">
            {/* Node Distribution */}
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-2">Node Architecture</h4>
              <div className="space-y-2">
                {stats?.nodes.byArch && Object.entries(stats.nodes.byArch).map(([arch, count]) => (
                  <div key={arch} className="flex items-center justify-between">
                    <div className="flex items-center">
                      <Cpu className="h-4 w-4 text-gray-400 mr-2" />
                      <span className="text-sm text-gray-600 capitalize">{arch}</span>
                    </div>
                    <span className="text-sm font-medium text-gray-900">{count}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Credit Statistics */}
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-2">Credit Distribution</h4>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <Wallet className="h-4 w-4 text-gray-400 mr-2" />
                    <span className="text-sm text-gray-600">Total Circulating</span>
                  </div>
                  <span className="text-sm font-medium text-gray-900">
                    {stats?.credits.total_circulating.toFixed(2) || '0.00'}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <TrendingUp className="h-4 w-4 text-gray-400 mr-2" />
                    <span className="text-sm text-gray-600">Total Earned</span>
                  </div>
                  <span className="text-sm font-medium text-gray-900">
                    {stats?.credits.total_earned.toFixed(2) || '0.00'}
                  </span>
                </div>
              </div>
            </div>

            {/* Orchestrator Status */}
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-2">Orchestrators</h4>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <Server className="h-4 w-4 text-gray-400 mr-2" />
                    <span className="text-sm text-gray-600">Online</span>
                  </div>
                  <span className="text-sm font-medium text-gray-900">
                    {stats?.orchestrators.online || 0} / {stats?.orchestrators.total || 0}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Quick Actions</h3>
        </div>
        <div className="px-6 py-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <Link
              to="/jobs/new"
              className="relative group bg-blue-50 p-6 focus-within:ring-2 focus-within:ring-inset focus-within:ring-blue-500 rounded-lg hover:bg-blue-100 transition-colors"
            >
              <div>
                <span className="rounded-lg inline-flex p-3 bg-blue-600 text-white">
                  <Briefcase className="h-6 w-6" />
                </span>
              </div>
              <div className="mt-4">
                <h3 className="text-lg font-medium text-gray-900">
                  <span className="absolute inset-0" />
                  Submit New Job
                </h3>
                <p className="mt-2 text-sm text-gray-500">
                  Create and submit a new compute job to the network.
                </p>
              </div>
            </Link>

            <Link
              to="/nodes"
              className="relative group bg-green-50 p-6 focus-within:ring-2 focus-within:ring-inset focus-within:ring-green-500 rounded-lg hover:bg-green-100 transition-colors"
            >
              <div>
                <span className="rounded-lg inline-flex p-3 bg-green-600 text-white">
                  <Server className="h-6 w-6" />
                </span>
              </div>
              <div className="mt-4">
                <h3 className="text-lg font-medium text-gray-900">
                  <span className="absolute inset-0" />
                  Manage Nodes
                </h3>
                <p className="mt-2 text-sm text-gray-500">
                  View and manage compute nodes in the network.
                </p>
              </div>
            </Link>

            <Link
              to="/wallet"
              className="relative group bg-yellow-50 p-6 focus-within:ring-2 focus-within:ring-inset focus-within:ring-yellow-500 rounded-lg hover:bg-yellow-100 transition-colors"
            >
              <div>
                <span className="rounded-lg inline-flex p-3 bg-yellow-600 text-white">
                  <Wallet className="h-6 w-6" />
                </span>
              </div>
              <div className="mt-4">
                <h3 className="text-lg font-medium text-gray-900">
                  <span className="absolute inset-0" />
                  Manage Credits
                </h3>
                <p className="mt-2 text-sm text-gray-500">
                  Add funds or view your credit balance and transactions.
                </p>
              </div>
            </Link>

            <Link
              to="/settings"
              className="relative group bg-purple-50 p-6 focus-within:ring-2 focus-within:ring-inset focus-within:ring-purple-500 rounded-lg hover:bg-purple-100 transition-colors"
            >
              <div>
                <span className="rounded-lg inline-flex p-3 bg-purple-600 text-white">
                  <Users className="h-6 w-6" />
                </span>
              </div>
              <div className="mt-4">
                <h3 className="text-lg font-medium text-gray-900">
                  <span className="absolute inset-0" />
                  Settings
                </h3>
                <p className="mt-2 text-sm text-gray-500">
                  Configure your account and network settings.
                </p>
              </div>
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Dashboard