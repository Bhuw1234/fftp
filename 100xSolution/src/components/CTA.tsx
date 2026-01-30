import { motion } from 'framer-motion'

export default function CTA() {
    return (
        <section className="py-32 px-6 relative bg-brand-dark overflow-hidden flex flex-col items-center justify-center text-center border-t border-white/5">
            {/* Background Particles/Glow */}
            <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[120%] h-[120%] bg-radial-gradient from-brand-accent/10 to-transparent opacity-50 blur-[100px] pointer-events-none" />

            <div className="relative z-10 max-w-4xl mx-auto">
                <motion.div
                    initial={{ opacity: 0, scale: 0.9 }}
                    whileInView={{ opacity: 1, scale: 1 }}
                    viewport={{ once: true }}
                    className="mb-8 inline-block px-4 py-1 rounded-full bg-brand-glow/10 border border-brand-glow/20 text-brand-glow text-sm font-semibold tracking-wider uppercase"
                >
                    Ready to scale?
                </motion.div>

                <motion.h2
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    className="text-5xl md:text-7xl font-bold mb-8 text-white tracking-tight leading-tight"
                >
                    Let AI Grow Your Business <br />
                    <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-glow to-brand-accent drop-shadow-[0_0_20px_rgba(0,255,255,0.4)]">
                        24/7
                    </span>
                </motion.h2>

                <motion.p
                    initial={{ opacity: 0 }}
                    whileInView={{ opacity: 1 }}
                    transition={{ delay: 0.2 }}
                    viewport={{ once: true }}
                    className="text-xl md:text-2xl text-brand-text/80 mb-12 max-w-2xl mx-auto"
                >
                    Stop guessing. Start using smart marketing powered by AI.
                </motion.p>

                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.4 }}
                    viewport={{ once: true }}
                    className="flex flex-col md:flex-row gap-6 justify-center items-center"
                >
                    <button className="px-10 py-5 bg-brand-accent hover:bg-brand-accent/80 text-white font-bold text-lg rounded-full shadow-[0_0_30px_rgba(124,58,237,0.5)] transition-all transform hover:scale-105 hover:-translate-y-1">
                        Get Started Now
                    </button>
                    <button className="px-10 py-5 bg-transparent border border-white/20 hover:bg-white/10 hover:border-white/40 text-white font-semibold text-lg rounded-full transition-all hover:scale-105">
                        Book a Free Stategy Call
                    </button>
                </motion.div>
            </div>

            <footer className="absolute bottom-6 text-brand-text/20 text-sm w-full text-center">
                © {new Date().getFullYear()} 100xSolution. All rights reserved.
            </footer>
        </section>
    )
}
