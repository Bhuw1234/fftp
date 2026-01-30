import { motion } from 'framer-motion'
import AICore from '../canvas/AICore'
import { ArrowDown } from 'lucide-react'

export default function Hero() {
    return (
        <section className="relative h-screen w-full flex items-center justify-center overflow-hidden bg-brand-dark">
            {/* 3D Background */}
            <AICore />

            {/* Overlay Content */}
            <div className="relative z-10 text-center px-4 w-full max-w-5xl mx-auto pointer-events-none">
                <motion.h1
                    initial={{ opacity: 0, y: 30 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 1, ease: [0.22, 1, 0.36, 1] }}
                    className="text-6xl md:text-8xl lg:text-9xl font-extrabold tracking-tighter text-transparent bg-clip-text bg-gradient-to-r from-brand-glow via-white to-brand-accent drop-shadow-[0_0_30px_rgba(0,255,255,0.3)]"
                >
                    100xSolution
                </motion.h1>

                <motion.p
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.8, delay: 0.3, ease: "easeOut" }}
                    className="mt-6 text-xl md:text-2xl text-brand-text font-light tracking-wide max-w-2xl mx-auto leading-relaxed"
                >
                    AI Marketing That Thinks For You
                </motion.p>

                <motion.div
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ duration: 0.8, delay: 0.6 }}
                    className="mt-12 pointer-events-auto inline-block"
                >
                    <button className="group relative px-8 py-4 bg-transparent overflow-hidden rounded-full transition-all duration-300 hover:scale-105">
                        <div className="absolute inset-0 bg-brand-accent/20 backdrop-blur-md border border-brand-accent/50 rounded-full group-hover:bg-brand-accent/30 transition-all duration-300 shadow-[0_0_20px_rgba(124,58,237,0.3)]"></div>
                        <span className="relative z-10 text-brand-glow font-semibold tracking-wider group-hover:text-white transition-colors">
                            EXPLORE INTELLIGENCE
                        </span>
                    </button>
                </motion.div>
            </div>

            {/* Scroll indicator */}
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1, y: [0, 10, 0] }}
                transition={{ delay: 1.5, duration: 2, repeat: Infinity }}
                className="absolute bottom-10 left-1/2 transform -translate-x-1/2 text-brand-text/50 flex flex-col items-center gap-2 pointer-events-none"
            >
                <span className="text-xs uppercase tracking-widest">Scroll</span>
                <ArrowDown className="w-5 h-5 text-brand-glow/70" />
            </motion.div>
        </section>
    )
}
