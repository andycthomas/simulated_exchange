# 🎥 Quick Recording Guide - Start to Finish

## Fast Track: Record in One Session (2-3 hours)

### Before You Start (15 minutes)

```bash
# 1. Set up clean environment
cd ~/simulated_exchange
make docker-down
make clean-docker
make docker-up

# 2. Verify everything works
curl http://localhost:8080/health
open http://localhost:3000  # Grafana
open http://localhost       # Dashboard

# 3. Prepare terminal
# - Increase font size to 18pt
# - Use dark theme (Dracula, Solarized Dark, or similar)
# - Clear history: history -c
# - Simple prompt: export PS1='$ '
```

---

## Recording Setup Checklist

### Software
- [ ] **OBS Studio** installed (or ScreenFlow/Camtasia)
- [ ] **Audacity** for voiceover (or built into screen recorder)
- [ ] Browser: Chrome/Firefox with clean profile
- [ ] Terminal: iTerm2/Terminal with readable theme

### Hardware
- [ ] Microphone connected and tested
- [ ] Quiet room (turn off fans, AC, close windows)
- [ ] Notifications disabled (Do Not Disturb mode)
- [ ] Phone silenced
- [ ] Water nearby

### Screen
- [ ] Resolution set to 1920x1080
- [ ] Desktop clean (no sensitive info)
- [ ] Hide desktop icons
- [ ] Hide menu bar extras (Mac) or system tray (Linux)

---

## Quick OBS Studio Setup

### 1. Create Scene
```
Sources:
├─ Display Capture (fullscreen)
├─ Audio Input Capture (microphone)
└─ (Optional) Webcam if doing picture-in-picture
```

### 2. Settings
- **Video:** 1920x1080, 30 fps
- **Audio:** 44.1kHz, Stereo
- **Output:** MP4, High Quality

### 3. Test Recording
- Record 30 seconds
- Check audio levels (speak normally, watch meters)
- Verify video is smooth
- Check file plays back correctly

---

## Recording Order (Efficient Workflow)

### Phase 1: Screen Recordings (1 hour)
Record all terminal/browser scenes without voiceover:

1. **Terminal Setup Shots:**
   ```bash
   # Record these commands with pauses:
   make docker-up
   # Wait for complete startup

   curl http://localhost:8080/health
   # Show response

   ./chaos-experiments/01-service-failure.sh
   # Let it run completely

   ./scripts/generate-flamegraph.sh cpu 30 8080
   # Show full 30-second capture

   sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* { @[probe] = count(); }'
   # Let it collect data for 10 seconds, then Ctrl+C
   ```

2. **Browser Shots:**
   ```
   http://localhost          # Trading dashboard
   http://localhost:3000     # Grafana (during chaos experiment)
   flamegraphs/*.svg         # Open generated flamegraph
   docs.andythomas-sre.com   # Documentation site
   ```

3. **Tips for Terminal Recording:**
   - Type slowly and deliberately
   - Wait 2-3 seconds after each command completes
   - If you make a typo, pause and start the command again
   - Keep recordings as separate files (easier to edit)

### Phase 2: Voiceover (1 hour)
Record voiceover in quiet room:

1. **Warm up your voice:**
   - Drink water
   - Read script aloud once
   - Do some vocal exercises

