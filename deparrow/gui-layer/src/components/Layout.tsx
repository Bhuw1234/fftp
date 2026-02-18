import React from 'react'
import { Outlet, Link, useLocation } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { useCredits } from '../hooks/useDeparrow'
import { 
  LayoutDashboard, 
  Briefcase, 
  Server, 
  Wallet, 
  Settings,
  Bot,
  Cloud,
  LogOut,
  Menu,
  X,
  Activity
} from 'lucide-react'
import { useState, useEffect } from 'react'

const Layout: React.FC = () => {
  const { user, logout } = useAuth()
  const { balance } = useCredits()
  const wsConnected = false
  const location = useLocation()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<'connected' | 'disconnected' | 'connecting'>('disconnected')

  const navigation = [
    { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
    { name: 'Jobs', href: '/jobs', icon: Briefcase },
    { name: 'Nodes', href: '/nodes', icon: Server },
    { name: 'Providers', href: '/providers', icon: Cloud },
    { name: 'Wallet', href: '/wallet', icon: Wallet },
    { name: 'AI Agent', href: '/agent', icon: Bot },
    { name: 'Settings', href: '/settings', icon: Settings },
  ]

  useEffect(() => {
    // Simulate connection status
    const checkConnection = () => {
      setConnectionStatus(wsConnected ? 'connected' : 'connecting')
    }
    
    checkConnection()
    const interval = setInterval(checkConnection, 5000)
    return () => clearInterval(interval)
  }, [wsConnected])

  const handleLogout = async () => {
    await logout()
  }

  const getConnectionColor = () => {
    switch (connectionStatus) {
      case 'connected': return 'bg-green-500'
      case 'disconnected': return 'bg-red-500'
      case 'connecting': return 'bg-yellow-500 animate-pulse'
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mobile sidebar */}
      <div className={`fixed inset-0 z-40 lg:hidden ${sidebarOpen ? 'block' : 'hidden'}`}>
        <div className="fixed inset-0 bg-gray-600 bg-opacity-75" onClick={() => setSidebarOpen(false)} />
        <div className="fixed inset-y-0 left-0 flex w-64 flex-col bg-white">
          <div className="flex h-16 items-center justify-between px-4 border-b">
            <div className="flex items-center">
              <div className="h-8 w-8 rounded-lg bg-gradient-to-r from-blue-500 to-cyan-500 flex items-center justify-center">
                <span className="text-white font-bold text-sm">D</span>
              </div>
              <span className="ml-2 text-lg font-semibold text-gray-900">DEparrow</span>
            </div>
            <button
              onClick={() => setSidebarOpen(false)}
              className="p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100"
            >
              <X className="h-6 w-6" />
            </button>
          </div>
          <nav className="flex-1 space-y-1 px-2 py-4 overflow-y-auto">
            {navigation.map((item) => {
              const Icon = item.icon
              const isActive = location.pathname === item.href
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  className={`flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                    isActive
                      ? 'bg-blue-50 text-blue-700'
                      : 'text-gray-700 hover:bg-gray-100 hover:text-gray-900'
                  }`}
                  onClick={() => setSidebarOpen(false)}
                >
                  <Icon className={`mr-3 h-5 w-5 ${isActive ? 'text-blue-500' : 'text-gray-400'}`} />
                  {item.name}
                </Link>
              )
            })}
          </nav>
          
          {/* Connection Status */}
          <div className="px-4 py-3 border-t">
            <div className="flex items-center space-x-2">
              <div className={`h-2 w-2 rounded-full ${getConnectionColor()}`} />
              <span className="text-xs text-gray-500 capitalize">{connectionStatus}</span>
            </div>
          </div>
          
          {/* User Info */}
          <div className="border-t p-4">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="h-8 w-8 rounded-full bg-gradient-to-r from-blue-500 to-cyan-500 flex items-center justify-center">
                  <span className="text-white font-medium text-sm">
                    {user?.name?.charAt(0).toUpperCase() || 'U'}
                  </span>
                </div>
              </div>
              <div className="ml-3 flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">{user?.name}</p>
                <p className="text-xs text-gray-500 truncate">{user?.email}</p>
              </div>
            </div>
            
            {/* Credits Balance */}
            <div className="mt-3 p-2 bg-gradient-to-r from-blue-50 to-cyan-50 rounded-md">
              <div className="flex items-center justify-between">
                <span className="text-xs text-gray-500">Balance</span>
                <span className="text-sm font-semibold text-blue-600">{balance.toFixed(2)} credits</span>
              </div>
            </div>
            
            <button
              onClick={handleLogout}
              className="mt-4 w-full flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700 transition-colors"
            >
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </button>
          </div>
        </div>
      </div>

      {/* Desktop sidebar */}
      <div className="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col">
        <div className="flex flex-col flex-grow border-r border-gray-200 bg-white pt-5">
          {/* Logo */}
          <div className="flex items-center px-4">
            <div className="h-8 w-8 rounded-lg bg-gradient-to-r from-blue-500 to-cyan-500 flex items-center justify-center">
              <span className="text-white font-bold text-sm">D</span>
            </div>
            <span className="ml-2 text-lg font-semibold text-gray-900">DEparrow</span>
          </div>
          
          {/* Navigation */}
          <nav className="mt-8 flex-1 space-y-1 px-2 overflow-y-auto">
            {navigation.map((item) => {
              const Icon = item.icon
              const isActive = location.pathname === item.href
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  className={`flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                    isActive
                      ? 'bg-blue-50 text-blue-700'
                      : 'text-gray-700 hover:bg-gray-100 hover:text-gray-900'
                  }`}
                >
                  <Icon className={`mr-3 h-5 w-5 ${isActive ? 'text-blue-500' : 'text-gray-400'}`} />
                  {item.name}
                  {item.name === 'AI Agent' && (
                    <span className="ml-auto inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-700">
                      Beta
                    </span>
                  )}
                </Link>
              )
            })}
          </nav>
          
          {/* Connection Status */}
          <div className="px-4 py-3 border-t">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <div className={`h-2 w-2 rounded-full ${getConnectionColor()}`} />
                <span className="text-xs text-gray-500 capitalize">{connectionStatus}</span>
              </div>
              <Activity className="h-4 w-4 text-gray-400" />
            </div>
          </div>
          
          {/* User Info */}
          <div className="border-t p-4">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="h-8 w-8 rounded-full bg-gradient-to-r from-blue-500 to-cyan-500 flex items-center justify-center">
                  <span className="text-white font-medium text-sm">
                    {user?.name?.charAt(0).toUpperCase() || 'U'}
                  </span>
                </div>
              </div>
              <div className="ml-3 flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">{user?.name}</p>
                <p className="text-xs text-gray-500 truncate">{user?.email}</p>
              </div>
            </div>
            
            {/* Credits Balance */}
            <div className="mt-3 p-2 bg-gradient-to-r from-blue-50 to-cyan-50 rounded-md">
              <div className="flex items-center justify-between">
                <span className="text-xs text-gray-500">Balance</span>
                <span className="text-sm font-semibold text-blue-600">{balance.toFixed(2)} credits</span>
              </div>
            </div>
            
            <button
              onClick={handleLogout}
              className="mt-4 w-full flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700 transition-colors"
            >
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </button>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="lg:pl-64">
        {/* Mobile header */}
        <div className="sticky top-0 z-10 flex h-16 flex-shrink-0 bg-white border-b border-gray-200 lg:hidden">
          <button
            type="button"
            className="px-4 text-gray-500 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
            onClick={() => setSidebarOpen(true)}
          >
            <span className="sr-only">Open sidebar</span>
            <Menu className="h-6 w-6" />
          </button>
          <div className="flex flex-1 justify-between px-4">
            <div className="flex flex-1 items-center">
              <div className="flex items-center space-x-2">
                <div className="h-8 w-8 rounded-lg bg-gradient-to-r from-blue-500 to-cyan-500 flex items-center justify-center">
                  <span className="text-white font-bold text-sm">D</span>
                </div>
                <h1 className="text-lg font-semibold text-gray-900">
                  {navigation.find((item) => item.href === location.pathname)?.name || 'DEparrow'}
                </h1>
              </div>
            </div>
            <div className="ml-4 flex items-center space-x-4">
              <div className="flex items-center space-x-1">
                <div className={`h-2 w-2 rounded-full ${getConnectionColor()}`} />
              </div>
              <div className="flex-shrink-0">
                <div className="h-8 w-8 rounded-full bg-gradient-to-r from-blue-500 to-cyan-500 flex items-center justify-center">
                  <span className="text-white font-medium text-sm">
                    {user?.name?.charAt(0).toUpperCase() || 'U'}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Page content */}
        <main className="py-6">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <Outlet />
          </div>
        </main>
        
        {/* Footer */}
        <footer className="border-t border-gray-200 py-4">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="flex items-center justify-between text-xs text-gray-500">
              <div className="flex items-center space-x-4">
                <span>© 2024 DEparrow</span>
                <span>•</span>
                <span>Global Virtual Machine</span>
              </div>
              <div className="flex items-center space-x-4">
                <a href="#" className="hover:text-gray-700">Documentation</a>
                <a href="#" className="hover:text-gray-700">Support</a>
                <a href="#" className="hover:text-gray-700">Status</a>
              </div>
            </div>
          </div>
        </footer>
      </div>
    </div>
  )
}

export default Layout
