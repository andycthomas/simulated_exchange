# 🎬 Explainer Video Script - Simulated Exchange Learning Lab

## Video Overview

**Title:** "Safe Sandbox for SRE Learning: The Simulated Exchange Lab"
**Duration:** 5-7 minutes
**Target Audience:** SRE team members, engineering managers
**Format:** Screen recording + voiceover with occasional text overlays
**Tone:** Professional but approachable, educational

---

## 🎯 Video Structure

### Act 1: The Problem (0:00 - 1:30)
### Act 2: The Solution (1:30 - 2:30)
### Act 3: The Demo (2:30 - 5:30)
### Act 4: The Value (5:30 - 7:00)

---

## 📝 SCRIPT

### SCENE 1: Title Card (0:00 - 0:10)

**Visual:**
- Black screen fading to title
- Logo or system diagram in background
- Text: "Simulated Exchange: Your Safe SRE Learning Lab"

**Voiceover:**
> "What if you could learn chaos engineering, performance profiling, and eBPF tools without any risk to production systems?"

**On-screen text:**
- "Zero Risk Learning"
- "Real SRE Skills"

---

### SCENE 2: The Challenge (0:10 - 0:45)

**Visual:**
- Split screen showing:
  - LEFT: Production system diagram (C++, Python, Java services)
  - RIGHT: Warning symbols, "DO NOT TOUCH" signs

**Voiceover:**
> "As an SRE, you need to master advanced observability tools, chaos engineering, and performance analysis. But here's the problem..."

**Pause 1 second**

> "Your production systems run critical C++, Python, and Java applications. They handle real money, real customers, real risk. You can't experiment there."

**On-screen text overlay:**
- "Production: Real customers, real risk"
- "Test Environment: Shared by entire team"

**Visual transition:**
- Show calendar with "Test Env: BOOKED" across multiple days

**Voiceover:**
> "Even your 17-server test environment is shared across teams and used for critical validation. Breaking it affects everyone."

---

### SCENE 3: The Learning Dilemma (0:45 - 1:15)

**Visual:**
- Show frustrated developer at computer
- Screen shows: "How do I learn eBPF without breaking things?"
- Bullet points appearing:
  - ❌ Can't experiment in production
  - ❌ Can't monopolize test environment
  - ❌ No safe place to make mistakes
  - ❌ Fear of breaking things while learning

**Voiceover:**
> "So how do you learn? You can't practice chaos engineering in production. You can't crash test servers while others are using them. And reading documentation only gets you so far."

**Visual:**
- Show books/documentation piling up
- Person looking overwhelmed

**Voiceover:**
> "You need hands-on experience. You need to break things. But safely."

---

### SCENE 4: Introducing the Solution (1:15 - 2:00)

**Visual:**
- Screen wipe to clean, modern interface
- Show terminal with: `git clone simulated_exchange && make docker-up`
- Containers spinning up in animation

**Voiceover:**
> "Introducing the Simulated Exchange Learning Lab. A complete, realistic trading system that you can break, crash, and experiment with - completely risk-free."

**Visual:**
- Show system architecture diagram appearing piece by piece:
  - Trading API (Go)
  - Market Simulator
  - PostgreSQL
  - Redis
  - Prometheus
  - Grafana

**Voiceover:**
> "It's a fully functional trading exchange with microservices, databases, message queues, and real metrics - processing over 2,500 orders per second."

**On-screen text:**
- "Realistic: 2,500+ orders/sec"
- "Complete: 6 services, full observability"
- "Safe: Runs on your laptop"

---

### SCENE 5: Key Point - Language Agnostic (2:00 - 2:30)

**Visual:**
- Split screen showing:
  - LEFT: Go code (simulated exchange)
  - RIGHT: C++, Python, Java code (production)

**Voiceover:**
> "Now, you might be thinking: 'My production systems are C++, Python, and Java - why learn from a Go application?'"

**Visual:**
- Erase the code, replace with:
  - LEFT: eBPF terminal, Flamegraph, Grafana
  - RIGHT: Same eBPF terminal, Flamegraph, Grafana

**Voiceover:**
> "Here's the key: You're not learning Go. You're learning eBPF, chaos engineering, and performance profiling - skills that are 90% language-agnostic."

**Visual:**
- Show three columns appearing:
  - "eBPF: Traces Linux kernel - works on ANY language"
  - "Chaos Engineering: Infrastructure level - language independent"
  - "Flamegraphs: Visual format - same for C++, Python, Java, Go"

