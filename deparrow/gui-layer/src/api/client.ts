import axios from 'axios'
import toast from 'react-hot-toast'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

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
    toast.error(message)
    
    return Promise.reject(error)
  }
)

// Auth API
export const authAPI = {
  login: (email: string, password: string) =>
    api.post('/auth/login', { email, password }),
  register: (data: { email: string; password: string; name: string }) =>
    api.post('/auth/register', data),
  logout: () => api.post('/auth/logout'),
  me: () => api.get('/auth/me'),
}

// Jobs API
export const jobsAPI = {
  list: (params?: { page?: number; limit?: number; status?: string }) =>
    api.get('/jobs', { params }),
  get: (id: string) => api.get(`/jobs/${id}`),
  create: (data: any) => api.post('/jobs', data),
  cancel: (id: string) => api.post(`/jobs/${id}/cancel`),
  logs: (id: string) => api.get(`/jobs/${id}/logs`),
  results: (id: string) => api.get(`/jobs/${id}/results`),
  estimate: (spec: any) => api.post('/jobs/estimate', spec),
}

// Nodes API
export const nodesAPI = {
  list: (params?: { status?: string; arch?: string; region?: string }) =>
    api.get('/nodes', { params }),
  get: (id: string) => api.get(`/nodes/${id}`),
  stats: () => api.get('/nodes/stats'),
  update: (id: string, data: { labels?: Record<string, string>; status?: string }) =>
    api.put(`/nodes/${id}`, data),
}

// Wallet API
export const walletAPI = {
  balance: () => api.get('/wallet/balance'),
  transactions: (params?: { page?: number; limit?: number }) =>
    api.get('/wallet/transactions', { params }),
  deposit: (amount: number) => api.post('/wallet/deposit', { amount }),
  withdraw: (amount: number) => api.post('/wallet/withdraw', { amount }),
  transfer: (recipientId: string, amount: number) =>
    api.post('/wallet/transfer', { recipientId, amount }),
  history: (params?: { period?: string }) =>
    api.get('/wallet/history', { params }),
}

// System API
export const systemAPI = {
  health: () => api.get('/health'),
  stats: () => api.get('/stats'),
  config: () => api.get('/config'),
  metrics: () => api.get('/metrics'),
}

// Agent API (for Agent.tsx)
export const agentAPI = {
  status: () => api.get('/agent/status'),
  start: () => api.post('/agent/start'),
  pause: () => api.post('/agent/pause'),
  stop: () => api.post('/agent/stop'),
  chat: (message: string) => api.post('/agent/chat', { message }),
  updateTool: (toolName: string, enabled: boolean) =>
    api.put(`/agent/tools/${toolName}`, { enabled }),
}

// Providers API (for Providers.tsx)
export const providersAPI = {
  list: (params?: { status?: string; region?: string; gpu?: boolean }) =>
    api.get('/providers', { params }),
  get: (id: string) => api.get(`/providers/${id}`),
  estimate: (providerId: string, spec: any) =>
    api.post('/providers/estimate', { providerId, spec }),
  submitJob: (providerId: string, job: any) =>
    api.post(`/providers/${providerId}/jobs`, job),
}

export default api
