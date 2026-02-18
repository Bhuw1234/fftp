import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { 
  Plus, 
  Pause, 
  Square, 
  Eye,
  Search,
  Calendar,
  Clock,
  CreditCard,
  AlertCircle,
  CheckCircle,
  XCircle,
  Loader
} from 'lucide-react'
import { jobsAPI } from '../api/client'
import toast from 'react-hot-toast'
import { format } from 'date-fns'

interface Job {
  id: string
  name: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  progress?: number
  credit_cost: number
  created_at: string
  started_at?: string
  completed_at?: string
  spec: any
  logs?: string[]
  results?: any
}

const Jobs: React.FC = () => {
  const [jobs, setJobs] = useState<Job[]>([])
  const [filteredJobs, setFilteredJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [selectedJob, setSelectedJob] = useState<Job | null>(null)
  const [showJobModal, setShowJobModal] = useState(false)

  useEffect(() => {
    loadJobs()
  }, [])

  useEffect(() => {
    filterJobs()
  }, [jobs, searchTerm, statusFilter])

  const loadJobs = async () => {
    try {
      setLoading(true)
      const response = await jobsAPI.list()
      setJobs(response.data.jobs || [])
    } catch (error: any) {
      console.error('Jobs loading error:', error)
      // Load mock data for demonstration
      loadMockJobs()
      toast.error('Using demo data - connect to backend for live data')
    } finally {
      setLoading(false)
    }
  }

  const loadMockJobs = () => {
    const mockJobs: Job[] = [
      {
        id: 'job-1',
        name: 'Data Processing Pipeline',
        status: 'running',
        progress: 65,
        credit_cost: 15.5,
        created_at: '2024-01-15T10:30:00Z',
        spec: {
          engine: 'docker',
          image: 'python:3.9-slim',
          command: 'python process_data.py',
          resources: { cpu: 2, memory: '4Gi' }
        }
      },
      {
        id: 'job-2',
        name: 'ML Model Training',
        status: 'completed',
        progress: 100,
        credit_cost: 25.0,
        created_at: '2024-01-15T09:15:00Z',
        started_at: '2024-01-15T09:16:00Z',
        completed_at: '2024-01-15T11:45:00Z',
        spec: {
          engine: 'docker',
          image: 'tensorflow/tensorflow:latest',
          command: 'python train_model.py',
          resources: { cpu: 4, memory: '8Gi', gpu: 1 }
        },
        results: {
          accuracy: 0.94,
          training_time: '2h 29m',
          model_size: '45MB'
        }
      },
      {
        id: 'job-3',
        name: 'Image Processing',
        status: 'failed',
        progress: 30,
        credit_cost: 8.0,
        created_at: '2024-01-15T08:45:00Z',
        started_at: '2024-01-15T08:46:00Z',
        spec: {
          engine: 'docker',
          image: 'opencv:latest',
          command: 'python process_images.py',
          resources: { cpu: 2, memory: '2Gi' }
        }
      },
      {
        id: 'job-4',
        name: 'Batch Data Analysis',
        status: 'pending',
        progress: 0,
        credit_cost: 12.5,
        created_at: '2024-01-15T14:20:00Z',
        spec: {
          engine: 'docker',
          image: 'python:3.9-slim',
          command: 'python analyze_batch.py',
          resources: { cpu: 1, memory: '2Gi' }
        }
      },
      {
        id: 'job-5',
        name: 'WebAssembly Computing',
        status: 'completed',
        progress: 100,
        credit_cost: 5.0,
        created_at: '2024-01-15T07:30:00Z',
        started_at: '2024-01-15T07:31:00Z',
        completed_at: '2024-01-15T08:15:00Z',
        spec: {
          engine: 'wasm',
          wasm_file: 'compute.wasm',
          function: 'fibonacci',
          input: 40
        },
        results: {
          result: 102334155,
          execution_time: '44s',
          memory_usage: '16MB'
        }
      }
    ]
    setJobs(mockJobs)
  }

  const filterJobs = () => {
    let filtered = jobs

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(job =>
        job.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        job.id.toLowerCase().includes(searchTerm.toLowerCase())
      )
    }

    // Status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(job => job.status === statusFilter)
    }

    setFilteredJobs(filtered)
  }

  const handleJobAction = async (action: string, jobId: string) => {
    try {
      switch (action) {
        case 'cancel':
          await jobsAPI.cancel(jobId)
          toast.success('Job cancelled successfully')
          loadJobs()
          break
        case 'view':
          const job = jobs.find(j => j.id === jobId)
          if (job) {
            setSelectedJob(job)
            setShowJobModal(true)
          }
          break
        default:
          console.info(`${action} action not implemented yet`)
      }
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Action failed')
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-4 w-4 text-yellow-500" />
      case 'running':
        return <Loader className="h-4 w-4 text-blue-500 animate-spin" />
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />
      case 'cancelled':
        return <Square className="h-4 w-4 text-gray-500" />
      default:
        return <AlertCircle className="h-4 w-4 text-gray-400" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-100 text-yellow-800'
      case 'running': return 'bg-blue-100 text-blue-800'
      case 'completed': return 'bg-green-100 text-green-800'
      case 'failed': return 'bg-red-100 text-red-800'
      case 'cancelled': return 'bg-gray-100 text-gray-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const formatDate = (dateString: string) => {
    return format(new Date(dateString), 'MMM dd, yyyy HH:mm')
  }

  const getEngineIcon = (engine: string) => {
    switch (engine) {
      case 'docker': return '🐳'
      case 'wasm': return '🧩'
      case 'native': return '⚡'
      default: return '🔧'
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
      <div className="sm:flex sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Jobs</h1>
          <p className="text-gray-600">
            Manage and monitor your compute jobs across the DEparrow network.
          </p>
        </div>
        <div className="mt-4 sm:mt-0">
          <Link
            to="/jobs/new"
            className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            <Plus className="mr-2 h-4 w-4" />
            New Job
          </Link>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white shadow rounded-lg p-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Search */}
          <div>
            <label htmlFor="search" className="block text-sm font-medium text-gray-700 mb-1">
              Search Jobs
            </label>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                type="text"
                id="search"
                placeholder="Search by name or ID..."
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
              <option value="pending">Pending</option>
              <option value="running">Running</option>
              <option value="completed">Completed</option>
              <option value="failed">Failed</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>

          {/* Quick Stats */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Summary
            </label>
            <div className="text-sm text-gray-600">
              {filteredJobs.length} of {jobs.length} jobs
            </div>
            <div className="text-xs text-gray-500">
              Total cost: {filteredJobs.reduce((sum, job) => sum + job.credit_cost, 0).toFixed(2)} credits
            </div>
          </div>
        </div>
      </div>

      {/* Jobs List */}
      <div className="bg-white shadow rounded-lg overflow-hidden">
        {filteredJobs.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Job
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Engine
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Cost
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {filteredJobs.map((job) => (
                  <tr key={job.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <div className="text-sm font-medium text-gray-900">{job.name}</div>
                        <div className="text-sm text-gray-500">{job.id}</div>
                        {job.status === 'running' && (
                          <div className="mt-1">
                            <div className="bg-gray-200 rounded-full h-2">
                              <div 
                                className="bg-blue-600 h-2 rounded-full transition-all duration-300" 
                                style={{ width: `${job.progress || 0}%` }}
                              ></div>
                            </div>
                            <div className="text-xs text-gray-500 mt-1">
                              {job.progress}% complete
                            </div>
                          </div>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        {getStatusIcon(job.status)}
                        <span className={`ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(job.status)}`}>
                          {job.status}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <div className="flex items-center">
                        <span className="mr-2 text-lg">{getEngineIcon(job.spec.engine)}</span>
                        <span className="capitalize">{job.spec.engine}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      <div className="flex items-center">
                        <CreditCard className="h-4 w-4 text-gray-400 mr-1" />
                        {job.credit_cost} credits
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      <div className="flex items-center">
                        <Calendar className="h-4 w-4 mr-1" />
                        {formatDate(job.created_at)}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => handleJobAction('view', job.id)}
                          className="text-blue-600 hover:text-blue-900 p-1 rounded"
                          title="View Details"
                        >
                          <Eye className="h-4 w-4" />
                        </button>
                        {job.status === 'running' && (
                          <button
                            onClick={() => handleJobAction('pause', job.id)}
                            className="text-yellow-600 hover:text-yellow-900 p-1 rounded"
                            title="Pause"
                          >
                            <Pause className="h-4 w-4" />
                          </button>
                        )}
                        {(job.status === 'running' || job.status === 'pending') && (
                          <button
                            onClick={() => handleJobAction('cancel', job.id)}
                            className="text-red-600 hover:text-red-900 p-1 rounded"
                            title="Cancel"
                          >
                            <Square className="h-4 w-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="text-center py-12">
            <div className="mx-auto h-12 w-12 text-gray-400">
              <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No jobs found</h3>
            <p className="mt-1 text-sm text-gray-500">
              {searchTerm || statusFilter !== 'all' 
                ? 'Try adjusting your filters or search terms.'
                : 'Get started by creating your first job.'
              }
            </p>
            {!searchTerm && statusFilter === 'all' && (
              <div className="mt-6">
                <Link
                  to="/jobs/new"
                  className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Create Job
                </Link>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Job Detail Modal */}
      {showJobModal && selectedJob && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" onClick={() => setShowJobModal(false)}></div>
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <div className="sm:flex sm:items-start">
                  <div className="w-full">
                    <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                      Job Details: {selectedJob.name}
                    </h3>
                    
                    <div className="space-y-4">
                      <div>
                        <h4 className="text-sm font-medium text-gray-900">Basic Information</h4>
                        <dl className="mt-2 grid grid-cols-1 gap-x-4 gap-y-2 sm:grid-cols-2">
                          <div>
                            <dt className="text-sm text-gray-500">Job ID</dt>
                            <dd className="text-sm text-gray-900">{selectedJob.id}</dd>
                          </div>
                          <div>
                            <dt className="text-sm text-gray-500">Status</dt>
                            <dd className="text-sm text-gray-900 capitalize">{selectedJob.status}</dd>
                          </div>
                          <div>
                            <dt className="text-sm text-gray-500">Engine</dt>
                            <dd className="text-sm text-gray-900 capitalize">{selectedJob.spec.engine}</dd>
                          </div>
                          <div>
                            <dt className="text-sm text-gray-500">Cost</dt>
                            <dd className="text-sm text-gray-900">{selectedJob.credit_cost} credits</dd>
                          </div>
                        </dl>
                      </div>

                      <div>
                        <h4 className="text-sm font-medium text-gray-900">Specifications</h4>
                        <pre className="mt-2 text-xs bg-gray-100 p-2 rounded overflow-auto">
                          {JSON.stringify(selectedJob.spec, null, 2)}
                        </pre>
                      </div>

                      {selectedJob.results && (
                        <div>
                          <h4 className="text-sm font-medium text-gray-900">Results</h4>
                          <pre className="mt-2 text-xs bg-green-50 p-2 rounded overflow-auto">
                            {JSON.stringify(selectedJob.results, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={() => setShowJobModal(false)}
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

export default Jobs