# DEparrow Social App - Prototype Design Document

**Created:** 2026-03-26
**Status:** Design Complete, Ready for Development

---

## Executive Summary

**Concept:** A social media platform where users earn DPC tokens by using the app while their device contributes compute power to the DEparrow network in the background.

**Core Positioning:** "Facebook pays shareholders. We pay YOU."

**Key Insight:** The average smartphone sits idle 80% of the time while charging. That's wasted compute. We monetize it for the user.

---

## 1. User Experience Flow

### 1.1 Onboarding Process (Target: < 60 seconds)

```
[Download] → [Phone Verify] → [Create Username] → [Wallet Created] → [One Toggle] → [EARN MODE: ON]
     │              │                │                  │                │              │
     ▼              ▼                ▼                  ▼                ▼              ▼
 App Store    SMS Code (30s)    Unique @handle    Auto-generated    "Enable Compute    User is now
                                                      (User sees      Contribution?"   earning DPC
                                                      nothing yet)    [Yes/No]          immediately!

TOTAL TIME: 45-60 seconds
FRICTION: Minimal (no crypto knowledge required)
```

### 1.2 Main Screens

#### Feed (Home Screen)
```
┌─────────────────────────────────────────────────────────────────┐
│  ≡  DEparrow                          💰 12.45 DPC  ⚡ ON     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  @alice • 2h ago                    [Earning: +0.03 DPC]│   │
│  │                                                         │   │
│  │  Just earned 5 DPC while sleeping! 🎉 The future is    │   │
│  │  here - getting paid to use social media.              │   │
│  │                                                         │   │
│  │  💬 23   🔁 12   ❤️ 45   🪙 Tip                         │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  [🏠 Feed] [💬 Chat] [➕ Post] [📊 Stats] [👤 Profile]        │
└─────────────────────────────────────────────────────────────────┘
```

#### Earnings Dashboard
```
┌─────────────────────────────────────────────────────────────────┐
│  ← Earnings Dashboard                                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│              🪙 12.45 DPC                                       │
│              ≈ $1.24 USD                                        │
│                                                                 │
│         [Withdraw]    [Use for Compute]                        │
│                                                                 │
│  TODAY                      THIS WEEK                          │
│  +0.82 DPC                  +5.41 DPC                          │
│  ▓▓▓▓▓▓░░░░ 65%            ▓▓▓▓▓▓▓▓░░ 78%                     │
│                                                                 │
│  COMPUTE CONTRIBUTION                                          │
│  Jobs Completed: 127                                           │
│  CPU Hours: 3.2                                                │
│  Active Time: 14h 23m                                          │
│  Efficiency: ████████░░ 82%                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. Technical Architecture

### 2.1 System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    MOBILE CLIENT                                │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│   │ React       │  │ Bacalhau    │  │ Embedded    │            │
│   │ Native UI   │  │ Lite Engine │  │ Wallet      │            │
│   │             │  │ (WASM)      │  │ (Keychain)  │            │
│   └─────────────┘  └─────────────┘  └─────────────┘            │
│                          │                                      │
│                   ┌──────┴──────┐                               │
│                   │  Background │                               │
│                   │  Worker     │                               │
│                   │  (Compute)  │                               │
│                   └─────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    BACKEND SERVICES                             │
│   ┌───────────┐  ┌───────────┐  ┌───────────┐                 │
│   │ API       │  │ Job       │  │ Reward    │                 │
│   │ Gateway   │  │ Scheduler │  │ Service   │                 │
│   └───────────┘  └───────────┘  └───────────┘                 │
│   ┌───────────┐  ┌───────────┐  ┌───────────┐                 │
│   │ Feed      │  │ Chat      │  │ DPC       │                 │
│   │ Service   │  │ Service   │  │ Blockchain│                 │
│   └───────────┘  └───────────┘  └───────────┘                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    DEPARROW NETWORK                             │
│   Bootstrap Server (34.180.51.11:8080)                         │
│   Orchestrator Nodes                                            │
│   Compute Nodes (Phones + Servers)                             │
│   NATS Message Bus                                              │
│   IPFS Content Storage                                          │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Mobile App Stack

| Layer | Technology | Rationale |
|-------|------------|-----------|
| **Framework** | React Native 0.73+ | Cross-platform, Expo support |
| **Language** | TypeScript 5.x | Type safety |
| **State** | Redux Toolkit + RTK Query | Predictable state, caching |
| **Navigation** | React Navigation 6 | Standard, performant |
| **UI Library** | Tamagui or NativeWind | Styled components |
| **Storage** | SQLite + MMKV | Fast local storage |
| **Crypto** | ethers.js + WalletConnect SDK | Web3 standard |
| **Secure Storage** | iOS Keychain / Android Keystore | Key protection |

### 2.3 Embedded Bacalhau Lite Engine

**Constraints:**
- Max CPU: 10% (configurable)
- Max RAM: 100MB
- Max Battery Drain: 2%/hour
- Runs only when: Charging OR Battery > 50%
- Network: WiFi preferred, cellular optional

**Execution:**
- WASM Runtime: WasmEdge (lightweight, ~3MB)
- Job Types: ML inference, data processing, WASM tasks
- No Docker (too heavy)
- Sandbox: Isolated process

### 2.4 Compute Modes

| Mode | CPU | RAM | Battery | Network | Est. Earnings |
|------|-----|-----|---------|---------|---------------|
| **Power User** | 10% | 100MB | ≤2%/hr | WiFi only | 0.8-2.5 DPC/day |
| **Balanced** | 5% | 50MB | ≤1%/hr | WiFi+Cell | 0.4-1.2 DPC/day |
| **Eco** | 2% | 20MB | Charging only | WiFi only | 0.1-0.5 DPC/day |
| **Off** | 0% | 0MB | None | None | 0 DPC/day |

---

## 3. Monetization Model

### 3.1 DPC Earning Formula

```
DPC_per_job = Base_Rate × Job_Complexity × Time_Spent

