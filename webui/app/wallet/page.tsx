'use client'

import React, { Suspense } from 'react'
import { WalletOverview } from '@/components/wallet'

function WalletContent() {
  return <WalletOverview />
}

export default function WalletPage() {
  return (
    <Suspense fallback={<div className="flex items-center justify-center min-h-[400px]"><div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div></div>}>
      <WalletContent />
    </Suspense>
  )
}