**Voiceover:**
> "eBPF traces the Linux kernel, not your application language. Chaos engineering targets containers and infrastructure. Flamegraphs look identical whether they're from C++, Python, Java, or Go."

---

### SCENE 6: Demo - Getting Started (2:30 - 3:00)

**Visual:**
- Terminal recording showing actual commands

**Voiceover:**
> "Let's see it in action. Starting the lab takes just one command."

**Screen shows:**
```bash
make docker-up
```

**Visual:**
- Containers starting up (real terminal output)
- Health checks turning green

**Voiceover:**
> "In 30 seconds, you have a complete trading exchange running locally. Let's check the dashboard."

**Visual:**
- Browser opening to http://localhost
- Trading dashboard with live metrics

---

### SCENE 7: Demo - Chaos Engineering (3:00 - 3:45)

**Visual:**
- Split screen:
  - LEFT: Terminal
  - RIGHT: Grafana dashboard with metrics

**Voiceover:**
> "Now let's break something. We'll run a chaos experiment - killing the trading API service."

**Screen shows:**
```bash
./chaos-experiments/01-service-failure.sh
```

**Visual:**
- Grafana dashboard showing:
  - Error rate spiking
  - Request count dropping
  - Service going red

**Voiceover:**
> "Watch the metrics. The service crashes, error rates spike, and then... it recovers automatically. In production, this would be a P1 incident. Here? It's Tuesday's learning exercise."

**Visual:**
- Show service restarting
- Metrics returning to normal
- Green checkmark appearing

**Voiceover:**
> "You just practiced incident response in a completely safe environment."

---

### SCENE 8: Demo - Performance Profiling (3:45 - 4:30)

**Visual:**
- Terminal showing flamegraph generation

**Voiceover:**
> "Next, let's do performance profiling. We'll generate a CPU flamegraph with AI-powered analysis."

**Screen shows:**
```bash
./scripts/generate-flamegraph.sh cpu 30 8080
```

**Visual:**
- Progress bar showing 30-second capture
- File being created

**Voiceover:**
> "The script captures 30 seconds of CPU activity and generates an interactive flamegraph plus AI analysis."

**Visual:**
- Interactive flamegraph opening in browser
- Zooming into a function
- Showing wide bars (hotspots)

**Voiceover:**
> "The flamegraph shows exactly where the CPU time is spent. Wide bars are hotspots - the functions consuming the most resources."

**Visual:**
- Switch to AI analysis report showing:
  - "Top Hotspot: runtime.futex - 12.5% CPU"
  - "Recommendations: Reduce lock contention"

**Voiceover:**
> "And the AI analysis identifies the top performance bottlenecks with specific recommendations. These same techniques work on your C++, Python, and Java production systems."

---

### SCENE 9: Demo - eBPF Tracing (4:30 - 5:00)

**Visual:**
- Terminal showing eBPF command

**Voiceover:**
> "Want to learn eBPF tools? Try them here first."

**Screen shows:**
```bash
sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* { @[probe] = count(); }'
```

**Visual:**
- Real-time output showing system calls
- Histogram appearing

**Voiceover:**
> "This eBPF script traces all system calls in real-time. It's a powerful technique, but you wouldn't want to experiment with it in production first."

**Visual:**
- Side-by-side comparison:
  - LEFT: "Learning Lab - Safe to experiment"
  - RIGHT: "Production - Apply with confidence"

**Voiceover:**
> "Learn the syntax here, understand the overhead, then apply the exact same commands to your production C++ and Java applications."

---

### SCENE 10: Web Documentation Hub (5:00 - 5:30)

**Visual:**
- Browser showing https://docs.andythomas-sre.com

**Voiceover:**
> "Everything is documented in a beautiful web interface. Start with the Learning Lab Purpose to understand the 'why'."

**Visual:**
- Click through to "SKILLS_TRANSFERABILITY.md"
- Scroll through showing the 90/10 rule diagram

**Voiceover:**
> "The Skills Transferability guide shows exactly how each skill applies to your C++, Python, and Java systems. Spoiler: 90% of what you learn transfers directly."

**Visual:**
- Show flamegraph analysis page
- Click to view an AI-generated analysis

**Voiceover:**
> "AI-powered flamegraph analysis helps you understand performance bottlenecks. And there's a complete 8-week learning curriculum to take you from beginner to expert."

---

### SCENE 11: The Value Proposition (5:30 - 6:15)

