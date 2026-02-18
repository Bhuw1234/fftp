import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from 'react-query'
import {
  jobsAPI,
  nodesAPI,
  walletAPI,
  agentAPI,
  providersAPI,
  systemAPI,
  getWebSocket,
  Job,
  Node,
  Provider,
  Transaction,
  CreditBalance,
  AgentStatus,
  SystemStats,
  SystemHealth,
  JobSpec,
  DeparrowWebSocket,
} from '../api/deparrow'
import { useAuth } from '../contexts/AuthContext'
import toast from 'react-hot-toast'

// ==================== useCredits ====================

export function useCredits() {
  const { user } = useAuth()
  const queryClient = useQueryClient()

  const { data: balance, isLoading, error, refetch } = useQuery<CreditBalance>(
    ['credits', 'balance'],
    () => walletAPI.balance().then(res => res.data),
    {
      enabled: !!user,
      refetchInterval: 30000, // Refresh every 30 seconds
      staleTime: 10000,
    }
  )

  const depositMutation = useMutation(
    (amount: number) => walletAPI.deposit(amount).then(res => res.data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['credits', 'balance'])
        toast.success('Deposit initiated')
      },
      onError: (error: any) => {
        toast.error(error.response?.data?.message || 'Deposit failed')
      },
    }
  )

  const withdrawMutation = useMutation(
    (amount: number) => walletAPI.withdraw(amount).then(res => res.data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['credits', 'balance'])
        toast.success('Withdrawal initiated')
      },
      onError: (error: any) => {
        toast.error(error.response?.data?.message || 'Withdrawal failed')
      },
    }
  )

  const transferMutation = useMutation(
    ({ recipientId, amount }: { recipientId: string; amount: number }) =>
      walletAPI.transfer(recipientId, amount).then(res => res.data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['credits', 'balance'])
        toast.success('Transfer completed')
      },
      onError: (error: any) => {
        toast.error(error.response?.data?.message || 'Transfer failed')
      },
    }
  )

  return {
    balance: balance?.balance || 0,
    totalEarned: balance?.total_earned || 0,
    totalSpent: balance?.total_spent || 0,
    pendingTransactions: balance?.pending_transactions || 0,
    isLoading,
    error,
    refetch,
    deposit: depositMutation.mutate,
    withdraw: withdrawMutation.mutate,
    transfer: transferMutation.mutate,
    isDepositing: depositMutation.isLoading,
    isWithdrawing: withdrawMutation.isLoading,
    isTransferring: transferMutation.isLoading,
  }
}

// ==================== useNodes ====================

export function useNodes(filters?: {
  status?: string
  arch?: string
  region?: string
  gpu?: boolean
}) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [wsConnected, setWsConnected] = useState(false)

  const { data, isLoading, error, refetch } = useQuery<{ nodes: Node[] }>(
    ['nodes', filters],
    () => nodesAPI.list(filters).then(res => res.data),
    {
      enabled: !!user,
      refetchInterval: 60000, // Refresh every minute
      staleTime: 30000,
    }
  )

  // WebSocket subscription for real-time updates
  useEffect(() => {
    if (!user) return

    const ws = getWebSocket()
    
    const unsubscribe = ws.subscribe('node_update', (data) => {
      queryClient.setQueryData<Node[]>(['nodes', filters], (old = []) => {
        const index = old.findIndex(n => n.id === data.node.id)
        if (index >= 0) {
          const updated = [...old]
          updated[index] = { ...updated[index], ...data.node }
          return updated
        }
        return [...old, data.node]
      })
    })

    setWsConnected(ws.isConnected())

    return () => {
      unsubscribe()
    }
  }, [user, filters, queryClient])

  const statsQuery = useQuery(
    ['nodes', 'stats'],
    () => nodesAPI.stats().then(res => res.data),
    {
      enabled: !!user,
      staleTime: 30000,
    }
  )

  const updateNodeMutation = useMutation(
    ({ id, data }: { id: string; data: { labels?: Record<string, string>; status?: string } }) =>
      nodesAPI.update(id, data).then(res => res.data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['nodes'])
        toast.success('Node updated')
      },
    }
  )

  return {
    nodes: data?.nodes || [],
    stats: statsQuery.data,
    isLoading,
    error,
    refetch,
    updateNode: updateNodeMutation.mutate,
    wsConnected,
  }
}

