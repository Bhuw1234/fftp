import Hero from './components/Hero'
import WhatIsAI from './components/WhatIsAI'
import HowItWorks from './components/HowItWorks'
import Benefits from './components/Benefits'
import CTA from './components/CTA'

// Placeholder components if they don't exist yet, but I will create them next.
// Actually, I'll allow imports and fail if file missing, forcing me to create them.
// But better to define them now or stub them.
// I'll create the files immediately after.

function App() {
  return (
    <main className="w-full min-h-screen bg-brand-dark text-brand-text selection:bg-brand-accent/30 selection:text-brand-glow">
      <Hero />
      <WhatIsAI />
      <HowItWorks />
      <Benefits />
      <CTA />
    </main>
  )
}

export default App