2. **Recording segments:**
   - Record in scene-sized chunks (30-60 seconds each)
   - Do 2-3 takes of each segment
   - Leave 2 seconds silence before and after
   - Make mouth noises between segments (you'll cut these out)

3. **Voice recording tips:**
   - Sit up straight for better projection
   - Stay consistent distance from mic (6-12 inches)
   - Speak at moderate pace (140-160 words/min)
   - Put energy in your voice but don't oversell
   - If you mess up, pause, then start the sentence again

### Phase 3: Editing (30-45 minutes)

1. **Import all files into editor**
2. **Lay out video clips** on timeline
3. **Add voiceover** and sync to visuals
4. **Add transitions** between scenes (0.5 sec fade)
5. **Add text overlays** for key points
6. **Add background music** (quiet, -20dB)
7. **Color correct** if needed
8. **Export** in 1080p

---

## Scene-by-Scene Recording Commands

### Scene 6: Getting Started (Record this exactly)

```bash
# Start recording
$ make docker-up

# Let it run completely until you see:
# "simulated-exchange-caddy started"

# Then run:
$ curl http://localhost:8080/health

# Should show:
# {"status":"healthy",...}

# Open browser to localhost (record this)

# Stop recording
```

**Files produced:**
- `scene-06-startup.mov` (terminal)
- `scene-06-browser.mov` (dashboard)

---

### Scene 7: Chaos Engineering (Record this exactly)

```bash
# Start recording terminal AND browser side-by-side

# Terminal:
$ ./chaos-experiments/01-service-failure.sh

# In Grafana, navigate to:
# http://localhost:3000/d/trading-overview

# You'll see:
# - Errors spike
# - Requests drop
# - Service goes red
# - Then recovers

# Let it run completely (about 60 seconds)

# Stop recording
```

**Files produced:**
- `scene-07-chaos-split.mov` (terminal + Grafana side-by-side)

**Alternative:** Record separately and composite in editing

---

### Scene 8: Flamegraph (Record this exactly)

```bash
# Start recording
$ ./scripts/generate-flamegraph.sh cpu 30 8080

# This will take 30 seconds - let it run

# When complete, you'll see:
# ✓ Flamegraph generated: ...
# ✓ Analysis report generated: ...

# Then open the SVG:
$ open flamegraphs/cpu_*.svg

# In browser:
# - Zoom into a few functions
# - Hover to show details
# - Show the wide bars (hotspots)

# Then open analysis:
$ cat flamegraphs/cpu_*_analysis.md | head -50

# Stop recording
```

**Files produced:**
- `scene-08-profile-generation.mov` (terminal)
- `scene-08-flamegraph-interactive.mov` (browser)
- `scene-08-analysis.mov` (terminal showing analysis)

---

### Scene 9: eBPF (Record this exactly)

```bash
# Start recording
$ sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* { @[probe] = count(); }'

# Let it run for about 10 seconds while system is active

# You'll see real-time updates

# Press Ctrl+C to stop

# Output shows histogram of syscalls

# Stop recording
```

**Files produced:**
- `scene-09-ebpf.mov`

---

### Scene 10: Documentation (Record this exactly)

```bash
# Open browser to:
https://docs.andythomas-sre.com

# Start recording browser window

# Show main index:
# - Scroll down slowly
# - Show the categories

# Click on:
"LEARNING_LAB_PURPOSE.md"

# Scroll through showing the diagrams

# Go back, click on:
"Flamegraph Analysis Center"

# Show the flamegraph cards

# Click to view one analysis

# Show the markdown rendering

# Stop recording
```

**Files produced:**
- `scene-10-docs-index.mov`
- `scene-10-docs-purpose.mov`
- `scene-10-docs-flamegraphs.mov`

---

## Voiceover Recording Template

### Setup Audacity
1. **File → Preferences → Recording:**
   - Channels: Mono (or Stereo if you prefer)
   - Sample Rate: 44100 Hz

2. **Effects to use after recording:**
   - Noise Reduction (capture room tone first)
   - Normalize to -3dB
   - Compressor (Ratio: 3:1, Threshold: -20dB)
   - EQ if needed (boost presence around 3-5kHz slightly)

### Recording Script Format

For each scene, record like this:

```
[SCENE 1 - Take 1]
[2 seconds silence]
What if you could learn chaos engineering, performance profiling,
and eBPF tools without any risk to production systems?
[2 seconds silence]

[SCENE 1 - Take 2]
[2 seconds silence]
What if you could learn chaos engineering, performance profiling,
and eBPF tools without any risk to production systems?
[2 seconds silence]
```

**Then pick the best take in editing.**

---

## Time-Saving Tips

### Batch Record Similar Scenes
Group terminal scenes together, browser scenes together:

**Terminal batch:**
- Setup
- Chaos experiment
- Flamegraph generation
- eBPF tracing

**Browser batch:**
- Dashboard
- Grafana during chaos
- Flamegraph viewing
- Documentation browsing

### Use Multiple Takes Strategy
- Record each terminal command 2x
- If you mess up, just start that command again
- Mark good takes with a clap or comment
- Pick best takes in editing

### Pre-Stage Everything
```bash
# Before recording, prepare:
make docker-up
# Let it fully start

# Open all browsers you'll need:
open http://localhost
open http://localhost:3000
open https://docs.andythomas-sre.com

# Then you can just switch between them while recording
```

---

## Common Issues & Solutions

### Issue: Audio has background noise
**Solution:**
- Record 5 seconds of room tone (silence)
- Use Audacity's Noise Reduction effect
- Settings: Noise Reduction: 12dB, Sensitivity: 6.00

### Issue: Text too small to read
**Solution:**
- Terminal: Increase font to 20pt
- Browser: Zoom to 125% (Cmd/Ctrl + +)
- Re-record the problematic scenes

### Issue: Commands run too fast
**Solution:**
- In editing, slow down video to 0.5x speed if needed
- Or re-record with deliberate pauses

### Issue: Made a mistake while recording
**Solution:**
- Don't stop! Pause, take a breath
- Start the command/sentence again
- You'll edit out the mistake

### Issue: Voiceover doesn't match video timing
**Solution:**
- Cut video to match voiceover (easier than vice versa)
- Add B-roll or screenshots to fill gaps
- Speed up/slow down video slightly (90-110%)

---

## Export Settings

### For Web (YouTube, Vimeo)
**Video:**
- Format: MP4 (H.264)
- Resolution: 1920x1080
- Frame rate: 30 fps
- Bitrate: 8-12 Mbps

**Audio:**
- Codec: AAC
- Bitrate: 192 kbps
- Sample rate: 44.1 kHz

### For Internal Hosting
**Same as above, or:**
- Consider 720p for smaller file size
- Lower bitrate (5-8 Mbps) acceptable

### For Archive/Master
- Keep original resolution
- Higher bitrate (15-20 Mbps)
- Separate audio track
- Save project file for future edits

---

## Quick Editing Workflow (DaVinci Resolve - Free)

### 1. Import Media
- Drag all video files into Media Pool
- Organize in bins: "Terminal", "Browser", "Graphics"

### 2. Rough Cut
- Drag clips to timeline in order
- Trim to remove mistakes/pauses
- Leave gaps for voiceover

### 3. Add Voiceover
- Import audio files
- Align with video
- Adjust timing of video to match audio

### 4. Transitions
- Add dissolve between major scenes (0.5 sec)
- Keep cuts within scenes

### 5. Text Overlays
- Use Fusion titles
- Keep simple and readable
- Fade in/out (0.3 sec)

### 6. Color Correction
- Apply "Film Look" LUT if desired
- Or just normalize brightness/contrast
- Ensure terminal and browser are visible

### 7. Audio Polish
- Normalize voiceover to -3dB
- Add subtle background music at -20dB
- Fade music in/out

### 8. Export
- Format: H.264
- Preset: YouTube 1080p
- Filename: `simulated-exchange-explainer-v1.mp4`

---

## Checklist Before Final Export

- [ ] Watch entire video start to finish
- [ ] Check audio levels (nothing peaking)
- [ ] Verify all text is readable
- [ ] Ensure smooth transitions
- [ ] No jarring cuts
- [ ] Background music not too loud
- [ ] Voiceover clear and synchronized
- [ ] End card visible for 3+ seconds
- [ ] Total length under 8 minutes
- [ ] Exported in correct format

---

## After Publishing

### YouTube Setup
- **Title:** "SRE Learning Lab - Safe Chaos Engineering & Performance Profiling"
- **Description:**
  ```
  Learn SRE skills safely with the Simulated Exchange Learning Lab.

  Practice chaos engineering, performance profiling, and eBPF tracing
  without risking production systems. 90% of skills transfer directly
  to C++, Python, and Java applications.

  Resources:
  📖 Documentation: https://docs.andythomas-sre.com
  🚀 Get Started: [link to repo]
  🔥 Flamegraph Guide: [link to guide]

  Timestamps:
  0:00 - Introduction
  0:10 - The Challenge
  1:15 - The Solution
  2:30 - Live Demo
  5:30 - The Value
  6:45 - Getting Started
  ```

- **Tags:** SRE, DevOps, Chaos Engineering, Performance, eBPF, Observability, Flamegraphs, Trading Systems, Learning Lab

- **Thumbnail:** Create eye-catching image with:
  - "ZERO RISK LEARNING"
  - Flamegraph visual
  - "SRE Skills Lab"

### Track Metrics
- Views in first week
- Average watch time (aim for >70%)
- Click-through to documentation
- Team member completion of onboarding

---

## Estimated Time Breakdown

- **Pre-production:** 30 minutes
- **Screen recording:** 60 minutes
- **Voiceover:** 45 minutes
- **Editing:** 45 minutes
- **Review & export:** 15 minutes
- **TOTAL:** ~3.5 hours

With practice, you can reduce this to 2 hours!

---

## Need Help?

- **OBS Studio tutorials:** https://obsproject.com/wiki/
- **Audacity guides:** https://manual.audacityteam.org/
- **DaVinci Resolve:** https://www.blackmagicdesign.com/products/davinciresolve/training

---

**You're ready to create a professional explainer video!** Follow this guide step-by-step and you'll have a polished video in one afternoon.
