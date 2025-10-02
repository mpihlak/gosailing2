# Go Sailing Game Makefile

.PHONY: build run web clean help

# Default target
help:
	@echo "Go Sailing Game Build Commands:"
	@echo "  make build       - Build native desktop version"
	@echo "  make run         - Build and run desktop version"
	@echo "  make web         - Build and serve web version (WASM)"
	@echo "  make wasm        - Build WASM only (may overwrite index.html)"
	@echo "  make wasm-static - Build WASM preserving existing index.html"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make help        - Show this help"

# Build desktop version
build:
	@echo "🚢 Building desktop version..."
	go build -o sailing ./cmd/gosailing

# Run desktop version
run: build
	@echo "🚢 Running desktop version..."
	./sailing

# Build and serve web version
web:
	@echo "🌐 Building and serving web version..."
	@chmod +x serve-web.sh
	@./serve-web.sh

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f sailing wasm_server
	rm -rf web/
	@echo "Clean complete!"

# Build WASM version only (without serving)
wasm:
	@echo "🕸️ Building WASM version..."
	@mkdir -p web
	GOOS=js GOARCH=wasm go build -o web/sailing.wasm ./cmd/gosailing
	@cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/
	@echo "WASM build complete in web/ directory"

# Build WASM with static index.html (preserves Firebase config)
wasm-static:
	@echo "🕸️ Building WASM with static index.html..."
	@mkdir -p web
	GOOS=js GOARCH=wasm go build -o web/sailing.wasm ./cmd/gosailing
	@cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/
	@if [ ! -f web/index.html ]; then \
		echo "⚠️  No index.html found in web/ directory"; \
		echo "Please ensure web/index.html exists with Firebase configuration"; \
	else \
		echo "✅ Using existing index.html (Firebase config preserved)"; \
	fi
	@echo "WASM build complete in web/ directory"
