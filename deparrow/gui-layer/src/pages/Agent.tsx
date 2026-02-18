import React, { useState, useEffect, useRef } from 'react'
import { 
  Bot, 
  Send, 
  Zap, 
  Server, 
  Wallet,
  Activity,
  Terminal,
  Cpu,
  HardDrive,
  RefreshCw,
  Pause,
  Play,
  Trash2,
  ChevronDown,
  ChevronUp,
  Sparkles
} from 'lucide-react'
import { agentAPI } from '../api/deparrow'
import toast from 'react-hot-toast'
import { format } from 'date-fns'

interface AgentStatusType {
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

interface AgentTool {
  name: string
  description: string
  enabled: boolean
  calls: number
}

interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
  tool_calls?: ToolCall[]
}

interface ToolCall {
  tool: string
  args: Record<string, any>
  result: any
  status: 'pending' | 'success' | 'error'
}

const Agent: React.FC = () => {
  const [agent, setAgent] = useState<AgentStatusType | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [inputMessage, setInputMessage] = useState('')
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(false)
  const [showTools, setShowTools] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    loadAgentData()
    // Simulate real-time updates
    const interval = setInterval(loadAgentData, 5000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const loadAgentData = async () => {
    try {
      const response = await agentAPI.status()
      setAgent(response.data)
    } catch (error: any) {
      console.error('Agent data loading error:', error)
      // Load mock data
      loadMockAgentData()
    } finally {
      setLoading(false)
    }
  }

  const loadMockAgentData = () => {
    const mockAgent: AgentStatusType = {
      id: 'agent-pico-001',
      name: 'PicoClaw Agent',
      status: 'running',
      uptime: '2d 14h 32m',
      credits_earned: 45.75,
      credits_spent: 12.50,
      tasks_completed: 23,
      current_task: 'Monitoring compute opportunities...',
      tools: [
        { name: 'job_submit', description: 'Submit compute jobs to the network', enabled: true, calls: 15 },
        { name: 'job_status', description: 'Check status of running jobs', enabled: true, calls: 42 },
        { name: 'node_discover', description: 'Discover available compute nodes', enabled: true, calls: 8 },
        { name: 'credit_check', description: 'Check credit balance', enabled: true, calls: 56 },
        { name: 'wallet_transfer', description: 'Transfer credits to other users', enabled: false, calls: 0 },
        { name: 'task_schedule', description: 'Schedule automated tasks', enabled: true, calls: 12 },
      ],
      resources: {
        cpu_usage: 23.5,
        memory_usage: 45.2,
        disk_usage: 12.8
      },
      last_heartbeat: new Date().toISOString()
    }
    setAgent(mockAgent)

    // Load mock chat history
    setMessages([
      {
        id: '1',
        role: 'assistant',
        content: 'Hello! I\'m your PicoClaw AI Agent. I can help you manage compute jobs, monitor nodes, and optimize your credit earnings. What would you like me to do?',
        timestamp: new Date(Date.now() - 3600000).toISOString()
      },
      {
        id: '2',
        role: 'user',
        content: 'What\'s my current credit balance?',
        timestamp: new Date(Date.now() - 3500000).toISOString()
      },
      {
        id: '3',
        role: 'assistant',
        content: 'Your current credit balance is 150.75 credits. You\'ve earned 45.75 credits and spent 12.50 credits in the last 24 hours. Would you like me to submit any compute jobs to earn more credits?',
        timestamp: new Date(Date.now() - 3400000).toISOString(),
        tool_calls: [
          { tool: 'credit_check', args: {}, result: { balance: 150.75 }, status: 'success' }
        ]
      },
      {
        id: '4',
        role: 'user',
        content: 'Are there any high-paying jobs available?',
        timestamp: new Date(Date.now() - 1800000).toISOString()
      },
      {
        id: '5',
        role: 'assistant',
        content: 'I found 3 high-paying jobs currently available:\n\n1. **ML Model Training** - 25 credits (GPU required)\n2. **Data Processing Pipeline** - 18 credits\n3. **Video Encoding** - 15 credits\n\nWould you like me to bid on any of these jobs?',
        timestamp: new Date(Date.now() - 1700000).toISOString(),
        tool_calls: [
          { tool: 'job_status', args: { filter: 'high_pay' }, result: { jobs: 3 }, status: 'success' }
        ]
      }
    ])
  }

  const handleSendMessage = async () => {
    if (!inputMessage.trim() || sending) return

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: inputMessage,
      timestamp: new Date().toISOString()
    }

    setMessages(prev => [...prev, userMessage])
    setInputMessage('')
    setSending(true)

    try {
      const response = await agentAPI.chat(inputMessage)
      setMessages(prev => [...prev, response.data.message])
    } catch (error: any) {
      // Simulate response for demo
      const assistantMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: generateMockResponse(inputMessage),
        timestamp: new Date().toISOString(),
        tool_calls: Math.random() > 0.5 ? [
          { tool: 'task_process', args: { query: inputMessage }, result: { success: true }, status: 'success' }
        ] : undefined
      }
      setMessages(prev => [...prev, assistantMessage])
    } finally {
      setSending(false)
    }
  }

  const generateMockResponse = (_input: string): string => {
    const responses = [
      'I understand. Let me process that for you...',
      'Great question! Based on the current network status, I recommend the following action...',
      'I\'ve analyzed the available resources and found some opportunities for you.',
      'Processing your request. The compute network is responding well today.',
      'I can help you with that. Let me check the current network conditions first.'
    ]
    return responses[Math.floor(Math.random() * responses.length)]
  }

  const handleAgentAction = async (action: 'start' | 'pause' | 'stop') => {
    try {
      if (action === 'start') {
        await agentAPI.start()
        toast.success('Agent started successfully')
      } else if (action === 'pause') {
        await agentAPI.pause()
        toast.success('Agent paused')
      } else {
        await agentAPI.stop()
        toast.success('Agent stopped')
      }
      loadAgentData()
    } catch (error: any) {
      // Update mock data
      if (agent) {
        setAgent({
          ...agent,
          status: action === 'start' ? 'running' : action === 'pause' ? 'paused' : 'stopped'
        })
      }
      toast.success(`Agent ${action}${action === 'pause' ? 'd' : 'ed'}`)
    }
  }

  const toggleTool = async (toolName: string, enabled: boolean) => {
    try {
      await agentAPI.updateTool(toolName, enabled)
      toast.success(`Tool ${toolName} ${enabled ? 'enabled' : 'disabled'}`)
      loadAgentData()
    } catch (error: any) {
      // Update mock data
      if (agent) {
        setAgent({
          ...agent,
          tools: agent.tools.map(t => 
            t.name === toolName ? { ...t, enabled } : t
          )
        })
      }
      toast.success(`Tool ${toolName} ${enabled ? 'enabled' : 'disabled'}`)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'bg-green-100 text-green-800'
      case 'paused': return 'bg-yellow-100 text-yellow-800'
      case 'stopped': return 'bg-gray-100 text-gray-800'
      case 'error': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running': return <Play className="h-4 w-4 text-green-500" />
      case 'paused': return <Pause className="h-4 w-4 text-yellow-500" />
      case 'stopped': return <Pause className="h-4 w-4 text-gray-500" />
      case 'error': return <Activity className="h-4 w-4 text-red-500" />
      default: return <Activity className="h-4 w-4 text-gray-500" />
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
          <h1 className="text-2xl font-bold text-gray-900">AI Agent</h1>
          <p className="text-gray-600">
            Manage your PicoClaw AI agent and interact with the compute network.
          </p>
        </div>
        <div className="mt-4 sm:mt-0 flex space-x-2">
          {agent?.status === 'running' ? (
            <button
              onClick={() => handleAgentAction('pause')}
              className="inline-flex items-center px-4 py-2 border border-yellow-300 shadow-sm text-sm font-medium rounded-md text-yellow-700 bg-yellow-50 hover:bg-yellow-100"
            >
              <Pause className="mr-2 h-4 w-4" />
              Pause
            </button>
          ) : (
            <button
              onClick={() => handleAgentAction('start')}
              className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700"
            >
              <Play className="mr-2 h-4 w-4" />
              Start
            </button>
          )}
          <button
            onClick={() => handleAgentAction('stop')}
            className="inline-flex items-center px-4 py-2 border border-red-300 shadow-sm text-sm font-medium rounded-md text-red-700 bg-red-50 hover:bg-red-100"
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Stop
          </button>
        </div>
      </div>

      {/* Agent Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center justify-between">
              <div>
                <dt className="text-sm font-medium text-gray-500">Status</dt>
                <dd className="mt-1 flex items-center space-x-2">
                  {getStatusIcon(agent?.status || 'stopped')}
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(agent?.status || 'stopped')}`}>
                    {agent?.status || 'Unknown'}
                  </span>
                </dd>
              </div>
              <Bot className="h-8 w-8 text-gray-400" />
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center justify-between">
              <div>
                <dt className="text-sm font-medium text-gray-500">Uptime</dt>
                <dd className="mt-1 text-lg font-semibold text-gray-900">
                  {agent?.uptime || '0m'}
                </dd>
              </div>
              <Activity className="h-8 w-8 text-gray-400" />
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center justify-between">
              <div>
                <dt className="text-sm font-medium text-gray-500">Tasks Completed</dt>
                <dd className="mt-1 text-lg font-semibold text-gray-900">
                  {agent?.tasks_completed || 0}
                </dd>
              </div>
              <Zap className="h-8 w-8 text-gray-400" />
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center justify-between">
              <div>
                <dt className="text-sm font-medium text-gray-500">Net Credits</dt>
                <dd className="mt-1 text-lg font-semibold text-green-600">
                  +{((agent?.credits_earned || 0) - (agent?.credits_spent || 0)).toFixed(2)}
                </dd>
              </div>
              <Wallet className="h-8 w-8 text-gray-400" />
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Chat Interface */}
        <div className="lg:col-span-2 bg-white shadow rounded-lg flex flex-col h-[600px]">
          <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Sparkles className="h-5 w-5 text-blue-500" />
              <h3 className="text-lg font-medium text-gray-900">Chat with Agent</h3>
            </div>
            <button
              onClick={() => setMessages([])}
              className="text-sm text-gray-500 hover:text-gray-700"
            >
              Clear Chat
            </button>
          </div>

          {/* Messages */}
          <div className="flex-1 overflow-y-auto p-4 space-y-4">
            {messages.map((message) => (
              <div
                key={message.id}
                className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
              >
                <div className={`max-w-[80%] ${message.role === 'user' ? 'order-2' : 'order-1'}`}>
                  <div className={`rounded-lg px-4 py-2 ${
                    message.role === 'user'
                      ? 'bg-blue-600 text-white'
                      : message.role === 'system'
                      ? 'bg-gray-100 text-gray-800'
                      : 'bg-gray-100 text-gray-900'
                  }`}>
                    <p className="text-sm whitespace-pre-wrap">{message.content}</p>
                  </div>
                  {message.tool_calls && message.tool_calls.length > 0 && (
                    <div className="mt-2 space-y-1">
                      {message.tool_calls.map((tc, idx) => (
                        <div key={idx} className="flex items-center space-x-2 text-xs text-gray-500 bg-gray-50 rounded px-2 py-1">
                          <Terminal className="h-3 w-3" />
                          <span>Tool: {tc.tool}</span>
                          <span className={`inline-flex items-center px-1.5 py-0.5 rounded-full text-xs font-medium ${
                            tc.status === 'success' ? 'bg-green-100 text-green-800' : 
                            tc.status === 'error' ? 'bg-red-100 text-red-800' : 
                            'bg-yellow-100 text-yellow-800'
                          }`}>
                            {tc.status}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                  <p className="mt-1 text-xs text-gray-400">
                    {format(new Date(message.timestamp), 'HH:mm')}
                  </p>
                </div>
              </div>
            ))}
            <div ref={messagesEndRef} />
          </div>

          {/* Input */}
          <div className="border-t border-gray-200 p-4">
            <div className="flex space-x-2">
              <input
                type="text"
                value={inputMessage}
                onChange={(e) => setInputMessage(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSendMessage()}
                placeholder="Ask your agent something..."
                className="flex-1 rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                disabled={sending}
              />
              <button
                onClick={handleSendMessage}
                disabled={sending || !inputMessage.trim()}
                className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {sending ? (
                  <RefreshCw className="h-4 w-4 animate-spin" />
                ) : (
                  <Send className="h-4 w-4" />
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Sidebar - Tools & Resources */}
        <div className="space-y-6">
          {/* Resource Usage */}
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Resource Usage</h3>
            <div className="space-y-4">
              <div>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm text-gray-500 flex items-center">
                    <Cpu className="h-4 w-4 mr-1" /> CPU
                  </span>
                  <span className="text-sm font-medium text-gray-900">{agent?.resources.cpu_usage}%</span>
                </div>
                <div className="bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-blue-600 h-2 rounded-full transition-all duration-300" 
                    style={{ width: `${agent?.resources.cpu_usage || 0}%` }}
                  ></div>
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm text-gray-500 flex items-center">
                    <Server className="h-4 w-4 mr-1" /> Memory
                  </span>
                  <span className="text-sm font-medium text-gray-900">{agent?.resources.memory_usage}%</span>
                </div>
                <div className="bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-green-600 h-2 rounded-full transition-all duration-300" 
                    style={{ width: `${agent?.resources.memory_usage || 0}%` }}
                  ></div>
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm text-gray-500 flex items-center">
                    <HardDrive className="h-4 w-4 mr-1" /> Disk
                  </span>
                  <span className="text-sm font-medium text-gray-900">{agent?.resources.disk_usage}%</span>
                </div>
                <div className="bg-gray-200 rounded-full h-2">
                  <div 
                    className="bg-purple-600 h-2 rounded-full transition-all duration-300" 
                    style={{ width: `${agent?.resources.disk_usage || 0}%` }}
                  ></div>
                </div>
              </div>
            </div>
          </div>

          {/* Agent Tools */}
          <div className="bg-white shadow rounded-lg">
            <button
              onClick={() => setShowTools(!showTools)}
              className="w-full px-6 py-4 flex items-center justify-between border-b border-gray-200"
            >
              <div className="flex items-center space-x-2">
                <Zap className="h-5 w-5 text-yellow-500" />
                <h3 className="text-lg font-medium text-gray-900">Agent Tools</h3>
              </div>
              {showTools ? <ChevronUp className="h-5 w-5" /> : <ChevronDown className="h-5 w-5" />}
            </button>
            
            {showTools && (
              <div className="p-4 space-y-2">
                {agent?.tools.map((tool) => (
                  <div key={tool.name} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                    <div className="flex-1">
                      <p className="text-sm font-medium text-gray-900">{tool.name}</p>
                      <p className="text-xs text-gray-500">{tool.description}</p>
                    </div>
                    <div className="flex items-center space-x-2">
                      <span className="text-xs text-gray-500">{tool.calls} calls</span>
                      <input
                        type="checkbox"
                        checked={tool.enabled}
                        onChange={(e) => toggleTool(tool.name, e.target.checked)}
                        className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Current Task */}
          {agent?.current_task && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="flex items-center space-x-2 mb-2">
                <Activity className="h-5 w-5 text-blue-500 animate-pulse" />
                <h4 className="text-sm font-medium text-blue-900">Current Task</h4>
              </div>
              <p className="text-sm text-blue-700">{agent.current_task}</p>
            </div>
          )}

          {/* Agent Info */}
          <div className="bg-white shadow rounded-lg p-4">
            <h4 className="text-sm font-medium text-gray-900 mb-2">Agent Info</h4>
            <dl className="space-y-1 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-500">ID</dt>
                <dd className="text-gray-900 font-mono text-xs">{agent?.id}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Last Heartbeat</dt>
                <dd className="text-gray-900">{agent?.last_heartbeat ? format(new Date(agent.last_heartbeat), 'HH:mm:ss') : '-'}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Credits Earned</dt>
                <dd className="text-green-600">+{agent?.credits_earned.toFixed(2)}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">Credits Spent</dt>
                <dd className="text-red-600">-{agent?.credits_spent.toFixed(2)}</dd>
              </div>
            </dl>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Agent
