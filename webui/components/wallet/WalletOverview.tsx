'use client'

import React, { useState, useEffect, useCallback } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { 
  Wallet as WalletIcon, 
  Copy, 
  RefreshCw, 
  Coins, 
  TrendingUp,
  Clock,
  CheckCircle2,
  AlertCircle,
  ExternalLink
} from 'lucide-react'
import { TransactionHistory } from './TransactionHistory'
import { EarningsChart } from './EarningsChart'

// DPC Blockchain RPC endpoint
const DPC_RPC_ENDPOINT = process.env.NEXT_PUBLIC_DPC_RPC || 'http://34.180.51.11:26657'
const DPC_CHAIN_ID = process.env.NEXT_PUBLIC_DPC_CHAIN_ID || 'dpc-testnet-1'

interface WalletState {
  address: string
  balance: string
  balanceFormatted: string
  chainId: string
  blockHeight: number
  isConnected: boolean
  isLoading: boolean
  error: string | null
}

interface Transaction {
  id: string
  type: 'earning' | 'spending' | 'transfer'
  amount: string
  description: string
  timestamp: string
  status: 'completed' | 'pending' | 'failed'
  jobId?: string
}

// Mock transaction data for demo (will be replaced with real blockchain data)
const mockTransactions: Transaction[] = [
  {
    id: 'tx-001',
    type: 'earning',
    amount: '2.5',
    description: 'Compute job completed: Model training',
    timestamp: new Date(Date.now() - 3600000).toISOString(),
    status: 'completed',
    jobId: 'j-abc123'
  },
  {
    id: 'tx-002',
    type: 'earning',
    amount: '1.2',
    description: 'Compute job completed: Data processing',
    timestamp: new Date(Date.now() - 7200000).toISOString(),
    status: 'completed',
    jobId: 'j-def456'
  },
  {
    id: 'tx-003',
    type: 'spending',
    amount: '0.5',
    description: 'Job submission fee',
    timestamp: new Date(Date.now() - 10800000).toISOString(),
    status: 'completed'
  },
  {
    id: 'tx-004',
    type: 'earning',
    amount: '5.0',
    description: 'Compute job completed: GPU rendering',
    timestamp: new Date(Date.now() - 86400000).toISOString(),
    status: 'completed',
    jobId: 'j-ghi789'
  }
]

