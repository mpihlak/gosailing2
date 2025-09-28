# Go Sailing Game 🚢

A realistic sailing simulation built with LLMs, Go and Ebitengine, playable both as a desktop application and in web browsers via WebAssembly.

## Features

- **Realistic sailing physics** with polar diagrams
- **Complete race system** with countdown timer and race finishing
- **OCS detection** (On Course Side) rule enforcement
- **Mark rounding detection** with proper sailing rules
- **Race timing** with finish time display
- **VMG calculations** (Velocity Made Good)
- **Interactive starting line** with pin flag and committee boat
- **Infinite sailing world** without boundaries
- **Cross-platform** - Desktop and Web/WASM support

## Quick Start

### Desktop Version
```bash
make run
```

### Web Version (WASM)
```bash
make web
```
This will:
1. Build the WASM version
2. Start a local web server on http://localhost:8080
3. Automatically open your browser
4. Serve the game with proper WASM headers

## Build Commands

| Command | Description |
|---------|-------------|
| `make help` | Show available commands |
| `make build` | Build desktop version |
| `make run` | Build and run desktop version |
| `make web` | Build and serve web version |
| `make wasm` | Build WASM version only (no server) |
| `make clean` | Clean build artifacts |

## Game Controls

| Key | Action |
|-----|--------|
| ← → | Steer left/right |
| Space | Pause/Resume game |
| J | Jump timer forward 10 seconds |
| Q | Quit game |

## Racing Rules

1. **Pre-start**: Position your boat behind the starting line
2. **Countdown**: Watch the timer count down from 1 minute
3. **OCS Warning**: Red banner appears if you cross the line early
4. **Race Start**: Green starting line when timer reaches zero
5. **Starting**: Cross the starting line from south to north (bow must cross)
6. **Mark Rounding**: Sail to upwind mark and round it properly:
   - Sail past the mark (south to north)
   - Pass to the left side (east to west while north of mark)
   - Sail below the mark (north to south)
7. **Finishing**: Cross the finish line from north to south after rounding
8. **Race Complete**: Timer stops and "RACE FINISHED" banner displays your time

## Game Mechanics

- **Wind**: Constant 15 knots from North (0°)
- **Boat Speed**: Determined by realistic polar curves
- **VMG**: Velocity Made Good towards/away from wind
- **Starting Line**: 400 meter line with pin flag and committee boat
- **Mark Laylines**: Visual aids showing optimal sailing angles to the upwind mark
- **Infinite World**: Sail in any direction without boundaries

## Technical Details

- **Engine**: Ebitengine v2
- **Language**: Go 1.21+
- **Web Support**: WebAssembly (WASM)
- **Coordinates**: Meter-based world coordinates
- **Graphics**: 2D pixel-perfect rendering

## Development

### Project Structure
```
cmd/
  gosailing/     - Desktop version main
  wasm_server/   - Web server for WASM version
pkg/
  game/          - Core game logic
  coordinates/   - World coordinate system
  dashboard/     - UI and instrumentation
  geometry/      - Mathematical utilities
  polars/        - Boat performance curves
web/             - Generated WASM build output
```

### Building for Different Platforms

**Desktop (native):**
```bash
go build ./cmd/gosailing
```

**Web (WASM):**
```bash
GOOS=js GOARCH=wasm go build -o sailing.wasm ./cmd/gosailing
```

**Web Server:**
```bash
go run ./cmd/wasm_server
```

## Browser Compatibility

The web version works in all modern browsers that support:
- WebAssembly
- Canvas API
- Keyboard input

Tested on:
- Chrome 90+
- Firefox 89+
- Safari 14+
- Edge 90+

## Performance Notes

- **Desktop**: 60 FPS, native performance
- **Web/WASM**: ~45-60 FPS, depending on browser
- **Memory**: Low memory usage with efficient rendering
- **Network**: Initial WASM download ~2-5MB

Enjoy sailing! ⛵
