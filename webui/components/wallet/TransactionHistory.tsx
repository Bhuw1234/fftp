'use client'

import React from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { 
  ArrowDownLeft, 
  ArrowUpRight, 
  ArrowRightLeft,
  Clock,
  CheckCircle2,
  AlertCircle,
  ExternalLink
} from 'lucide-react'

export interface Transaction {
  id: string
  type: 'earning' | 'spending' | 'transfer'
  amount: string
  description: string
  timestamp: string
  status: 'completed' | 'pending' | 'failed'
  jobId?: string
}

interface TransactionHistoryProps {
  transactions: Transaction[]
}

const typeConfig = {
  earning: {
    icon: ArrowDownLeft,
    label: 'Earning',
    color: 'text-green-600 bg-green-50 dark:bg-green-950'
  },
  spending: {
    icon: ArrowUpRight,
    label: 'Spending',
    color: 'text-red-600 bg-red-50 dark:bg-red-950'
  },
  transfer: {
    icon: ArrowRightLeft,
    label: 'Transfer',
    color: 'text-blue-600 bg-blue-50 dark:bg-blue-950'
  }
}

const statusConfig = {
  completed: {
    icon: CheckCircle2,
    label: 'Completed',
    color: 'default'
  },
  pending: {
    icon: Clock,
    label: 'Pending',
    color: 'secondary'
  },
  failed: {
    icon: AlertCircle,
    label: 'Failed',
    color: 'destructive'
  }
}

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  
  return date.toLocaleDateString()
}

export function TransactionHistory({ transactions }: TransactionHistoryProps) {
  if (transactions.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <Clock className="h-12 w-12 mx-auto mb-4 opacity-50" />
        <p>No transactions yet</p>
        <p className="text-sm mt-2">Complete compute jobs to earn DPC tokens</p>
      </div>
    )
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[100px]">Type</TableHead>
            <TableHead>Description</TableHead>
            <TableHead className="text-right">Amount</TableHead>
            <TableHead className="w-[120px]">Status</TableHead>
            <TableHead className="w-[140px]">Time</TableHead>
            <TableHead className="w-[80px]">Details</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {transactions.map((tx) => {
            const type = typeConfig[tx.type]
            const status = statusConfig[tx.status]
            const TypeIcon = type.icon
            const StatusIcon = status.icon
            const isEarning = tx.type === 'earning'
            const amount = parseFloat(tx.amount)

            return (
              <TableRow key={tx.id}>
                <TableCell>
                  <div className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-md ${type.color}`}>
                    <TypeIcon className="h-3.5 w-3.5" />
                    <span className="text-xs font-medium">{type.label}</span>
                  </div>
                </TableCell>
                <TableCell>
                  <div className="max-w-[300px]">
                    <div className="font-medium truncate">{tx.description}</div>
                    {tx.jobId && (
                      <div className="text-xs text-muted-foreground font-mono">
                        Job: {tx.jobId}
                      </div>
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-right">
                  <span className={`font-semibold ${isEarning ? 'text-green-600' : 'text-red-600'}`}>
                    {isEarning ? '+' : '-'}{amount.toFixed(2)} DPC
                  </span>
                </TableCell>
                <TableCell>
                  <Badge variant={status.color as any} className="gap-1">
                    <StatusIcon className="h-3 w-3" />
                    {status.label}
                  </Badge>
                </TableCell>
                <TableCell className="text-muted-foreground text-sm">
                  {formatTimestamp(tx.timestamp)}
                </TableCell>
                <TableCell>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 w-8 p-0"
                    title="View transaction details"
                  >
                    <ExternalLink className="h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
