# 🎨 Video Storyboard - Visual Guide

## Overview
This document provides visual descriptions for each scene in the explainer video.

---

## Scene-by-Scene Visual Breakdown

### SCENE 1: Title Card (0:00 - 0:10)
```
┌─────────────────────────────────────────┐
│                                         │
│    [Fade in from black]                 │
│                                         │
│     Simulated Exchange                  │
│   Your Safe SRE Learning Lab            │
│                                         │
│   [Subtle system architecture in bg]   │
│                                         │
└─────────────────────────────────────────┘
```

**Animation:** Gentle fade in, slight parallax on background

---

### SCENE 2: The Challenge (0:10 - 0:45)
```
┌──────────────────┬──────────────────┐
│  PRODUCTION      │   ⚠️ WARNING    │
│  ─────────────   │                 │
│  🔴 C++ Engine   │  DO NOT TOUCH   │
│  🟡 Python Risk  │                 │
│  🟢 Java System  │  Critical       │
│                  │  Systems        │
│  💰 Real Money   │                 │
│  👥 Real Users   │  ⛔ No Testing │
└──────────────────┴──────────────────┘
```

**Animation:** Warning symbols pulsing, red border around production

---

### SCENE 3: The Learning Dilemma (0:45 - 1:15)
```
┌─────────────────────────────────────────┐
│   How do I learn...                     │
│                                         │
│   ❌ Can't use Production               │
│   ❌ Can't monopolize Test Env          │
│   ❌ No safe place for mistakes         │
│   ❌ Fear of breaking things            │
│                                         │
│   [Frustrated SRE at desk]              │
│   [Stack of unread documentation]       │
└─────────────────────────────────────────┘
```

**Animation:** Checkmarks appearing as X's, growing pile of docs

---

### SCENE 4: Introducing the Solution (1:15 - 2:00)
```
┌─────────────────────────────────────────┐
│  Terminal:                              │
│  $ make docker-up                       │
│  ✓ Starting postgres...                 │
│  ✓ Starting redis...                    │
│  ✓ Starting trading-api...              │
│  ✓ Starting market-simulator...         │
│                                         │
│  Architecture:                          │
│  ┌─────┐ ┌─────┐ ┌──────┐              │
│  │ API │─│Redis│─│Prom  │              │
│  └─────┘ └─────┘ └──────┘              │
│     │                                   │
│  ┌──────┐ ┌────────┐                   │
│  │ DB   │ │Grafana │                   │
│  └──────┘ └────────┘                   │
└─────────────────────────────────────────┘
```

**Animation:** Containers spinning up one by one, architecture diagram building

---

### SCENE 5: Language Agnostic (2:00 - 2:30)
```
┌───────────────┬───────────────┐
│   LEARNING    │  PRODUCTION   │
│   LAB (Go)    │  (C++/Py/Java)│
├───────────────┼───────────────┤
│  [Go code]    │ [C++ code]    │
│               │               │
│  ↓ Transform ↓│               │
│               │               │
│  eBPF Traces  │ eBPF Traces   │
│  Flamegraphs  │ Flamegraphs   │
│  Dashboards   │ Dashboards    │
│               │               │
│   [Same!]     │   [Same!]     │
└───────────────┴───────────────┘
```

**Animation:** Code fading out, revealing identical tools underneath

---

### SCENE 6: Demo - Getting Started (2:30 - 3:00)
```
┌─────────────────────────────────────────┐
│  Terminal:                              │
│  $ make docker-up                       │
│  [Real terminal output scrolling]       │
│                                         │
│  ↓ 30 seconds later...                  │
│                                         │
│  Browser: http://localhost              │
│  ┌───────────────────────────┐         │
│  │ Trading Dashboard         │         │
│  │ ───────────────────────── │         │
│  │ Orders/sec: 2,547         │         │
│  │ Latency P99: 23ms         │         │
│  │ Active Trades: 1,234      │         │
│  └───────────────────────────┘         │
└─────────────────────────────────────────┘
```

**Animation:** Terminal output scrolling, browser opening with fade

---

### SCENE 7: Demo - Chaos Engineering (3:00 - 3:45)
```
┌──────────────────┬──────────────────┐
│  Terminal        │  Grafana         │
│                  │                  │
│  $ ./chaos-      │  [Normal state]  │
│  experiments/    │  🟢 All healthy  │
│  01-service-     │                  │
│  failure.sh      │                  │
│                  │                  │
│  ✓ Killing API   │  [Impact!]       │
│                  │  🔴 Errors spike │
│                  │  📉 Requests ↓   │
│                  │                  │
│  ✓ Waiting...    │  [Recovery]      │
│                  │  🟡 Restarting   │
│                  │  🟢 Back online  │
└──────────────────┴──────────────────┘
```

**Animation:** Metrics in Grafana updating in real-time, dramatic color changes