WHERE:
├── Base_Rate = 0.001 DPC (minimum per job)
├── Job_Complexity = 1x (simple) to 5x (complex ML)
└── Time_Spent = seconds × 0.0001
```

### 3.2 Example Jobs

| Job Type | Duration | Complexity | DPC Earned |
|----------|----------|------------|------------|
| ML Inference | 30 sec | 3x | 0.09 DPC |
| Data Processing | 60 sec | 2x | 0.12 DPC |
| WASM Compute | 15 sec | 1x | 0.015 DPC |
| Light Task | 5 sec | 1x | 0.005 DPC |

### 3.3 Reward Distribution

```
┌─────────────────────────────────────────────────────────────────┐
│   User Wallet: 70% of job reward                               │
│   Platform Treasury: 20% (development, ops)                    │
│   Network Fee: 10% (validator rewards)                         │
└─────────────────────────────────────────────────────────────────┘
```

### 3.4 Bonuses

- **Referral:** +5 DPC when referred user earns first DPC
- **Streak:** +10% bonus for 7-day consecutive activity
- **Power User:** +15% bonus when in Power mode
- **Early Adopter:** +25% bonus for first 100K users

### 3.5 Platform Economics

| Metric | Conservative | Moderate | Optimistic |
|--------|--------------|----------|------------|
| **Users (Year 1)** | 50,000 | 500,000 | 5,000,000 |
| **DPC/User/Day** | 0.3 | 0.5 | 0.8 |
| **Total DPC Distributed** | 5.5M | 91M | 1.46B |
| **USD Value (@$0.10/DPC)** | $550K | $9.1M | $146M |
| **Platform Revenue (20%)** | $110K | $1.82M | $29.2M |

---

## 4. MVP Features

### 4.1 Phase 1: Foundation (4-6 weeks)

| Feature | Priority | Description |
|---------|----------|-------------|
| User Registration | P0 | Email/phone signup, username picker |
| Basic Feed | P0 | Text-only posts, infinite scroll |
| Post Creation | P0 | Create text post, character limit |
| Embedded Wallet | P0 | Auto-generated, secure storage |
| Background Compute | P0 | Bacalhau Lite integration, toggle on/off |
| Earnings Display | P0 | Real-time DPC counter in header |
| Compute Dashboard | P0 | Basic stats: jobs completed, DPC earned |
| DPC Balance | P0 | Current balance, pending, total earned |

### 4.2 Phase 2: Engagement (6-8 weeks)

| Feature | Priority | Description |
|---------|----------|-------------|
| Comments & Likes | P0 | Interact with posts |
| Images in Posts | P0 | Photo upload, IPFS storage |
| Direct Messaging | P1 | 1:1 chat, real-time |
| Profile Pages | P1 | Avatar, bio, stats |
| Push Notifications | P1 | Earnings alerts, messages |
| Withdraw DPC | P1 | Send to external wallet |
| Transaction History | P1 | Filterable list |
| Follow/Following | P2 | Social graph |
| Trending Page | P2 | Popular posts discovery |

### 4.3 Phase 3: Scale (8-12 weeks)

| Feature | Priority | Description |
|---------|----------|-------------|
| Groups | P1 | Topic-based communities |
| Events | P2 | Time-based activities |
| Advanced Compute Settings | P1 | Job type preferences |
| Video Posts | P1 | Short-form video, IPFS |
| DPC Marketplace | P2 | Spend DPC on services |
| Staking | P2 | Lock DPC for higher rewards |
| Tipping | P1 | Send DPC to other users |

### 4.4 Development Roadmap

```
2026 Q2 (Phase 1)         2026 Q3 (Phase 2)
├── User Registration     ├── Comments & Likes
├── Basic Feed            ├── Images
├── Post Creation         ├── Direct Messaging
├── Embedded Wallet       ├── Profile Pages
├── Background Compute    ├── Push Notifications
├── Earnings Display      ├── Withdraw DPC
└── Compute Dashboard     └── Follow System

