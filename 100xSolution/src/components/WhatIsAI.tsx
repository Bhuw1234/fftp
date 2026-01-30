import { motion } from 'framer-motion'

export default function WhatIsAI() {
    return (
        <section className="py-32 px-6 relative overflow-hidden bg-brand-dark">
            {/* Background Gradient Blob */}
            <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-brand-accent/5 rounded-full blur-[120px] pointer-events-none" />

            <div className="max-w-4xl mx-auto text-center relative z-10">
                <motion.h2
                    initial={{ opacity: 0, y: 30 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true, margin: "-100px" }}
                    transition={{ duration: 0.7 }}
                    className="text-4xl md:text-6xl font-bold mb-10 text-white"
                >
                    What is <span className="text-brand-glow drop-shadow-[0_0_15px_rgba(0,255,255,0.4)]">AI Marketing</span>?
                </motion.h2>

                <motion.div
                    initial={{ opacity: 0, y: 30 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true, margin: "-100px" }}
                    transition={{ delay: 0.2, duration: 0.7 }}
                    className="space-y-8 text-xl md:text-2xl text-brand-text/90 leading-relaxed font-light"
                >
                    <p>
                        AI Marketing uses smart technology to understand what customers like, predict their behavior, and show them the right message at the right time.
                    </p>
                    <p className="border-l-4 border-brand-accent pl-6 italic text-brand-text/70">
                        "It’s like having a <span className="text-brand-accent font-semibold not-italic">digital brain</span> that studies your customers and helps your business grow faster."
                    </p>
                </motion.div>
            </div>
        </section>
    )
}
