import React, { useState, useEffect } from 'react'
import { 
  Wallet as WalletIcon, 
  Plus, 
  Minus, 
  ArrowUpRight, 
  ArrowDownLeft,
  TrendingUp,
  TrendingDown,
  CreditCard,
  Download,
  Upload
} from 'lucide-react'
import { walletAPI } from '../api/client'
import toast from 'react-hot-toast'
import { format } from 'date-fns'

interface Transaction {
  id: string
  type: 'earn' | 'spend' | 'transfer_in' | 'transfer_out'
  amount: number
  description: string
  timestamp: string
  status: 'completed' | 'pending' | 'failed'
  job_id?: string
  counterparty?: string
}

interface CreditBalance {
  balance: number
  total_earned: number
  total_spent: number
  pending_transactions: number
}

const Wallet: React.FC = () => {
  const [balance, setBalance] = useState<CreditBalance | null>(null)
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [showDepositModal, setShowDepositModal] = useState(false)
  const [showWithdrawModal, setShowWithdrawModal] = useState(false)
  const [showTransferModal, setShowTransferModal] = useState(false)
  const [depositAmount, setDepositAmount] = useState('')
  const [withdrawAmount, setWithdrawAmount] = useState('')
  const [transferAmount, setTransferAmount] = useState('')
  const [transferUserId, setTransferUserId] = useState('')

  useEffect(() => {
    loadWalletData()
  }, [])

  const loadWalletData = async () => {
    try {
      setLoading(true)
      const [balanceResponse, transactionsResponse] = await Promise.all([
        walletAPI.balance(),
        walletAPI.transactions({ limit: 50 })
      ])
      
      setBalance(balanceResponse.data)
      setTransactions(transactionsResponse.data.transactions || [])
    } catch (error: any) {
      console.error('Wallet data loading error:', error)
      // Load mock data for demonstration
      loadMockWalletData()
      toast.error('Using demo data - connect to backend for live data')
    } finally {
      setLoading(false)
    }
  }

  const loadMockWalletData = () => {
    const mockBalance: CreditBalance = {
      balance: 150.75,
      total_earned: 340.50,
      total_spent: 189.75,
      pending_transactions: 2
    }

    const mockTransactions: Transaction[] = [
      {
        id: 'tx-1',
        type: 'earn',
        amount: 25.00,
        description: 'Node computation reward - Job #12345',
        timestamp: '2024-01-15T14:30:00Z',
        status: 'completed',
        job_id: 'job-12345'
      },
      {
        id: 'tx-2',
        type: 'spend',
        amount: -15.50,
        description: 'Job submission - Data Processing Pipeline',
        timestamp: '2024-01-15T10:30:00Z',
        status: 'completed',
        job_id: 'job-67890'
      },
      {
        id: 'tx-3',
        type: 'transfer_in',
        amount: 50.00,
        description: 'Credit transfer from user@example.com',
        timestamp: '2024-01-15T08:15:00Z',
        status: 'completed',
        counterparty: 'user@example.com'
      },
      {
        id: 'tx-4',
        type: 'earn',
        amount: 12.75,
        description: 'Node computation reward - Job #12344',
        timestamp: '2024-01-14T16:45:00Z',
        status: 'completed',
        job_id: 'job-12344'
      },
      {
        id: 'tx-5',
        type: 'spend',
        amount: -8.00,
        description: 'Job submission - Image Processing',
        timestamp: '2024-01-14T12:20:00Z',
        status: 'failed',
        job_id: 'job-11111'
      },
      {
        id: 'tx-6',
        type: 'transfer_out',
        amount: -20.00,
        description: 'Credit transfer to node-operator@company.com',
        timestamp: '2024-01-14T09:10:00Z',
        status: 'completed',
        counterparty: 'node-operator@company.com'
      },
      {
        id: 'tx-7',
        type: 'earn',
        amount: 18.25,
        description: 'Node computation reward - Job #12343',
        timestamp: '2024-01-13T20:30:00Z',
        status: 'pending'
      }
    ]

    setBalance(mockBalance)
    setTransactions(mockTransactions)
  }

  const handleDeposit = async () => {
    try {
      const amount = parseFloat(depositAmount)
      if (isNaN(amount) || amount <= 0) {
        toast.error('Please enter a valid amount')
        return
      }

      await walletAPI.deposit(amount)
      toast.success(`Successfully deposited ${amount} credits`)
      setShowDepositModal(false)
      setDepositAmount('')
      loadWalletData()
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Deposit failed')
    }
  }

  const handleWithdraw = async () => {
    try {
      const amount = parseFloat(withdrawAmount)
      if (isNaN(amount) || amount <= 0) {
        toast.error('Please enter a valid amount')
        return
      }

      if (balance && amount > balance.balance) {
        toast.error('Insufficient balance')
        return
      }

      await walletAPI.withdraw(amount)
      toast.success(`Successfully withdrew ${amount} credits`)
      setShowWithdrawModal(false)
      setWithdrawAmount('')
      loadWalletData()
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Withdrawal failed')
    }
  }

  const handleTransfer = async () => {
    try {
      const amount = parseFloat(transferAmount)
      if (isNaN(amount) || amount <= 0) {
        toast.error('Please enter a valid amount')
        return
      }

      if (!transferUserId) {
        toast.error('Please enter a user ID')
        return
      }

      await walletAPI.transfer(transferUserId, amount)
      toast.success(`Successfully transferred ${amount} credits`)
      setShowTransferModal(false)
      setTransferAmount('')
      setTransferUserId('')
      loadWalletData()
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Transfer failed')
    }
  }

  const getTransactionIcon = (type: string) => {
    switch (type) {
      case 'earn':
        return <TrendingUp className="h-5 w-5 text-green-500" />
      case 'spend':
        return <TrendingDown className="h-5 w-5 text-red-500" />
      case 'transfer_in':
        return <ArrowDownLeft className="h-5 w-5 text-blue-500" />
      case 'transfer_out':
        return <ArrowUpRight className="h-5 w-5 text-orange-500" />
      default:
        return <CreditCard className="h-5 w-5 text-gray-500" />
    }
  }

  const getTransactionColor = (type: string) => {
    switch (type) {
      case 'earn':
      case 'transfer_in':
        return 'text-green-600'
      case 'spend':
      case 'transfer_out':
        return 'text-red-600'
      default:
        return 'text-gray-600'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800'
      case 'pending':
        return 'bg-yellow-100 text-yellow-800'
      case 'failed':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const formatDate = (dateString: string) => {
    return format(new Date(dateString), 'MMM dd, yyyy HH:mm')
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
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Wallet</h1>
        <p className="text-gray-600">
          Manage your DEparrow credits and view transaction history.
        </p>
      </div>

      {/* Balance Card */}
      <div className="bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg shadow-lg p-6 text-white">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-medium opacity-90">Current Balance</h2>
            <p className="text-4xl font-bold">{balance?.balance.toFixed(2) || '0.00'} Credits</p>
            <div className="mt-2 grid grid-cols-3 gap-4 text-sm">
              <div>
                <p className="opacity-75">Total Earned</p>
                <p className="font-semibold">{balance?.total_earned.toFixed(2) || '0.00'}</p>
              </div>
              <div>
                <p className="opacity-75">Total Spent</p>
                <p className="font-semibold">{balance?.total_spent.toFixed(2) || '0.00'}</p>
              </div>
              <div>
                <p className="opacity-75">Pending</p>
                <p className="font-semibold">{balance?.pending_transactions || 0}</p>
              </div>
            </div>
          </div>
          <div className="hidden md:block">
            <WalletIcon className="h-16 w-16 opacity-50" />
          </div>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <button
          onClick={() => setShowDepositModal(true)}
          className="flex items-center justify-center px-4 py-3 bg-green-50 border border-green-200 rounded-lg hover:bg-green-100 transition-colors"
        >
          <Plus className="h-5 w-5 text-green-600 mr-2" />
          <span className="text-green-700 font-medium">Add Credits</span>
        </button>
        
        <button
          onClick={() => setShowWithdrawModal(true)}
          className="flex items-center justify-center px-4 py-3 bg-red-50 border border-red-200 rounded-lg hover:bg-red-100 transition-colors"
        >
          <Minus className="h-5 w-5 text-red-600 mr-2" />
          <span className="text-red-700 font-medium">Withdraw</span>
        </button>
        
        <button
          onClick={() => setShowTransferModal(true)}
          className="flex items-center justify-center px-4 py-3 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 transition-colors"
        >
          <Upload className="h-5 w-5 text-blue-600 mr-2" />
          <span className="text-blue-700 font-medium">Transfer</span>
        </button>
        
        <button
          onClick={() => loadWalletData()}
          className="flex items-center justify-center px-4 py-3 bg-gray-50 border border-gray-200 rounded-lg hover:bg-gray-100 transition-colors"
        >
          <Download className="h-5 w-5 text-gray-600 mr-2" />
          <span className="text-gray-700 font-medium">Refresh</span>
        </button>
      </div>

      {/* Transaction History */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Transaction History</h3>
        </div>
        
        {transactions.length > 0 ? (
          <div className="divide-y divide-gray-200">
            {transactions.map((transaction) => (
              <div key={transaction.id} className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <div className="flex-shrink-0">
                      {getTransactionIcon(transaction.type)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {transaction.description}
                      </p>
                      <div className="flex items-center mt-1 space-x-2">
                        <p className="text-sm text-gray-500">
                          {formatDate(transaction.timestamp)}
                        </p>
                        {transaction.job_id && (
                          <span className="text-xs text-blue-600">
                            Job: {transaction.job_id}
                          </span>
                        )}
                        {transaction.counterparty && (
                          <span className="text-xs text-gray-500">
                            {transaction.type === 'transfer_in' ? 'From' : 'To'}: {transaction.counterparty}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-3">
                    <span className={`text-lg font-semibold ${getTransactionColor(transaction.type)}`}>
                      {transaction.amount > 0 ? '+' : ''}{transaction.amount.toFixed(2)}
                    </span>
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(transaction.status)}`}>
                      {transaction.status}
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-12">
            <WalletIcon className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">No transactions</h3>
            <p className="mt-1 text-sm text-gray-500">
              Your transaction history will appear here.
            </p>
          </div>
        )}
      </div>

      {/* Deposit Modal */}
      {showDepositModal && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" onClick={() => setShowDepositModal(false)}></div>
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Add Credits to Wallet
                </h3>
                <div>
                  <label htmlFor="amount" className="block text-sm font-medium text-gray-700">
                    Amount (Credits)
                  </label>
                  <div className="mt-1 relative rounded-md shadow-sm">
                    <input
                      type="number"
                      id="amount"
                      min="0"
                      step="0.01"
                      value={depositAmount}
                      onChange={(e) => setDepositAmount(e.target.value)}
                      className="block w-full pr-12 border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="0.00"
                    />
                    <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                      <span className="text-gray-500 sm:text-sm">Credits</span>
                    </div>
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={handleDeposit}
                >
                  Add Credits
                </button>
                <button
                  type="button"
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={() => setShowDepositModal(false)}
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Withdraw Modal */}
      {showWithdrawModal && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" onClick={() => setShowWithdrawModal(false)}></div>
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Withdraw Credits
                </h3>
                <div>
                  <label htmlFor="withdraw-amount" className="block text-sm font-medium text-gray-700">
                    Amount (Credits)
                  </label>
                  <div className="mt-1 relative rounded-md shadow-sm">
                    <input
                      type="number"
                      id="withdraw-amount"
                      min="0"
                      max={balance?.balance || 0}
                      step="0.01"
                      value={withdrawAmount}
                      onChange={(e) => setWithdrawAmount(e.target.value)}
                      className="block w-full pr-12 border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="0.00"
                    />
                    <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                      <span className="text-gray-500 sm:text-sm">Credits</span>
                    </div>
                  </div>
                  {balance && (
                    <p className="mt-2 text-sm text-gray-500">
                      Available balance: {balance.balance.toFixed(2)} credits
                    </p>
                  )}
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-red-600 text-base font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={handleWithdraw}
                >
                  Withdraw
                </button>
                <button
                  type="button"
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={() => setShowWithdrawModal(false)}
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Transfer Modal */}
      {showTransferModal && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" onClick={() => setShowTransferModal(false)}></div>
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Transfer Credits
                </h3>
                <div className="space-y-4">
                  <div>
                    <label htmlFor="user-id" className="block text-sm font-medium text-gray-700">
                      Recipient User ID or Email
                    </label>
                    <input
                      type="text"
                      id="user-id"
                      value={transferUserId}
                      onChange={(e) => setTransferUserId(e.target.value)}
                      className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                      placeholder="user@example.com"
                    />
                  </div>
                  <div>
                    <label htmlFor="transfer-amount" className="block text-sm font-medium text-gray-700">
                      Amount (Credits)
                    </label>
                    <div className="mt-1 relative rounded-md shadow-sm">
                      <input
                        type="number"
                        id="transfer-amount"
                        min="0"
                        max={balance?.balance || 0}
                        step="0.01"
                        value={transferAmount}
                        onChange={(e) => setTransferAmount(e.target.value)}
                        className="block w-full pr-12 border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                        placeholder="0.00"
                      />
                      <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                        <span className="text-gray-500 sm:text-sm">Credits</span>
                      </div>
                    </div>
                    {balance && (
                      <p className="mt-2 text-sm text-gray-500">
                        Available balance: {balance.balance.toFixed(2)} credits
                      </p>
                    )}
                  </div>
                </div>
              </div>
              <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={handleTransfer}
                >
                  Transfer
                </button>
                <button
                  type="button"
                  className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                  onClick={() => setShowTransferModal(false)}
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Wallet