**Visual:**
- Three-column diagram:
  - Production (Red border)
  - Test Environment (Yellow border)
  - Learning Lab (Green border)

**Voiceover:**
> "This lab doesn't replace your test environment. It complements it. Think of it as a learning ladder."

**Visual:**
- Animation showing progression:
  1. LEARN in Learning Lab
  2. VALIDATE in Learning Lab
  3. APPLY to Test Environment
  4. DEPLOY to Production

**Voiceover:**
> "First, learn and experiment in the lab. Zero risk, unlimited retries. Then validate your approach works. Only after that do you apply it to the test environment - with confidence and a proven technique."

**Visual:**
- Show statistics appearing:
  - "Time to learn eBPF: Days → Hours"
  - "Risk to production: Eliminated"
  - "Team productivity: Unaffected"
  - "Confidence level: High"

**Voiceover:**
> "The result? Faster skill development, zero risk to critical systems, and SREs who can confidently use advanced tools."

---

### SCENE 12: Real-World Example (6:15 - 6:45)

**Visual:**
- Show scenario unfolding as animation/screenshots

**Voiceover:**
> "Here's a real scenario. Your C++ trading engine starts showing high latency. You use skills learned in this lab."

**Visual:**
- Show Grafana alert: "P99 Latency: 20ms → 200ms"

**Voiceover:**
> "You detect the problem in Grafana - a skill you practiced here."

**Visual:**
- Show eBPF trace output

**Voiceover:**
> "You use eBPF to trace system calls - a tool you mastered here safely."

**Visual:**
- Show flamegraph with hotspot highlighted

**Voiceover:**
> "You generate a flamegraph showing 60% time in vector push_back - analysis techniques learned here."

**Visual:**
- Show evidence package document

**Voiceover:**
> "You package the evidence and hand it to the C++ developers. Root cause identified in hours, not days. All without touching a line of production code."

---

### SCENE 13: Getting Started (6:45 - 7:00)

**Visual:**
- Show three simple steps appearing:

**Voiceover:**
> "Ready to start? Three steps:"

**Visual:**
- Step 1: Terminal with git clone
- Step 2: Browser showing docs
- Step 3: Person following tutorial

**Voiceover:**
> "One: Clone the repo and run make docker-up.
> Two: Read the documentation at docs.andythomas-sre.com.
> Three: Complete the 30-minute onboarding and start breaking things safely."

**On-screen text:**
- "1. `make docker-up`"
- "2. docs.andythomas-sre.com"
- "3. Break things safely!"

---

### SCENE 14: Call to Action (7:00 - 7:15)

**Visual:**
- Split screen montage:
  - Terminal showing chaos experiments
  - Flamegraphs zooming
  - Dashboards updating
  - Happy team members learning

**Voiceover:**
> "Your safe SRE learning environment is ready. Master chaos engineering, performance profiling, and eBPF - all without risking production systems."

**Visual:**
- Fade to final title card:
  - "Simulated Exchange Learning Lab"
  - "Learn. Break. Master."
  - "docs.andythomas-sre.com"

**Voiceover:**
> "The Simulated Exchange Learning Lab. Learn safely. Apply confidently."

**Fade to black**

---

## 🎨 VISUAL STYLE GUIDE