---

### SCENE 8: Demo - Performance Profiling (3:45 - 4:30)
```
┌─────────────────────────────────────────┐
│  Terminal:                              │
│  $ ./scripts/generate-flamegraph.sh     │
│    cpu 30 8080                          │
│                                         │
│  ⏱️  Profiling... [=========>  ] 75%   │
│                                         │
│  ✓ Generated: cpu_20251116.svg         │
│  ✓ AI Analysis: cpu_analysis.md        │
│                                         │
│  Browser: [Interactive Flamegraph]     │
│  ┌───────────────────────────┐         │
│  │ ┌─────────────────────┐   │         │
│  │ │ runtime.futex 12.5% │   │ ← Hotspot│
│  │ ├─────────────┬───────┤   │         │
│  │ │syscall 8.3% │other  │   │         │
│  └─┴─────────────┴───────┴───┘         │
│                                         │
│  AI Analysis:                           │
│  🔴 Top Hotspot: runtime.futex         │
│     Recommendation: Reduce locks        │
└─────────────────────────────────────────┘
```

**Animation:** Progress bar filling, flamegraph zooming interactively

---

### SCENE 9: Demo - eBPF Tracing (4:30 - 5:00)
```
┌─────────────────────────────────────────┐
│  Terminal:                              │
│  $ sudo bpftrace -e                     │
│  'tracepoint:syscalls:sys_enter_* {     │
│    @[probe] = count();                  │
│  }'                                     │
│                                         │
│  [Real-time output:]                    │
│  @[sys_enter_read]: 1234                │
│  @[sys_enter_write]: 567                │
│  @[sys_enter_futex]: 890                │
│  @[sys_enter_epoll_wait]: 456           │
│                                         │
│  [Histogram appearing]                  │
│  sys_enter_read    ████████████ 1234    │
│  sys_enter_futex   ██████       890     │
└─────────────────────────────────────────┘
```

**Animation:** Live output streaming, histogram building dynamically

---

### SCENE 10: Web Documentation (5:00 - 5:30)
```
┌─────────────────────────────────────────┐
│  Browser: docs.andythomas-sre.com       │
│                                         │
│  📚 Documentation Hub                   │
│  ─────────────────────────────          │
│                                         │
│  🎓 Learning Lab - START HERE           │
│  │                                      │
│  ├─ ⭐ LEARNING_LAB_PURPOSE.md          │
│  ├─ 🔄 SKILLS_TRANSFERABILITY.md        │
│  └─ 🚀 SRE_ONBOARDING.md                │
│                                         │
│  🔥 SRE & Performance                   │
│  ├─ 🔥 Flamegraph Analysis Center       │
│  ├─ 📖 SRE_RUNBOOK.md                   │
│  └─ 📊 Performance guides               │
│                                         │
│  [Click through to flamegraph analysis] │
│  [Scroll showing AI recommendations]    │
└─────────────────────────────────────────┘
```

**Animation:** Smooth scrolling, clicking through links

---

### SCENE 11: The Value Proposition (5:30 - 6:15)
```
┌─────────────────────────────────────────┐
│   The Learning Ladder                   │
│                                         │
│   🔴 PRODUCTION                         │
│   ─────────────                         │
│   Real systems, zero risk              │
│         ↑                               │
│         │ Apply with confidence         │
│         │                               │
│   🟡 TEST ENVIRONMENT                   │
│   ─────────────────────                 │
│   Validate & integrate                  │
│         ↑                               │
│         │ Proven techniques             │
│         │                               │
│   🟢 LEARNING LAB                       │
│   ─────────────                         │
│   Learn & experiment safely             │
│                                         │
│   Statistics:                           │
│   • Time to competency: ⏱️ 50% faster   │
│   • Risk to production: ⛔ ZERO         │
│   • Team impact: ✅ None                │
└─────────────────────────────────────────┘
```

**Animation:** Arrows flowing upward, statistics counting up

---

### SCENE 12: Real-World Example (6:15 - 6:45)
```
┌─────────────────────────────────────────┐
│  Incident Timeline                      │
│                                         │
│  🚨 14:00 - Alert: P99 Latency Spike    │
│  ┌─────────────────────────┐           │
│  │ Grafana: 20ms → 200ms   │           │
│  └─────────────────────────┘           │
│           ↓                             │
│  🔍 14:05 - eBPF Investigation          │
│  $ sudo bpftrace ...                    │
│  Finding: Excessive futex calls         │
│           ↓                             │
│  📊 14:10 - Flamegraph Analysis         │
│  [Flamegraph showing hotspot]           │
│  Finding: 60% in vector::push_back()    │
│           ↓                             │
│  📝 14:20 - Evidence Package            │
│  ✓ Root cause: Line 423                 │
│  ✓ Impact: 10x latency                  │
│  ✓ Fix: Pre-allocate capacity           │
│           ↓                             │
│  ✅ RESOLVED - Without touching code!   │
└─────────────────────────────────────────┘
```

