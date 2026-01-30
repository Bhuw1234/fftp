import { motion } from 'framer-motion'
import { Zap, TrendingUp, Clock, Target } from 'lucide-react'

const benefits = [
    { icon: Target, title: "Precision Targeting", desc: "Finds the right customers faster than humanly possible." },
    { icon: Zap, title: "Automated Ads", desc: "Shows personalized ads automatically. Reduces marketing costs." },
    { icon: TrendingUp, title: "Higher Sales", desc: "Increases sales and conversions with data-driven timing." },
    { icon: Clock, title: "24/7 Operation", desc: "Works 24/7 without stopping, sleeping, or taking breaks." },
]

export default function Benefits() {
    return (
        <section className="py-32 px-6 bg-brand-dark">
            <div className="max-w-6xl mx-auto">
                <motion.h2
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    className="text-4xl md:text-5xl lg:text-6xl font-bold text-center mb-20 text-white"
                >
                    Why Choose <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-glow to-brand-accent">AI Marketing</span>?
                </motion.h2>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                    {benefits.map((b, i) => (
                        <motion.div
                            key={i}
                            initial={{ opacity: 0, y: 30 }}
                            whileInView={{ opacity: 1, y: 0 }}
                            viewport={{ once: true }}
                            transition={{ delay: i * 0.1 }}
                            whileHover={{ scale: 1.02 }}
                            className="group p-10 rounded-3xl bg-white/5 border border-white/10 hover:border-brand-accent/50 hover:bg-white/10 transition-all duration-300 relative overflow-hidden cursor-default"
                        >
                            <div className="absolute inset-0 bg-gradient-to-br from-brand-accent/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />

                            <div className="flex flex-col md:flex-row items-start gap-6 relative z-10">
                                <div className="p-4 rounded-2xl bg-brand-accent/10 border border-brand-accent/20 text-brand-accent group-hover:scale-110 group-hover:bg-brand-accent group-hover:text-white transition-all duration-300 shadow-[0_0_15px_rgba(124,58,237,0.2)]">
                                    <b.icon size={32} />
                                </div>
                                <div>
                                    <h3 className="text-2xl font-bold mb-3 text-white group-hover:text-brand-glow transition-colors">{b.title}</h3>
                                    <p className="text-brand-text/70 text-lg leading-relaxed group-hover:text-brand-text/90 transition-colors">{b.desc}</p>
                                </div>
                            </div>
                        </motion.div>
                    ))}
                </div>
            </div>
        </section>
    )
}