// ==================== useJobs ====================

export function useJobs(filters?: {
  status?: string
  page?: number
  limit?: number
  sortBy?: string
}) {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [wsConnected, setWsConnected] = useState(false)

  const { data, isLoading, error, refetch } = useQuery<{ jobs: Job[]; total: number }>(
    ['jobs', filters],
    () => jobsAPI.list(filters).then(res => res.data),
    {
      enabled: !!user,
      refetchInterval: 15000, // Refresh every 15 seconds
      staleTime: 5000,
    }
  )

  // WebSocket subscription for real-time job updates
  useEffect(() => {
    if (!user) return

    const ws = getWebSocket()

    const unsubscribers = [
      ws.subscribe('job_update', (data) => {
        queryClient.invalidateQueries(['jobs', filters])
        // Also update the specific job if we have a detail view
        queryClient.setQueryData(['job', data.job.id], data.job)
      }),
      ws.subscribe('job_created', () => {
        queryClient.invalidateQueries(['jobs', filters])
      }),
    ]

    setWsConnected(ws.isConnected())

    return () => {
      unsubscribers.forEach(unsub => unsub())
    }
  }, [user, filters, queryClient])

  const createJobMutation = useMutation(
    (job: { name: string; spec: JobSpec; priority?: 'low' | 'normal' | 'high'; max_cost?: number }) =>
      jobsAPI.create(job).then(res => res.data),
    {
      onSuccess: (job) => {
        queryClient.invalidateQueries(['jobs'])
        toast.success(`Job "${job.name}" created`)
      },
      onError: (error: any) => {
        toast.error(error.response?.data?.message || 'Failed to create job')
      },
    }
  )

  const cancelJobMutation = useMutation(
    (id: string) => jobsAPI.cancel(id).then(res => res.data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['jobs'])
        toast.success('Job cancelled')
      },
    }
  )

  const estimateMutation = useMutation(
    (spec: JobSpec) => jobsAPI.estimate(spec).then(res => res.data),
    {
      onError: (error: any) => {
        toast.error(error.response?.data?.message || 'Failed to estimate job cost')
      },
    }
  )

  return {
    jobs: data?.jobs || [],
    total: data?.total || 0,
    isLoading,
    error,
    refetch,
    createJob: createJobMutation.mutate,
    cancelJob: cancelJobMutation.mutate,
    estimateJob: estimateMutation.mutate,
    estimateResult: estimateMutation.data,
    isCreating: createJobMutation.isLoading,
    isEstimating: estimateMutation.isLoading,
    wsConnected,
  }
}

// ==================== useJob ====================

export function useJob(id: string) {
  const { user } = useAuth()
  const queryClient = useQueryClient()

  const { data: job, isLoading, error, refetch } = useQuery<Job>(
    ['job', id],
    () => jobsAPI.get(id).then(res => res.data),
    {
      enabled: !!user && !!id,
      refetchInterval: 10000, // Refresh every 10 seconds
      staleTime: 5000,
    }
  )

  const { data: logs } = useQuery<{ logs: string[] }>(
    ['job', id, 'logs'],
    () => jobsAPI.logs(id).then(res => res.data),
    {
      enabled: !!user && !!id && job?.status === 'running',
      refetchInterval: 5000,
    }
  )

  const { data: results } = useQuery(
    ['job', id, 'results'],
    () => jobsAPI.results(id).then(res => res.data),
    {
      enabled: !!user && !!id && job?.status === 'completed',
    }
  )

  // WebSocket subscription for real-time updates
  useEffect(() => {
    if (!user || !id) return

    const ws = getWebSocket()

    const unsubscribe = ws.subscribe('job_update', (data) => {
      if (data.job.id === id) {
        queryClient.setQueryData(['job', id], (old: Job | undefined) => {
          return old ? { ...old, ...data.job } : data.job
        })
      }
    })

    return () => {
      unsubscribe()
    }
  }, [user, id, queryClient])

  return {
    job,
    logs: logs?.logs || [],
    results,
    isLoading,
    error,
    refetch,
  }
}

