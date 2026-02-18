import axios from 'axios'
import toast from 'react-hot-toast'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

// Create base axios instance
export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('deparrow_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('deparrow_token')
      localStorage.removeItem('deparrow_user')
      window.location.href = '/login'
    }
    
    const message = error.response?.data?.message || error.message || 'An error occurred'
    // Don't show toast for 401 errors (handled by redirect)
    if (error.response?.status !== 401) {
      toast.error(message)
    }
    
    return Promise.reject(error)
  }
)

// ==================== Types ====================

export interface User {
  id: string
  email: string
  name: string
  role: 'user' | 'admin' | 'node_operator'
  credits: number
  created_at: string
  updated_at: string
}

export interface Job {
  id: string
  name: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  progress?: number
  credit_cost: number
  created_at: string
  started_at?: string
  completed_at?: string
  spec: JobSpec
  logs?: string[]
  results?: Record<string, any>
  node_id?: string
  error?: string
}

export interface JobSpec {
  engine: 'docker' | 'wasm' | 'native'
  image?: string
  command?: string
  wasm_file?: string
  function?: string
  input?: any
  resources?: {
    cpu: number
    memory: string
    gpu?: number
  }
  timeout?: number
  priority?: 'low' | 'normal' | 'high'
}

export interface Node {
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

export interface Provider {
  id: string
  name: string
  status: 'online' | 'offline' | 'maintenance' | 'busy'
  location: {
    region: string
    country: string
    city: string
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
  joined_at: string
}

export interface Transaction {
  id: string
  type: 'earn' | 'spend' | 'transfer_in' | 'transfer_out'
  amount: number
  description: string
  timestamp: string
  status: 'completed' | 'pending' | 'failed'
  job_id?: string
  counterparty?: string
}

export interface CreditBalance {
  balance: number
  total_earned: number
  total_spent: number
  pending_transactions: number
}

export interface AgentStatus {
  id: string
  name: string
  status: 'running' | 'paused' | 'stopped' | 'error'
  uptime: string
  credits_earned: number
  credits_spent: number
  tasks_completed: number
  current_task?: string
  tools: AgentTool[]
  resources: {
    cpu_usage: number
    memory_usage: number
    disk_usage: number
  }
  last_heartbeat: string
}

export interface AgentTool {
  name: string
  description: string
  enabled: boolean
  calls: number
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
  tool_calls?: {
    tool: string
    args: Record<string, any>
    result: any
    status: 'pending' | 'success' | 'error'
  }[]
}

export interface SystemStats {
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
    completed: number
    failed: number
  }
}

export interface SystemHealth {
  status: string
  timestamp: string
  version: string
  components: Record<string, number>
}

// ==================== Auth API ====================

export const authAPI = {
  login: (email: string, password: string) =>
    api.post<{ token: string; user: User }>('/auth/login', { email, password }),
  
  register: (data: { email: string; password: string; name: string }) =>
    api.post<{ token: string; user: User }>('/auth/register', data),
  
  logout: () => 
    api.post('/auth/logout'),
  
  me: () => 
    api.get<User>('/auth/me'),
  
  refreshToken: () =>
    api.post<{ token: string }>('/auth/refresh'),
  
  updatePassword: (currentPassword: string, newPassword: string) =>
    api.put('/auth/password', { currentPassword, newPassword }),
}

// ==================== Jobs API ====================

export const jobsAPI = {
  list: (params?: { 
    page?: number
    limit?: number
    status?: string
    sortBy?: string
    sortOrder?: 'asc' | 'desc'
  }) => 
    api.get<{ jobs: Job[]; total: number; page: number; limit: number }>('/jobs', { params }),
  
  get: (id: string) => 
    api.get<Job>(`/jobs/${id}`),
  
  create: (data: {
    name: string
    spec: JobSpec
    priority?: 'low' | 'normal' | 'high'
    max_cost?: number
  }) => 
    api.post<Job>('/jobs', data),
  
  cancel: (id: string) => 
    api.post(`/jobs/${id}/cancel`),
  
  logs: (id: string, params?: { follow?: boolean; tail?: number }) => 
    api.get<{ logs: string[] }>(`/jobs/${id}/logs`, { params }),
  
  results: (id: string) => 
    api.get<{ results: Record<string, any>; outputs: Record<string, string> }>(`/jobs/${id}/results`),
  
  retry: (id: string) => 
    api.post<Job>(`/jobs/${id}/retry`),
  
  estimate: (data: JobSpec) => 
    api.post<{ estimated_cost: number; estimated_time: number }>('/jobs/estimate', data),
}

// ==================== Nodes API ====================

export const nodesAPI = {
  list: (params?: {
    status?: string
    arch?: string
    region?: string
    gpu?: boolean
  }) => 
    api.get<{ nodes: Node[] }>('/nodes', { params }),
  
  get: (id: string) => 
    api.get<Node>(`/nodes/${id}`),
  
  stats: () => 
    api.get<{
      total: number
      online: number
      offline: number
      total_cpu: number
      total_memory: string
      total_gpu: number
    }>('/nodes/stats'),
  
  register: (data: {
    name?: string
    labels?: Record<string, string>
    resources?: {
      cpu: number
      memory: string
      disk: string
      gpu?: number
    }
  }) => 
    api.post<Node>('/nodes/register', data),
  
  update: (id: string, data: {
    labels?: Record<string, string>
    status?: string
  }) => 
    api.put<Node>(`/nodes/${id}`, data),
  
  deregister: (id: string) => 
    api.delete(`/nodes/${id}`),
}

// ==================== Wallet API ====================

export const walletAPI = {
  balance: () => 
    api.get<CreditBalance>('/wallet/balance'),
  
  transactions: (params?: { 
    page?: number
    limit?: number
    type?: string
    startDate?: string
    endDate?: string
  }) => 
    api.get<{ transactions: Transaction[]; total: number }>('/wallet/transactions', { params }),
  
  deposit: (amount: number, paymentMethod?: string) => 
    api.post<{ transaction: Transaction; payment_url?: string }>('/wallet/deposit', { amount, paymentMethod }),
  
  withdraw: (amount: number, address?: string) => 
    api.post<{ transaction: Transaction }>('/wallet/withdraw', { amount, address }),
  
  transfer: (recipientId: string, amount: number, memo?: string) => 
    api.post<{ transaction: Transaction }>('/wallet/transfer', { recipientId, amount, memo }),
  
  history: (params?: { period?: 'day' | 'week' | 'month' | 'year' }) => 
    api.get<{
      earned: number
      spent: number
      transactions: Transaction[]
    }>('/wallet/history', { params }),
}

// ==================== Agent API ====================

export const agentAPI = {
  status: () => 
    api.get<AgentStatus>('/agent/status'),
  
  start: () => 
    api.post<AgentStatus>('/agent/start'),
  
  pause: () => 
    api.post<AgentStatus>('/agent/pause'),
  
  stop: () => 
    api.post<AgentStatus>('/agent/stop'),
  
  chat: (message: string, context?: Record<string, any>) => 
    api.post<{ message: ChatMessage }>('/agent/chat', { message, context }),
  
  history: (params?: { limit?: number; before?: string }) => 
    api.get<{ messages: ChatMessage[] }>('/agent/history', { params }),
  
  tools: () => 
    api.get<{ tools: AgentTool[] }>('/agent/tools'),
  
  updateTool: (toolName: string, enabled: boolean) => 
    api.put<{ tool: AgentTool }>(`/agent/tools/${toolName}`, { enabled }),
  
  configure: (config: {
    auto_accept_jobs?: boolean
    max_cost_per_job?: number
    priority_filter?: string[]
    resource_limits?: {
      max_cpu: number
      max_memory: string
      max_gpu: number
    }
  }) => 
    api.put<AgentStatus>('/agent/configure', config),
}

// ==================== Providers API ====================

export const providersAPI = {
  list: (params?: {
    status?: string
    region?: string
    gpu?: boolean
    minReputation?: number
    sortBy?: 'reputation' | 'price' | 'uptime'
    sortOrder?: 'asc' | 'desc'
  }) => 
    api.get<{ providers: Provider[] }>('/providers', { params }),
  
  get: (id: string) => 
    api.get<Provider>(`/providers/${id}`),
  
  estimate: (providerId: string, spec: {
    cpuHours: number
    memoryGb: number
    gpuHours: number
    storageGb: number
  }) => 
    api.post<{ estimated_cost: number }>('/providers/estimate', { providerId, spec }),
  
  submitJob: (providerId: string, job: {
    name: string
    spec: JobSpec
    priority?: 'low' | 'normal' | 'high'
    max_cost?: number
  }) => 
    api.post<Job>(`/providers/${providerId}/jobs`, job),
}

// ==================== System API ====================

export const systemAPI = {
  health: () => 
    api.get<SystemHealth>('/health'),
  
  stats: () => 
    api.get<{ metrics: SystemStats }>('/stats'),
  
  config: () => 
    api.get<{
      version: string
      network: string
      features: string[]
      limits: Record<string, number>
    }>('/config'),
  
  metrics: () => 
    api.get<{
      jobs_per_hour: number
      avg_job_duration: number
      network_throughput: number
      active_users: number
    }>('/metrics'),
  
  events: (params?: { 
    type?: string
    limit?: number
    since?: string
  }) => 
    api.get<{
      events: Array<{
        id: string
        type: string
        timestamp: string
        data: Record<string, any>
      }>
    }>('/events', { params }),
}

// ==================== WebSocket Connection ====================

export class DeparrowWebSocket {
  private ws: WebSocket | null = null
  private url: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private listeners: Map<string, Set<(data: any) => void>> = new Map()

