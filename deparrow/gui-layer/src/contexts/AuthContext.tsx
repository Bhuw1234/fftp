import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { authAPI } from '../api/client'

interface User {
  id: string
  email: string
  name: string
  role: 'user' | 'admin'
  credits: number
}

interface AuthContextType {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  register: (data: { email: string; password: string; name: string }) => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

interface AuthProviderProps {
  children: ReactNode
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('deparrow_token')
    if (token) {
      loadUser()
    } else {
      setLoading(false)
    }
  }, [])

  const loadUser = async () => {
    try {
      const response = await authAPI.me()
      setUser(response.data)
    } catch (error) {
      localStorage.removeItem('deparrow_token')
      localStorage.removeItem('deparrow_user')
    } finally {
      setLoading(false)
    }
  }

  const login = async (email: string, password: string) => {
    const response = await authAPI.login(email, password)
    const { token, user } = response.data
    localStorage.setItem('deparrow_token', token)
    localStorage.setItem('deparrow_user', JSON.stringify(user))
    setUser(user)
  }

  const logout = async () => {
    try {
      await authAPI.logout()
    } catch (error) {
      // Ignore logout errors
    } finally {
      localStorage.removeItem('deparrow_token')
      localStorage.removeItem('deparrow_user')
      setUser(null)
    }
  }

  const register = async (data: { email: string; password: string; name: string }) => {
    const response = await authAPI.register(data)
    const { token, user } = response.data
    localStorage.setItem('deparrow_token', token)
    localStorage.setItem('deparrow_user', JSON.stringify(user))
    setUser(user)
  }

  const value = {
    user,
    loading,
    login,
    logout,
    register,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}