2026 Q4 (Phase 3)
├── Groups & Events
├── Video Posts
├── DPC Marketplace
├── Staking
└── Tipping

MILESTONES:
├── Week 6: Phase 1 Alpha (internal testing)
├── Week 12: Phase 1 Beta (limited public)
├── Week 20: Phase 2 Launch (App Store)
└── Week 32: Phase 3 Scale (growth features)
```

---

## 5. Competitive Analysis

### 5.1 vs Traditional Social Media

| Aspect | Facebook/Twitter/TikTok | DEparrow Social |
|--------|-------------------------|-----------------|
| **Revenue Model** | User = Product | User = Partner |
| **Data Privacy** | Extensive tracking | Minimal tracking |
| **User Compensation** | None | DPC tokens |
| **Content Algorithm** | Opaque | Transparent |
| **Ads** | Intrusive, targeted | Optional, rewarded |
| **Lock-in** | Network effect | Wallet-based identity |

### 5.2 vs Other Earn-to-Use Apps

| App | Model | Earnings |
|-----|-------|----------|
| **Brave Browser** | BAT tokens for ads | $2-5/month |
| **Honeygain** | Sell bandwidth | $1-10/month |
| **Sweatcoin** | Steps → tokens | $0.50-2/month |
| **Pi Network** | Mobile mining | Uncertain |
| **DEparrow Social** | Compute + social | $3-25/month* |

*Based on DPC at $0.10

### 5.3 Positioning

**FOR:** Gen Z and Millennial users who are crypto-curious

**DEPARROW SOCIAL IS:** The first social platform that pays you to exist

**THAT:** Turns your phone's idle compute into real money

**UNLIKE:** Facebook, Twitter, TikTok that profit from your attention

### 5.4 Target Market

| Segment | Size | Motivation |
|---------|------|------------|
| **Crypto-Curious** | 100M | Want utility, not speculation |
| **Privacy-Conscious** | 50M | Distrust Big Tech |
| **Tech-Savvy** | 50M | Understand decentralized compute |
| **Passive Income Seekers** | 200M | Want money from nothing |

### 5.5 Go-to-Market Strategy

| Phase | Tactic | Goal |
|-------|--------|------|
| **Pre-Launch** | Crypto influencer partnerships | Build waitlist 100K+ |
| **Launch** | "First 10K users get 100 DPC bonus" | Drive early adoption |
| **Growth** | Referral program (5 DPC per referral) | Viral coefficient > 1.5 |
| **Scale** | Cross-promote with DEparrow compute users | Network effect |

**Launch Message:** "The app that mines while you mind your business."

---

## 6. Technical Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Battery drain complaints | High | High | Conservative defaults, clear opt-in |
| App Store rejection | Medium | Critical | Emphasize "voluntary compute" |
| Low job availability | Medium | Medium | Partner with compute providers |
| Security vulnerabilities | Medium | Critical | Security audit, bug bounty |
| Regulatory uncertainty | Low | High | Utility token design |

---

## 7. Success Metrics

### 7.1 KPIs

| Metric | Target (Year 1) |
|--------|-----------------|
| **Downloads** | 500,000 |
| **MAU** | 100,000 |
| **DAU** | 25,000 |
| **Compute Node Uptime** | 80% of eligible time |
| **DPC Distributed** | 10M DPC |
| **Avg. Earnings/User** | 0.5 DPC/day |
| **Withdrawal Rate** | < 20% (rest stays in ecosystem) |
| **App Store Rating** | > 4.5 |

### 7.2 Monitoring

- Real-time earnings dashboard
- Compute job success rate
- Battery impact metrics
- User retention cohorts
- DPC flow (earn → spend → withdraw)

---

## 8. Next Steps

1. **Immediate:**
   - Set up React Native project
   - Design UI/UX mockups in Figma
   - Implement wallet generation

2. **Week 1-2:**
   - Core UI components
   - Authentication flow
   - Basic feed

3. **Week 3-4:**
   - Bacalhau Lite integration
   - Background compute worker
   - Earnings display

4. **Week 5-6:**
   - Internal testing
   - Bug fixes
   - Phase 1 Alpha release

---

*Document created: 2026-03-26*