// ==================== useAgent ====================

export function useAgent() {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const [wsConnected, setWsConnected] = useState(false)

  const { data: status, isLoading, error, refetch } = useQuery<AgentStatus>(
    ['agent', 'status'],
    () => agentAPI.status().then(res => res.data),
    {
      enabled: !!user,
      refetchInterval: 10000,
      staleTime: 5000,
    }
  )

  // WebSocket subscription for agent updates
  useEffect(() => {
    if (!user) return

    const ws = getWebSocket()

    const unsubscribe = ws.subscribe('agent_update', (data) => {
      queryClient.setQueryData(['agent', 'status'], (old: AgentStatus | undefined) => {
        return old ? { ...old, ...data.agent } : data.agent
      })
    })

    setWsConnected(ws.isConnected())

    return () => {
      unsubscribe()
    }
  }, [user, queryClient])

  const startMutation = useMutation(
    () => agentAPI.start().then(res => res.data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['agent', 'status'])
        toast.success('Agent started')
      },
    }
  )

  const pauseMutation = useMutation(
    () => agentAPI.pause().then(res => res.data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['agent', 'status'])
        toast.success('Agent paused')
      },
    }
  )

  const stopMutation = useMutation(
    () => agentAPI.stop().then(res => res.data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['agent', 'status'])
        toast.success('Agent stopped')
      },
    }
  )

  const chatMutation = useMutation(
    (message: string) => agentAPI.chat(message).then(res => res.data),
    {
      onError: (error: any) => {
        toast.error(error.response?.data?.message || 'Chat failed')
      },
    }
  )

  const updateToolMutation = useMutation(
    ({ toolName, enabled }: { toolName: string; enabled: boolean }) =>
      agentAPI.updateTool(toolName, enabled).then(res => res.data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['agent', 'status'])
        toast.success('Tool updated')
      },
    }
  )

  return {
    status,
    isLoading,
    error,
    refetch,
    start: startMutation.mutate,
    pause: pauseMutation.mutate,
    stop: stopMutation.mutate,
    chat: chatMutation.mutate,
    updateTool: updateToolMutation.mutate,
    chatResponse: chatMutation.data,
    isStarting: startMutation.isLoading,
    isPausing: pauseMutation.isLoading,
    isStopping: stopMutation.isLoading,
    isChatting: chatMutation.isLoading,
    wsConnected,
  }
}

// ==================== useProviders ====================

export function useProviders(filters?: {
  status?: string
  region?: string
  gpu?: boolean
  minReputation?: number
  sortBy?: 'reputation' | 'price' | 'uptime'
}) {
  const { user } = useAuth()

  const { data, isLoading, error, refetch } = useQuery<{ providers: Provider[] }>(
    ['providers', filters],
    () => providersAPI.list(filters).then(res => res.data),
    {
      enabled: !!user,
      staleTime: 60000, // 1 minute
    }
  )

  const estimateMutation = useMutation(
    ({ providerId, spec }: { 
      providerId: string
      spec: {
        cpuHours: number
        memoryGb: number
        gpuHours: number
        storageGb: number
      }
    }) => providersAPI.estimate(providerId, spec).then(res => res.data),
    {
      onError: (error: any) => {
        toast.error(error.response?.data?.message || 'Failed to estimate cost')
      },
    }
  )

  const submitJobMutation = useMutation(
    ({ providerId, job }: { 
      providerId: string
      job: {
        name: string
        spec: JobSpec
        priority?: 'low' | 'normal' | 'high'
        max_cost?: number
      }
    }) => providersAPI.submitJob(providerId, job).then(res => res.data),
    {
      onSuccess: (job) => {
        toast.success(`Job "${job.name}" submitted to provider`)
      },
      onError: (error: any) => {
        toast.error(error.response?.data?.message || 'Failed to submit job')
      },
    }
  )

  return {
    providers: data?.providers || [],
    isLoading,
    error,
    refetch,
    estimateCost: estimateMutation.mutate,
    estimateResult: estimateMutation.data,
    submitJob: submitJobMutation.mutate,
    isEstimating: estimateMutation.isLoading,
    isSubmitting: submitJobMutation.isLoading,
  }
}

