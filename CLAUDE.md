# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A realistic sailing simulation game built with Go and Ebitengine, supporting both native desktop and WebAssembly (browser) deployments. Features include realistic polar-based boat physics, race timing, OCS detection, mark rounding, and an online leaderboard (WASM only).

## Build Commands

- `make build` - Build native desktop version
- `make run` - Build and run desktop version
- `make web` - Build WASM and start development server on http://localhost:8080
- `make wasm-static` - Build WASM preserving existing web/index.html (use when Firebase config is present)
- `make wasm` - Build WASM with generated index.html (may overwrite Firebase config)
- `make clean` - Clean build artifacts

## Development Workflow

**Desktop development:**
```bash
make run  # Fast iteration for game logic changes
```

**Web development:**
```bash
make web  # Builds WASM and serves on localhost:8080
```

**Deploying with Firebase:**
```bash
make wasm-static  # Preserves Firebase configuration in web/index.html
```

## Architecture

### Core Game Loop (pkg/game/game.go)

The `GameState` struct is the central coordinator implementing Ebitengine's `Game` interface:
- `Update()` - Processes input, updates physics, detects race events (60 FPS)
- `Draw()` - Renders game world with camera offset, UI overlays, and banners
- Manages race state: pre-start countdown, OCS detection, line crossing, mark rounding, finish detection
- Coordinates between boat physics, wind system, dashboard UI, and mobile controls

### Platform Abstraction

The codebase uses Go build tags for platform-specific functionality:
- `platform_wasm.go` (+build js,wasm) - WASM-specific code
- `platform_native.go` (+build !js !wasm) - Native desktop code
- `firebase_wasm.go` - Firebase integration (WASM only, connects to JavaScript Firebase SDK)
- `firebase_native.go` - No-op Firebase stubs for native builds

### Physics System

**Polars (pkg/polars/polars.go):**
- `RealisticPolar` implements actual sailboat performance data
- Interpolates boat speed from TWA (True Wind Angle) and TWS (True Wind Speed)
- Handles close-hauled sailing (30-52°) using beat VMG calculations
- Speed table covers 4-24 knots wind, 52-180° angles

**Wind System (pkg/game/world/wind.go):**
- `Wind` interface allows different wind implementations
- `OscillatingWind` provides realistic wind shifts (-10° to +10°) with three-phase cycles
- `VariableWind` interpolates speed across the course (favors left or right side randomly)
- Wind selected randomly at game start (50/50 left vs right side advantage)

**Boat Updates (pkg/game/objects/boat.go):**
- Physics update calculates TWA from heading and wind direction
- Queries polar diagram for target speed
- Accelerates/decelerates toward target speed
- Updates velocity components and position

### Race Mechanics

**Race State Flow:**
1. Pre-start (30s countdown) - boat positioned below starting line
2. OCS detection - bow crossing line early triggers warning
3. Race start at timer zero
4. Line crossing detection - tracks start time, VMG, speed percentage
5. Mark rounding phases - must sail north past mark, travel west, then sail south
6. Finish line crossing - returns from north to south through starting line
7. Leaderboard submission (WASM only)

**Coordinate System:**
- Meter-based world coordinates (1 pixel = 1 meter at base zoom)
- World size: 2000m x 3000m
- Screen viewport: 1280x720 pixels
- Camera pans to follow boat with 200px margins
- Wind direction: 0° = North (toward -Y), 90° = East (+X)

### UI and Controls

**Dashboard (pkg/dashboard/dashboard.go):**
- Displays heading, speed, VMG, wind info
- Shows laylines to upwind mark
- Renders starting/finish line
- Race progress indicators

**Mobile Controls (pkg/game/mobile_controls.go):**
- Detects touch input to show on-screen buttons
- Turn left/right, pause, restart buttons
- Auto-hides on keyboard-only devices

**Telltales (pkg/game/telltales.go):**
- Visual sailing feedback when close-hauled
- Shows optimal angle guidance

### Firebase Integration (WASM Only)

- Scoreboard submission after race finish
- Leaderboard query (top 50 times, mark_rounded=true)
- JavaScript interop via syscall/js
- Firestore collection: "race_results"

## Key Files

- `cmd/gosailing/main.go` - Entry point (both native and WASM)
- `cmd/wasm_server/main.go` - Development web server with proper WASM headers
- `pkg/game/game.go` - Main game state and race logic (800 lines)
- `pkg/game/objects/boat.go` - Boat physics and rendering
- `pkg/game/world/wind.go` - Wind implementations
- `pkg/game/world/arena.go` - Mark definitions and rendering
- `pkg/polars/polars.go` - Boat performance data
- `pkg/dashboard/dashboard.go` - UI and instrumentation
- `web/index.html` - Static HTML with Firebase config (preserved by wasm-static)

## Testing Notes

- Use `J` key to jump timer forward 10 seconds (pre-start only)
- Use `C` key to toggle mobile controls display for testing
- Use `R` key to restart game
- WASM build requires proper CORS headers (provided by wasm_server)

## Common Gotchas

- Always use `make wasm-static` when Firebase is configured to avoid overwriting web/index.html
- Camera uses world coordinates; boat position must be offset by CameraX/CameraY when drawing
- Wind shifts are time-based; use Update() to advance oscillation phases
- Mobile controls only visible when touch input detected (hasTouchInput flag)
- Race timer is elapsed-time-based to support pause functionality
