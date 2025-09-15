#!/bin/bash

echo "🚢 Building Go Sailing Game for GitHub Pages"
echo "============================================="

# Ensure web directory exists
mkdir -p web

# Build WASM version
echo "Building WASM version..."
GOOS=js GOARCH=wasm go build -o web/sailing.wasm ./cmd/gosailing

if [ $? -ne 0 ]; then
    echo "❌ Failed to build WASM"
    exit 1
fi

# Copy wasm_exec.js from Go installation
echo "Copying wasm_exec.js..."
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/

if [ $? -ne 0 ]; then
    echo "❌ Failed to copy wasm_exec.js"
    exit 1
fi

# Verify index.html exists
if [ ! -f "web/index.html" ]; then
    echo "❌ index.html not found in web/ directory"
    exit 1
fi

echo "✅ Build complete!"
echo ""
echo "📁 Files in web/ directory:"
ls -la web/
echo ""
echo "🌐 Ready for GitHub Pages deployment!"
echo ""
echo "To deploy to GitHub Pages:"
echo "1. Push this repository to GitHub"
echo "2. Go to repository Settings > Pages"
echo "3. Set source to 'GitHub Actions'"
echo "4. The workflow will automatically build and deploy"
echo ""
echo "Or commit and push the web/ directory contents to a gh-pages branch"
