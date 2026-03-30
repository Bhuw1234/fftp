'use client'

import React from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts'
import { Coins, TrendingUp, Calendar } from 'lucide-react'

interface Transaction {
  id: string
  type: 'earning' | 'spending' | 'transfer'
  amount: string
  description: string
  timestamp: string
  status: 'completed' | 'pending' | 'failed'
  jobId?: string
}

interface EarningsChartProps {
  transactions: Transaction[]
}

// Generate earnings data by day for the last 7 days
function generateEarningsData(transactions: Transaction[]) {
  const last7Days: { date: string; earnings: number; jobs: number }[] = []
  const now = new Date()
  
  for (let i = 6; i >= 0; i--) {
    const date = new Date(now)
    date.setDate(date.getDate() - i)
    const dateStr = date.toISOString().split('T')[0]
    const dayName = date.toLocaleDateString('en-US', { weekday: 'short' })
    
    const dayTransactions = transactions.filter(tx => {
      if (tx.type !== 'earning') return false
      const txDate = new Date(tx.timestamp).toISOString().split('T')[0]
      return txDate === dateStr
    })
    
    const earnings = dayTransactions.reduce((sum, tx) => sum + parseFloat(tx.amount), 0)
    
    last7Days.push({
      date: dayName,
      earnings,
      jobs: dayTransactions.length
    })
  }
  
  return last7Days
}

// Generate hourly earnings for today
function generateHourlyData(transactions: Transaction[]) {
  const today = new Date().toISOString().split('T')[0]
  const hourlyData: { hour: string; earnings: number }[] = []
  
  for (let i = 0; i < 24; i++) {
    const hourStr = i.toString().padStart(2, '0') + ':00'
    
    const hourTransactions = transactions.filter(tx => {
      if (tx.type !== 'earning') return false
      const txDate = new Date(tx.timestamp)
      const txDateStr = txDate.toISOString().split('T')[0]
      const txHour = txDate.getHours()
      return txDateStr === today && txHour === i
    })
    
    const earnings = hourTransactions.reduce((sum, tx) => sum + parseFloat(tx.amount), 0)
    
    hourlyData.push({
      hour: hourStr,
      earnings
    })
  }
  
  return hourlyData
}

export function EarningsChart({ transactions }: EarningsChartProps) {
  const weeklyData = generateEarningsData(transactions)
  const hourlyData = generateHourlyData(transactions)
  
  const totalWeeklyEarnings = weeklyData.reduce((sum, d) => sum + d.earnings, 0)
  const totalWeeklyJobs = weeklyData.reduce((sum, d) => sum + d.jobs, 0)
  const avgDailyEarnings = totalWeeklyEarnings / 7

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Coins className="h-5 w-5" />
          Earnings Overview
        </CardTitle>
        <CardDescription>Your DPC earnings over the last 7 days</CardDescription>
      </CardHeader>
      <CardContent>
        {/* Summary Stats */}
        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="text-center p-4 rounded-lg bg-muted">
            <div className="text-2xl font-bold text-green-600">{totalWeeklyEarnings.toFixed(2)}</div>
            <div className="text-sm text-muted-foreground">DPC Earned (7d)</div>
          </div>
          <div className="text-center p-4 rounded-lg bg-muted">
            <div className="text-2xl font-bold">{totalWeeklyJobs}</div>
            <div className="text-sm text-muted-foreground">Jobs Completed</div>
          </div>
          <div className="text-center p-4 rounded-lg bg-muted">
            <div className="text-2xl font-bold text-blue-600">{avgDailyEarnings.toFixed(2)}</div>
            <div className="text-sm text-muted-foreground">Avg Daily</div>
          </div>
        </div>

        {/* Weekly Earnings Chart */}
        <div className="h-[200px] w-full">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={weeklyData}>
              <defs>
                <linearGradient id="colorEarnings" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3}/>
                  <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0}/>
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis 
                dataKey="date" 
                className="text-xs"
                tick={{ fill: 'hsl(var(--muted-foreground))' }}
              />
              <YAxis 
                className="text-xs"
                tick={{ fill: 'hsl(var(--muted-foreground))' }}
                tickFormatter={(value) => `${value} DPC`}
              />
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: 'hsl(var(--card))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '8px'
                }}
                formatter={(value) => [`${Number(value).toFixed(2)} DPC`, 'Earnings']}
              />
              <Area 
                type="monotone" 
                dataKey="earnings" 
                stroke="hsl(var(--primary))" 
                fillOpacity={1} 
                fill="url(#colorEarnings)"
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Job Count Bar Chart */}
        <div className="mt-6">
          <h4 className="text-sm font-medium mb-3 flex items-center gap-2">
            <Calendar className="h-4 w-4" />
            Jobs per Day
          </h4>
          <div className="h-[80px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={weeklyData}>
                <XAxis 
                  dataKey="date" 
                  className="text-xs"
                  tick={{ fill: 'hsl(var(--muted-foreground))' }}
                />
                <YAxis 
                  className="text-xs"
                  tick={{ fill: 'hsl(var(--muted-foreground))' }}
                  allowDecimals={false}
                />
                <Tooltip 
                  contentStyle={{ 
                    backgroundColor: 'hsl(var(--card))',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '8px'
                  }}
                  formatter={(value) => [value, 'Jobs']}
                />
                <Bar 
                  dataKey="jobs" 
                  fill="hsl(var(--chart-1))" 
                  radius={[4, 4, 0, 0]}
                />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
