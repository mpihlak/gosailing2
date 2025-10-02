# Build Instructions

This document explains the different build targets and how to properly deploy the game with Firebase leaderboard support.

## Build Targets

### `make wasm-static` (Recommended for Firebase)
- Builds WASM version while preserving existing `web/index.html`
- **Use this when you have Firebase configuration in place**
- Will not overwrite your Firebase settings

### `make wasm`
- Builds WASM version with basic index.html
- **May overwrite existing Firebase configuration**
- Use only for clean builds or when Firebase is not needed

### `make web`
- Builds and serves the game using the development server
- Now respects existing `web/index.html` (won't overwrite Firebase config)
- Includes proper WASM headers for development

### `make build`
- Builds native desktop version (no Firebase/leaderboard)

### `make run`
- Builds and runs desktop version

## Firebase Setup Workflow

1. **Set up Firebase project** (see FIREBASE_SETUP.md)

2. **Configure web/index.html**:
   ```bash
   # Edit web/index.html and replace placeholder Firebase config
   # with your actual Firebase project details
   ```

3. **Build WASM version**:
   ```bash
   make wasm-static  # Preserves Firebase config
   ```

4. **Deploy to web server**:
   ```bash
   # Upload web/ directory contents to your web host
   # Or use GitHub Pages, Netlify, Vercel, etc.
   ```

## Development Workflow

For development with Firebase:
```bash
# Initial setup
make wasm-static          # Build with Firebase config preserved
make web                  # Start development server (preserves config)
```

For development without Firebase:
```bash
make web                  # Uses generated index.html without Firebase
```

## Important Notes

- **Always use `make wasm-static`** when you have Firebase configured
- The `web/index.html` file is now **static** and won't be overwritten
- Firebase configuration only affects WASM builds (web version)
- Desktop builds (`make build`, `make run`) ignore Firebase entirely
- If `web/index.html` gets corrupted, you can rebuild it with `make wasm` then re-add Firebase config

## File Structure

```
web/
├── index.html         # Static HTML with Firebase config (don't delete!)
├── sailing.wasm       # Game binary (rebuilt each time)
├── wasm_exec.js       # Go WASM runtime (copied from Go installation)
```

## Troubleshooting

- **"Firebase not initialized"**: Check your Firebase config in `web/index.html`
- **Index.html overwritten**: Use `make wasm-static` instead of `make wasm`
- **Game works but no leaderboard**: Firebase config may be missing or incorrect
- **Build fails**: Try `make clean` then rebuild
