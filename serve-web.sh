#!/bin/bash

echo "🚢 Building and serving Go Sailing Game for Web"
echo "================================================"

# Build and run the WASM server
echo "Building WASM server..."
go run ./cmd/wasm_server/
