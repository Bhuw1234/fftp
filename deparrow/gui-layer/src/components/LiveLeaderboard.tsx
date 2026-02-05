import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface LeaderboardEntry {
    rank: number;
    nodeId: string;
    tier: string;
    tierIcon: string;
    cpuHours: number;
    gpuHours: number;
    totalHours: number;
    liveGflops: number;
    creditsEarned: number;
    status: 'online' | 'offline';
    pulse: boolean;
}

const tierGradients: Record<string, string> = {
    legendary: 'from-purple-500/20 to-pink-500/20 border-purple-500/50',
    diamond: 'from-cyan-500/20 to-blue-500/20 border-cyan-500/50',
    gold: 'from-yellow-500/20 to-amber-500/20 border-yellow-500/50',
    silver: 'from-gray-400/20 to-gray-500/20 border-gray-400/50',
    bronze: 'from-amber-600/20 to-amber-700/20 border-amber-600/50',
};

export const LiveLeaderboard: React.FC = () => {
    const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        // Simulated data - in production would fetch from WebSocket
        const mockData: LeaderboardEntry[] = [
            { rank: 1, nodeId: 'node-a1b2...', tier: 'legendary', tierIcon: '🔥', cpuHours: 12500, gpuHours: 8900, totalHours: 21400, liveGflops: 892.4, creditsEarned: 45600, status: 'online', pulse: true },
            { rank: 2, nodeId: 'node-c3d4...', tier: 'diamond', tierIcon: '💎', cpuHours: 8900, gpuHours: 6200, totalHours: 15100, liveGflops: 654.2, creditsEarned: 32100, status: 'online', pulse: true },
            { rank: 3, nodeId: 'node-e5f6...', tier: 'diamond', tierIcon: '💎', cpuHours: 7600, gpuHours: 5100, totalHours: 12700, liveGflops: 521.8, creditsEarned: 28400, status: 'online', pulse: true },
            { rank: 4, nodeId: 'node-g7h8...', tier: 'gold', tierIcon: '🥇', cpuHours: 4200, gpuHours: 2800, totalHours: 7000, liveGflops: 312.5, creditsEarned: 15200, status: 'online', pulse: true },
            { rank: 5, nodeId: 'node-i9j0...', tier: 'gold', tierIcon: '🥇', cpuHours: 3800, gpuHours: 2100, totalHours: 5900, liveGflops: 245.6, creditsEarned: 12800, status: 'online', pulse: true },
            { rank: 6, nodeId: 'node-k1l2...', tier: 'silver', tierIcon: '🥈', cpuHours: 1200, gpuHours: 800, totalHours: 2000, liveGflops: 98.3, creditsEarned: 4500, status: 'offline', pulse: false },
            { rank: 7, nodeId: 'node-m3n4...', tier: 'silver', tierIcon: '🥈', cpuHours: 980, gpuHours: 650, totalHours: 1630, liveGflops: 78.2, creditsEarned: 3800, status: 'online', pulse: true },
            { rank: 8, nodeId: 'node-o5p6...', tier: 'bronze', tierIcon: '🥉', cpuHours: 320, gpuHours: 180, totalHours: 500, liveGflops: 42.1, creditsEarned: 1200, status: 'online', pulse: true },
        ];

        setTimeout(() => {
            setEntries(mockData);
            setIsLoading(false);
        }, 500);
    }, []);

    return (
        <div className="p-6 rounded-3xl bg-gradient-to-br from-gray-900/90 to-gray-800/90 backdrop-blur-xl border border-white/10">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                    <span className="text-3xl">🏆</span>
                    <div>
                        <h2 className="text-2xl font-bold text-white">Live Leaderboard</h2>
                        <p className="text-sm text-gray-400">Top contributors to DEparrow</p>
                    </div>
                </div>
                <motion.div
                    className="flex items-center gap-2 px-3 py-1 rounded-full bg-green-500/20 border border-green-500/50"
                    animate={{ opacity: [1, 0.5, 1] }}
                    transition={{ duration: 2, repeat: Infinity }}
                >
                    <div className="w-2 h-2 rounded-full bg-green-500" />
                    <span className="text-xs text-green-400 font-mono">LIVE</span>
                </motion.div>
            </div>

            {/* Table Header */}
            <div className="grid grid-cols-7 gap-4 px-4 py-2 text-xs text-gray-500 uppercase tracking-wider border-b border-white/10">
                <div>Rank</div>
                <div className="col-span-2">Node</div>
                <div className="text-right">Hours</div>
                <div className="text-right">GFLOPS</div>
                <div className="text-right">Credits</div>
                <div className="text-center">Status</div>
            </div>

            {/* Entries */}
            <AnimatePresence>
                {entries.map((entry, index) => (
                    <motion.div
                        key={entry.nodeId}
                        initial={{ opacity: 0, x: -20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: index * 0.1 }}
                        className={`grid grid-cols-7 gap-4 px-4 py-3 items-center border-b border-white/5 hover:bg-white/5 transition-colors bg-gradient-to-r ${tierGradients[entry.tier]} border-l-2`}
                    >
                        {/* Rank */}
                        <div className="flex items-center gap-2">
                            <span className="text-lg font-bold text-white">#{entry.rank}</span>
                        </div>

                        {/* Node */}
                        <div className="col-span-2 flex items-center gap-2">
                            <span className="text-2xl">{entry.tierIcon}</span>
                            <div>
                                <div className="font-mono text-sm text-white">{entry.nodeId}</div>
                                <div className="text-xs text-gray-500 uppercase">{entry.tier}</div>
                            </div>
                        </div>

                        {/* Hours */}
                        <div className="text-right">
                            <div className="font-bold text-white">{entry.totalHours.toLocaleString()}h</div>
                            <div className="text-xs text-gray-500">
                                CPU: {entry.cpuHours.toLocaleString()} | GPU: {entry.gpuHours.toLocaleString()}
                            </div>
                        </div>

                        {/* GFLOPS */}
                        <div className="text-right">
                            <motion.span
                                className="font-bold text-cyan-400"
                                animate={entry.pulse ? { opacity: [1, 0.7, 1] } : {}}
                                transition={{ duration: 1, repeat: Infinity }}
                            >
                                {entry.liveGflops}
                            </motion.span>
                        </div>

                        {/* Credits */}
                        <div className="text-right font-bold text-yellow-400">
                            {entry.creditsEarned.toLocaleString()}
                        </div>

                        {/* Status */}
                        <div className="text-center">
                            <motion.div
                                className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs ${entry.status === 'online'
                                        ? 'bg-green-500/20 text-green-400'
                                        : 'bg-gray-500/20 text-gray-400'
                                    }`}
                                animate={entry.pulse ? { scale: [1, 1.05, 1] } : {}}
                                transition={{ duration: 2, repeat: Infinity }}
                            >
                                <div
                                    className={`w-1.5 h-1.5 rounded-full ${entry.status === 'online' ? 'bg-green-500' : 'bg-gray-500'
                                        }`}
                                />
                                {entry.status}
                            </motion.div>
                        </div>
                    </motion.div>
                ))}
            </AnimatePresence>

            {/* Total Stats */}
            <div className="mt-4 pt-4 border-t border-white/10 flex justify-between text-sm text-gray-400">
                <span>📊 {entries.length} contributors shown</span>
                <span>🌐 127 total nodes in network</span>
            </div>
        </div>
    );
};

export default LiveLeaderboard;