### Color Palette
- **Safe/Learning Lab:** Green (#28a745)
- **Test Environment:** Yellow/Orange (#ffc107)
- **Production:** Red (#dc3545)
- **Neutral UI:** Dark blue/purple gradient (#667eea to #764ba2)

### Typography
- **Main titles:** Bold, 48pt, sans-serif
- **Subtitles:** Regular, 24pt, sans-serif
- **Code:** Monospace, 18pt
- **Voiceover captions:** Regular, 20pt (if needed for accessibility)

### Transitions
- **Between scenes:** 0.5 second fade
- **Within scenes:** Smooth animations, 0.3 second
- **Code appearing:** Typing effect or instant with highlight

### Screen Recording Settings
- **Resolution:** 1920x1080 (1080p)
- **Frame rate:** 30 fps minimum
- **Terminal:** Dark theme with high contrast
- **Browser:** Clean profile, no extra toolbars
- **Font size:** Large enough to read clearly (16-18pt in terminal)

---

## 🎙️ VOICEOVER NOTES

### Pacing
- Speak at 140-160 words per minute (moderate pace)
- Pause for 1-2 seconds at scene transitions
- Leave 2-3 seconds of silence for visual demonstration
- Total word count: ~1,100 words = ~7 minutes at 150 wpm

### Tone
- Professional but friendly
- Confident and reassuring
- Enthusiastic about the technology without overselling
- Clear and articulate

### Recording Tips
- Use a quality microphone (Blue Yeti, or similar)
- Record in a quiet room
- Do multiple takes of each section
- Leave pauses between sections for easy editing
- Record in segments matching the scenes above

---

## 📹 RECORDING CHECKLIST

### Pre-Production
- [ ] Set up clean demo environment
- [ ] Prepare terminal with proper theme/font size
- [ ] Clean browser (no personal bookmarks visible)
- [ ] Test all commands work smoothly
- [ ] Prepare any graphics/animations needed
- [ ] Write out full voiceover script with timing

### Recording Setup
- [ ] Screen recording software (OBS Studio, Camtasia, or ScreenFlow)
- [ ] Microphone tested and positioned
- [ ] Quiet environment
- [ ] Close unnecessary applications
- [ ] Turn off notifications
- [ ] Have water nearby for voiceover

### During Recording
- [ ] Record terminal sessions at 1920x1080
- [ ] Keep mouse movements smooth and purposeful
- [ ] Wait for commands to complete fully
- [ ] Record voiceover in consistent environment
- [ ] Get multiple takes for each section

### Post-Production
- [ ] Edit together screen recordings
- [ ] Sync voiceover to visuals
- [ ] Add transitions between scenes
- [ ] Add on-screen text overlays
- [ ] Add background music (subtle, non-intrusive)
- [ ] Color grade for consistency
- [ ] Add captions/subtitles
- [ ] Export in multiple formats (1080p, 720p)

---

## 🎬 ALTERNATIVE: QUICK 2-MINUTE VERSION

If you need a shorter elevator pitch version:

### Quick Script (2 minutes)

**0:00-0:20 - Problem:**
"SREs need to learn chaos engineering and performance tools, but can't experiment in production."

**0:20-0:40 - Solution:**
"The Simulated Exchange is a safe learning lab - a complete trading system you can break freely."

**0:40-1:20 - Demo:**
- Run chaos experiment (kill service)
- Generate flamegraph
- Show eBPF trace

**1:20-1:50 - Value:**
"90% of skills transfer to C++/Python/Java production systems. Learn safely, apply confidently."

**1:50-2:00 - CTA:**
"Get started at docs.andythomas-sre.com"

---

## 📊 METRICS TO TRACK

After publishing the video:

- [ ] View count
- [ ] Average watch time (aim for >70%)
- [ ] Click-through rate to documentation
- [ ] Number of team members who complete onboarding
- [ ] Feedback/questions received

---

## 🔧 TOOLS YOU'LL NEED

### Screen Recording
- **OBS Studio** (Free, open source)
- **Camtasia** (Paid, easier editing)
- **ScreenFlow** (Mac only, professional)
- **Loom** (Quick and easy, web-based)

### Video Editing
- **DaVinci Resolve** (Free, professional)
- **Adobe Premiere Pro** (Paid, industry standard)
- **Final Cut Pro** (Mac only, professional)
- **iMovie** (Mac only, simple)

### Audio Recording
- **Audacity** (Free, good quality)
- **Adobe Audition** (Paid, professional)
- **GarageBand** (Mac only, easy)

### Graphics/Animation
- **Canva** (Easy diagrams and text overlays)
- **Figma** (Professional design tool)
- **PowerPoint/Keynote** (For simple animations)

---

## 💡 PRO TIPS

1. **Record terminal sessions separately** from voiceover - easier to edit
2. **Use 1.5x speed** for command execution in post-production if needed
3. **Add subtle background music** at -20dB to keep energy up
4. **Include captions** - many people watch with sound off
5. **Keep cursor visible** but not distracting
6. **Highlight commands** before execution (zoom or color box)
7. **Show results** fully before moving to next scene
8. **Test on different screens** - what's readable on 27" might not be on 13"

---

## 📤 PUBLISHING

### Hosting Options
- **YouTube** - Best for public/team access
- **Vimeo** - Professional, more control
- **Internal portal** - If company has video hosting
- **GitHub** - Link from README to hosted video

### SEO/Discovery
- **Title:** "SRE Learning Lab - Safe Chaos Engineering & Performance Profiling"
- **Tags:** SRE, DevOps, Chaos Engineering, eBPF, Performance, Observability
- **Description:** Full script summary with links
- **Thumbnail:** Eye-catching with "Zero Risk Learning" text

---

**Ready to record!** This script gives you everything needed to create a professional explainer video.