export function WalletOverview() {
  const [wallet, setWallet] = useState<WalletState>({
    address: '',
    balance: '0',
    balanceFormatted: '0.00 DPC',
    chainId: DPC_CHAIN_ID,
    blockHeight: 0,
    isConnected: false,
    isLoading: true,
    error: null
  })
  
  const [transactions] = useState<Transaction[]>(mockTransactions)
  const [copied, setCopied] = useState(false)
  const [isRefreshDisabled, setIsRefreshDisabled] = useState(false)

  // Generate or load wallet address from localStorage
  const getWalletAddress = useCallback(() => {
    if (typeof window !== 'undefined') {
      let address = localStorage.getItem('dpc_wallet_address')
      if (!address) {
        // Generate a mock address (in production, this would use proper crypto)
        address = 'dpc1' + Math.random().toString(36).substring(2, 15) + 
                  Math.random().toString(36).substring(2, 15) + 
                  Math.random().toString(36).substring(2, 15)
        localStorage.setItem('dpc_wallet_address', address)
      }
      return address
    }
    return ''
  }, [])

  // Fetch blockchain status and balance
  const fetchWalletData = useCallback(async () => {
    const address = getWalletAddress()
    
    try {
      // Fetch chain status
      const statusResponse = await fetch(`${DPC_RPC_ENDPOINT}/status`)
      const statusData = await statusResponse.json()
      const blockHeight = parseInt(statusData.result?.sync_info?.latest_block_height || '0')
      
      // Try to fetch balance from blockchain
      let balance = '0'
      try {
        const balanceResponse = await fetch(
          `${DPC_RPC_ENDPOINT}/abci_query?path="/accounts/${address}"`
        )
        const balanceData = await balanceResponse.json()
        // Parse balance from response (format depends on chain implementation)
        const balanceValue = balanceData.result?.response?.value
        if (balanceValue) {
          const decoded = atob(balanceValue)
          const parsed = JSON.parse(decoded)
          balance = parsed.balance || '0'
        }
      } catch {
        // If balance query fails, show demo balance
        balance = '100.5'
      }

      const balanceNum = parseFloat(balance)
      const balanceFormatted = balanceNum.toFixed(2) + ' DPC'

      setWallet({
        address,
        balance,
        balanceFormatted,
        chainId: DPC_CHAIN_ID,
        blockHeight,
        isConnected: true,
        isLoading: false,
        error: null
      })
    } catch (error) {
      console.error('Error fetching wallet data:', error)
      setWallet(prev => ({
        ...prev,
        isLoading: false,
        error: 'Failed to connect to DPC blockchain',
        isConnected: false
      }))
    }
  }, [getWalletAddress])

  useEffect(() => {
    fetchWalletData()
  }, [fetchWalletData])

  const handleRefresh = useCallback(() => {
    setIsRefreshDisabled(true)
    fetchWalletData().then(() => {
      setTimeout(() => setIsRefreshDisabled(false), 1000)
    })
  }, [fetchWalletData])

  const handleCopyAddress = () => {
    navigator.clipboard.writeText(wallet.address)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleConnectWallet = () => {
    // In production, this would connect to a real wallet (Keplr, etc.)
    fetchWalletData()
  }

  // Calculate earnings summary
  const totalEarnings = transactions
    .filter(tx => tx.type === 'earning')
    .reduce((sum, tx) => sum + parseFloat(tx.amount), 0)
  
  const pendingEarnings = transactions
    .filter(tx => tx.type === 'earning' && tx.status === 'pending')
    .reduce((sum, tx) => sum + parseFloat(tx.amount), 0)

  if (wallet.isLoading) {
    return (
      <div className="container mx-auto">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="container mx-auto">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Wallet</h1>
        <div className="flex items-center gap-2">
          <Badge variant={wallet.isConnected ? "default" : "secondary"} className="text-sm px-3 py-1">
            {wallet.isConnected ? (
              <>
                <CheckCircle2 className="h-4 w-4 mr-2" />
                Connected
              </>
            ) : (
              <>
                <AlertCircle className="h-4 w-4 mr-2" />
                Disconnected
              </>
            )}
          </Badge>
          <Button
            onClick={handleRefresh}
            disabled={isRefreshDisabled}
            variant="outline"
            size="icon"
            aria-label="Refresh wallet"
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {wallet.error && (
        <Card className="mb-6 border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 text-destructive">
              <AlertCircle className="h-5 w-5" />
              <span>{wallet.error}</span>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 mb-8">
        {/* Balance Card */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <WalletIcon className="h-5 w-5" />
              DPC Balance
            </CardTitle>
            <CardDescription>Your DEparrow Coin balance on {wallet.chainId}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-4">
              <div>
                <div className="text-4xl font-bold text-primary mb-2">
                  {wallet.balanceFormatted}
                </div>
                <div className="text-sm text-muted-foreground">
                  Block Height: {wallet.blockHeight.toLocaleString()}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Input
                  value={wallet.address}
                  readOnly
                  className="font-mono text-sm bg-muted"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={handleCopyAddress}
                  className="shrink-0"
                >
                  {copied ? <CheckCircle2 className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Quick Stats Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-5 w-5" />
              Earnings
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Total Earned</span>
                <span className="font-semibold text-green-600">{totalEarnings.toFixed(2)} DPC</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Pending</span>
                <span className="font-semibold text-yellow-600">{pendingEarnings.toFixed(2)} DPC</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Jobs Completed</span>
                <span className="font-semibold">{transactions.filter(tx => tx.type === 'earning').length}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Earnings Chart */}
      <div className="mb-8">
        <EarningsChart transactions={transactions} />
      </div>

      {/* Transaction History */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Clock className="h-5 w-5" />
            Transaction History
          </CardTitle>
          <CardDescription>
            Your recent DPC transactions and earnings
          </CardDescription>
        </CardHeader>
        <CardContent>
          <TransactionHistory transactions={transactions} />
        </CardContent>
      </Card>

      {/* Network Info */}
      <div className="mt-6 text-sm text-muted-foreground">
        <div className="flex items-center gap-4 flex-wrap">
          <span>Network: <Badge variant="outline">{wallet.chainId}</Badge></span>
          <span>RPC: <code className="text-xs">{DPC_RPC_ENDPOINT}</code></span>
          <a 
            href={`${DPC_RPC_ENDPOINT}/status`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 hover:text-primary"
          >
            <ExternalLink className="h-3 w-3" />
            View Chain Status
          </a>
        </div>
      </div>
    </div>
  )
}
