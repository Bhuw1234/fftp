import DataFlow from '../canvas/DataFlow'
import { motion } from 'framer-motion'

export default function HowItWorks() {
    return (
        <section className="py-24 bg-brand-dark/50 relative z-10 border-t border-white/5">
            <div className="text-center mb-8 px-4">
                <motion.h2
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    className="text-3xl md:text-5xl font-bold text-white mb-4"
                >
                    How It Works
                </motion.h2>
                <motion.p
                    initial={{ opacity: 0 }}
                    whileInView={{ opacity: 1 }}
                    className="text-brand-text/70 text-lg"
                >
                    Simple. Automatic. Powerful.
                </motion.p>
            </div>

            <motion.div
                initial={{ opacity: 0 }}
                whileInView={{ opacity: 1 }}
                transition={{ duration: 1 }}
            >
                <DataFlow />
            </motion.div>

            <div className="max-w-6xl mx-auto px-6 grid grid-cols-1 md:grid-cols-3 gap-12 text-center -mt-10 relative z-20">
                <motion.div
                    whileHover={{ y: -5 }}
                    className="bg-brand-dark/80 p-6 rounded-2xl border border-white/10 backdrop-blur-sm"
                >
                    <h3 className="text-2xl font-semibold text-brand-glow mb-2">1. Collects Data</h3>
                    <p className="text-sm text-brand-text/60 leading-relaxed">
                        First, AI automatically aggregates customer data from your website and social channels securely.
                    </p>
                </motion.div>

                <motion.div
                    whileHover={{ y: -5 }}
                    className="bg-brand-dark/80 p-6 rounded-2xl border border-brand-accent/20 backdrop-blur-sm shadow-[0_0_20px_rgba(124,58,237,0.1)]"
                >
                    <h3 className="text-2xl font-semibold text-brand-accent mb-2">2. Studies Patterns</h3>
                    <p className="text-sm text-brand-text/60 leading-relaxed">
                        Then, our Deep Learning engine studies behaviors to identify high-value purchasing trends.
                    </p>
                </motion.div>

                <motion.div
                    whileHover={{ y: -5 }}
                    className="bg-brand-dark/80 p-6 rounded-2xl border border-white/10 backdrop-blur-sm"
                >
                    <h3 className="text-2xl font-semibold text-white mb-2">3. Personalizes Ads</h3>
                    <p className="text-sm text-brand-text/60 leading-relaxed">
                        Finally, it deploys hyper-personalized campaigns that convert visitors into buyers 24/7.
                    </p>
                </motion.div>
            </div>
        </section>
    )
}