// ==================== useTransactions ====================

export function useTransactions(filters?: {
  type?: string
  page?: number
  limit?: number
  startDate?: string
  endDate?: string
}) {
  const { user } = useAuth()
  const queryClient = useQueryClient()

  const { data, isLoading, error, refetch } = useQuery<{ transactions: Transaction[]; total: number }>(
    ['transactions', filters],
    () => walletAPI.transactions(filters).then(res => res.data),
    {
      enabled: !!user,
      staleTime: 30000,
    }
  )

  const { data: history } = useQuery(
    ['wallet', 'history'],
    () => walletAPI.history().then(res => res.data),
    {
      enabled: !!user,
      staleTime: 60000,
    }
  )

  // WebSocket subscription for new transactions
  useEffect(() => {
    if (!user) return

    const ws = getWebSocket()

    const unsubscribe = ws.subscribe('transaction', (data) => {
      queryClient.invalidateQueries(['transactions'])
      queryClient.invalidateQueries(['credits', 'balance'])
      
      if (data.transaction.type === 'earn') {
        toast.success(`Earned ${data.transaction.amount} credits!`)
      }
    })

    return () => {
      unsubscribe()
    }
  }, [user, queryClient])

  return {
    transactions: data?.transactions || [],
    total: data?.total || 0,
    history,
    isLoading,
    error,
    refetch,
  }
}

// ==================== useSystemStats ====================

export function useSystemStats() {
  const { user } = useAuth()

  const { data: stats, isLoading, error, refetch } = useQuery<SystemStats>(
    ['system', 'stats'],
    () => systemAPI.stats().then(res => res.data.metrics),
    {
      enabled: !!user,
      refetchInterval: 30000,
      staleTime: 15000,
    }
  )

  const { data: health } = useQuery<SystemHealth>(
    ['system', 'health'],
    () => systemAPI.health().then(res => res.data),
    {
      enabled: !!user,
      refetchInterval: 60000,
    }
  )

  const { data: metrics } = useQuery(
    ['system', 'metrics'],
    () => systemAPI.metrics().then(res => res.data),
    {
      enabled: !!user,
      refetchInterval: 15000,
    }
  )

  return {
    stats,
    health,
    metrics,
    isLoading,
    error,
    refetch,
  }
}

// ==================== useWebSocket ====================

export function useWebSocket() {
  const { user } = useAuth()
  const [connected, setConnected] = useState(false)
  const [ws, setWs] = useState<DeparrowWebSocket | null>(null)

  useEffect(() => {
    if (!user) return

    const websocket = getWebSocket()
    
    const unsubscribe = websocket.subscribe('*', (message) => {
      if (message.type === 'connected' || message.type === 'auth_success') {
        setConnected(true)
      } else if (message.type === 'disconnected') {
        setConnected(false)
      }
    })

    // Try to connect
    websocket.connect()
      .then(() => setConnected(true))
      .catch(err => {
        console.error('WebSocket connection failed:', err)
        setConnected(false)
      })

    setWs(websocket)

    return () => {
      unsubscribe()
    }
  }, [user])

  const subscribe = useCallback((type: string, callback: (data: any) => void) => {
    if (!ws) return () => {}
    return ws.subscribe(type, callback)
  }, [ws])

  const send = useCallback((type: string, data: any) => {
    ws?.send(type, data)
  }, [ws])

  return {
    connected,
    subscribe,
    send,
  }
}