**Animation:** Timeline progressing step-by-step with checkmarks

---

### SCENE 13: Getting Started (6:45 - 7:00)
```
┌─────────────────────────────────────────┐
│   Get Started in 3 Steps                │
│                                         │
│   1️⃣  Clone & Start                     │
│       $ git clone ...                   │
│       $ make docker-up                  │
│       ⏱️  30 seconds                    │
│                                         │
│   2️⃣  Read the Docs                     │
│       docs.andythomas-sre.com           │
│       📖 Learning paths                 │
│       🔄 Transferability guide          │
│                                         │
│   3️⃣  Start Breaking Things!            │
│       ./chaos-experiments/*.sh          │
│       💥 Zero risk                      │
│       ✅ Real learning                  │
└─────────────────────────────────────────┘
```

**Animation:** Each step appearing with icon bounce effect

---

### SCENE 14: Call to Action (7:00 - 7:15)
```
┌─────────────────────────────────────────┐
│   [Montage of previous scenes]          │
│   - Terminal commands executing         │
│   - Grafana dashboards updating         │
│   - Flamegraphs zooming                 │
│   - Team members learning               │
│                                         │
│   [Fade to final card]                  │
│                                         │
│     Simulated Exchange                  │
│     Learning Lab                        │
│                                         │
│     Learn. Break. Master.               │
│                                         │
│     docs.andythomas-sre.com             │
│                                         │
│   [Fade to black]                       │
└─────────────────────────────────────────┘
```

**Animation:** Quick cuts montage (0.5 sec each), fade out

---

## Color Coding Reference

```
🔴 RED     - Production/Critical
🟡 YELLOW  - Test Environment/Caution
🟢 GREEN   - Learning Lab/Safe
🔵 BLUE    - Information/Neutral
⚫ BLACK   - Background/Transitions
```

---

## Typography Hierarchy

```
TITLE TEXT
   48pt Bold Sans-Serif
   Used for: Scene titles, main points

Subtitle Text
   24pt Regular Sans-Serif
   Used for: Section headers, descriptions

Code Text
   18pt Monospace
   Used for: Commands, file names

Body Text
   20pt Regular Sans-Serif
   Used for: Bullet points, explanations
```

---

## Transition Types

1. **Scene Changes:** Fade (0.5 sec)
2. **Within Scene:** Smooth slide/zoom (0.3 sec)
3. **Text Appearance:** Fade up or slide in
4. **Code:** Typing effect or instant + highlight
5. **Browser:** Smooth scroll, realistic speed

---

## Recording Tips by Scene

### Scenes with Terminal
- Use large font (18-20pt)
- Dark theme with high contrast
- Clear prompt (simple `$` is fine)
- Pause after each command output

### Scenes with Browser
- Full screen or clear window
- Remove bookmarks bar
- Clear history/cookies for clean look
- Zoom to 110-125% for readability

### Scenes with Grafana
- Use pre-configured dashboards
- Set time range to show clear data
- Use auto-refresh for live effect
- Maximize panels being shown

### Scenes with Code/Flamegraphs
- Highlight the area being discussed
- Zoom in on important sections
- Use cursor to point to specific lines
- Allow time for viewers to read

---

## Alternative: Animated Explainer Version

If you prefer animation over screen recording:

### Tool Options
- **Powtoon** - Easy, template-based
- **Vyond** - Professional characters
- **After Effects** - Full control
- **Animaker** - Good middle ground

### Style Guide for Animation
- Use flat design, modern colors
- Character: Friendly SRE person
- Icons: Consistent style (outline or filled)
- Motion: Smooth, not too fast
- Voiceover: Same script

---

## Accessibility Checklist

- [ ] Captions for all spoken words
- [ ] High contrast visuals
- [ ] Text readable at 720p
- [ ] Audio clear and leveled
- [ ] Descriptions for visual-only elements
- [ ] Keyboard navigation shown clearly
- [ ] Color not sole indicator of meaning

---

## File Organization

```
video-project/
├── raw-footage/
│   ├── scene-01-title.mov
│   ├── scene-02-challenge.mov
│   ├── scene-03-dilemma.mov
│   └── ...
├── audio/
│   ├── voiceover-segment-1.wav
│   ├── voiceover-segment-2.wav
│   └── background-music.mp3
├── graphics/
│   ├── title-card.png
│   ├── diagram-architecture.png
│   └── end-card.png
├── edit/
│   └── simulated-exchange-explainer.prproj
└── export/
    ├── simulated-exchange-1080p.mp4
    └── simulated-exchange-720p.mp4
```

---

This storyboard provides visual guidance for every second of the video. Use it alongside the script for professional results!