  constructor(baseUrl?: string) {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = baseUrl || (API_BASE_URL.replace(/^https?:\/\//, '').replace(/\/api$/, ''))
    this.url = `${wsProtocol}//${host}/ws`
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url)

        this.ws.onopen = () => {
          console.log('WebSocket connected')
          this.reconnectAttempts = 0
          
          // Authenticate
          const token = localStorage.getItem('deparrow_token')
          if (token) {
            this.send('auth', { token })
          }
          
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            this.handleMessage(message)
          } catch (error) {
            console.error('WebSocket message parse error:', error)
          }
        }

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error)
          reject(error)
        }

        this.ws.onclose = () => {
          console.log('WebSocket disconnected')
          this.attemptReconnect()
        }
      } catch (error) {
        reject(error)
      }
    })
  }

  private handleMessage(message: { type: string; data: any }) {
    const listeners = this.listeners.get(message.type)
    if (listeners) {
      listeners.forEach(callback => callback(message.data))
    }
    
    // Also emit to 'all' listeners
    const allListeners = this.listeners.get('*')
    if (allListeners) {
      allListeners.forEach(callback => callback(message))
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max WebSocket reconnection attempts reached')
      return
    }

    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
    
    console.log(`Attempting to reconnect in ${delay}ms...`)
    
    setTimeout(() => {
      this.connect().catch(console.error)
    }, delay)
  }

  send(type: string, data: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }))
    }
  }

  subscribe(type: string, callback: (data: any) => void): () => void {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set())
    }
    
    this.listeners.get(type)!.add(callback)
    
    // Return unsubscribe function
    return () => {
      this.listeners.get(type)?.delete(callback)
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.listeners.clear()
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

// Create singleton instance
let wsInstance: DeparrowWebSocket | null = null

export const getWebSocket = (): DeparrowWebSocket => {
  if (!wsInstance) {
    wsInstance = new DeparrowWebSocket()
  }
  return wsInstance
}

export default api
