#!/bin/bash

echo "🚢 Building and serving Go Sailing Game for Web"
echo "================================================"

# Build and run the WASM server
echo "Building WASM server..."
go build -o wasm_server ./cmd/wasm_server

echo "Starting WASM server (this will build the game and serve it)..."
./wasm_server

# Clean up
rm -f wasm_server
