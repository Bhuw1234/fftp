import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface ContributionData {
    cpu: { percent: number; cores: number; hours: number };
    gpu: { percent: number; count: number; model: string; hours: number };
    ram: { percent: number; gb: number };
    liveGflops: number;
    rank: number;
    totalNodes: number;
    tier: string;
    tierIcon: string;
    creditsEarned: number;
    pulse: boolean;
}

const tierColors: Record<string, string> = {
    bronze: 'from-amber-600 to-amber-800',
    silver: 'from-gray-300 to-gray-500',
    gold: 'from-yellow-400 to-yellow-600',
    diamond: 'from-cyan-400 to-blue-600',
    legendary: 'from-purple-500 to-pink-500',
};

const ProgressRing: React.FC<{
    percent: number;
    label: string;
    color: string;
    size?: number;
}> = ({ percent, label, color, size = 120 }) => {
    const strokeWidth = 8;
    const radius = (size - strokeWidth) / 2;
    const circumference = radius * 2 * Math.PI;
    const offset = circumference - (percent / 100) * circumference;

    return (
        <div className="flex flex-col items-center">
            <svg width={size} height={size} className="transform -rotate-90">
                {/* Background circle */}
                <circle
                    cx={size / 2}
                    cy={size / 2}
                    r={radius}
                    fill="transparent"
                    stroke="rgba(255,255,255,0.1)"
                    strokeWidth={strokeWidth}
                />
                {/* Progress circle */}
                <motion.circle
                    cx={size / 2}
                    cy={size / 2}
                    r={radius}
                    fill="transparent"
                    stroke={color}
                    strokeWidth={strokeWidth}
                    strokeLinecap="round"
                    strokeDasharray={circumference}
                    initial={{ strokeDashoffset: circumference }}
                    animate={{ strokeDashoffset: offset }}
                    transition={{ duration: 1.5, ease: 'easeOut' }}
                />
            </svg>
            <motion.div
                className="absolute flex flex-col items-center justify-center"
                style={{ width: size, height: size }}
                initial={{ opacity: 0, scale: 0.5 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.5 }}
            >
                <span className="text-2xl font-bold text-white">{percent.toFixed(1)}%</span>
                <span className="text-xs text-gray-400 uppercase">{label}</span>
            </motion.div>
        </div>
    );
};

export const ContributionRings: React.FC<{ data?: ContributionData }> = ({ data }) => {
    const [contribution, setContribution] = useState<ContributionData | null>(data || null);
    const [isLive, setIsLive] = useState(false);

    useEffect(() => {
        // Simulate real-time updates
        const interval = setInterval(() => {
            setIsLive((prev) => !prev);
        }, 1000);
        return () => clearInterval(interval);
    }, []);

    // Demo data if none provided
    const displayData = contribution || {
        cpu: { percent: 12.5, cores: 8, hours: 156.3 },
        gpu: { percent: 8.3, count: 1, model: 'RTX 4090', hours: 89.2 },
        ram: { percent: 5.2, gb: 32 },
        liveGflops: 245.6,
        rank: 5,
        totalNodes: 127,
        tier: 'gold',
        tierIcon: '🥇',
        creditsEarned: 1256,
        pulse: true,
    };

    return (
        <div className="relative p-8 rounded-3xl bg-gradient-to-br from-gray-900/80 to-gray-800/80 backdrop-blur-xl border border-white/10">
            {/* Live indicator */}
            <div className="absolute top-4 right-4 flex items-center gap-2">
                <motion.div
                    className="w-2 h-2 rounded-full bg-green-500"
                    animate={{ opacity: isLive ? 1 : 0.3, scale: isLive ? 1.2 : 1 }}
                    transition={{ duration: 0.5 }}
                />
                <span className="text-xs text-green-400 font-mono">LIVE</span>
            </div>

            {/* Header */}
            <div className="text-center mb-8">
                <motion.h2
                    className="text-3xl font-bold bg-gradient-to-r from-cyan-400 to-purple-500 bg-clip-text text-transparent"
                    initial={{ opacity: 0, y: -20 }}
                    animate={{ opacity: 1, y: 0 }}
                >
                    Your Contribution
                </motion.h2>
                <p className="text-gray-400 mt-2">Real-time compute metrics</p>
            </div>

            {/* Progress Rings */}
            <div className="flex justify-center gap-8 mb-8">
                <ProgressRing percent={displayData.cpu.percent} label="CPU" color="#06b6d4" />
                <ProgressRing percent={displayData.gpu.percent} label="GPU" color="#8b5cf6" />
                <ProgressRing percent={displayData.ram.percent} label="RAM" color="#10b981" />
            </div>

            {/* Stats Grid */}
            <div className="grid grid-cols-2 gap-4 mb-6">
                <motion.div
                    className="p-4 rounded-2xl bg-white/5 backdrop-blur border border-white/10"
                    whileHover={{ scale: 1.02, borderColor: 'rgba(6, 182, 212, 0.5)' }}
                >
                    <div className="text-sm text-gray-400">Live Compute</div>
                    <div className="text-2xl font-bold text-cyan-400">
                        <motion.span
                            key={displayData.liveGflops}
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                        >
                            {displayData.liveGflops}
                        </motion.span>
                        <span className="text-sm ml-1">GFLOPS</span>
                    </div>
                </motion.div>

                <motion.div
                    className="p-4 rounded-2xl bg-white/5 backdrop-blur border border-white/10"
                    whileHover={{ scale: 1.02, borderColor: 'rgba(139, 92, 246, 0.5)' }}
                >
                    <div className="text-sm text-gray-400">Network Rank</div>
                    <div className="text-2xl font-bold text-purple-400">
                        #{displayData.rank}
                        <span className="text-sm text-gray-500 ml-1">of {displayData.totalNodes}</span>
                    </div>
                </motion.div>
            </div>

            {/* Tier Badge */}
            <motion.div
                className={`p-4 rounded-2xl bg-gradient-to-r ${tierColors[displayData.tier]} text-center`}
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                whileHover={{ scale: 1.02 }}
            >
                <div className="text-4xl mb-2">{displayData.tierIcon}</div>
                <div className="text-xl font-bold uppercase tracking-wider">
                    {displayData.tier} TIER
                </div>
                <div className="text-sm opacity-80">
                    💰 {displayData.creditsEarned.toLocaleString()} credits earned
                </div>
            </motion.div>
        </div>
    );
};

export default ContributionRings;
