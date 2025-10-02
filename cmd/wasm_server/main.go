package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	port    = ":8080"
	wasmDir = "web"
)

func main() {
	// Create web directory if it doesn't exist
	if err := os.MkdirAll(wasmDir, 0755); err != nil {
		log.Fatal("Failed to create web directory:", err)
	}

	// Build WASM version
	fmt.Println("Building WASM version...")
	if err := buildWASM(); err != nil {
		log.Fatal("Failed to build WASM:", err)
	}

	// Copy required files
	fmt.Println("Copying required files...")
	if err := copyWASMFiles(); err != nil {
		log.Fatal("Failed to copy files:", err)
	}

	// Create HTML file
	fmt.Println("Creating HTML file...")
	if err := createHTMLFile(); err != nil {
		log.Fatal("Failed to create HTML file:", err)
	}

	// Setup HTTP server with proper headers for WASM
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for WASM
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")

		// Handle WASM files with correct MIME type
		if filepath.Ext(r.URL.Path) == ".wasm" {
			w.Header().Set("Content-Type", "application/wasm")
		}

		// Serve files from web directory
		http.FileServer(http.Dir(wasmDir)).ServeHTTP(w, r)
	})

	fmt.Printf("🚢 Sailing game server starting on http://localhost%s\n", port)
	fmt.Printf("📁 Serving files from: %s/\n", wasmDir)
	fmt.Println("🌐 Open your browser to play the game!")

	// Try to open browser automatically
	openBrowser(fmt.Sprintf("http://localhost%s", port))

	log.Fatal(http.ListenAndServe(port, nil))
}

func buildWASM() error {
	cmd := exec.Command("go", "build", "-o", filepath.Join(wasmDir, "sailing.wasm"), "./cmd/gosailing")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyWASMFiles() error {
	// Copy wasm_exec.js from Go installation
	goRoot := runtime.GOROOT()
	wasmExecPath := filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")

	// Read the source file
	data, err := os.ReadFile(wasmExecPath)
	if err != nil {
		return fmt.Errorf("failed to read wasm_exec.js: %v", err)
	}

	// Write to web directory
	destPath := filepath.Join(wasmDir, "wasm_exec.js")
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to copy wasm_exec.js: %v", err)
	}

	return nil
}

func createHTMLFile() error {
	htmlPath := filepath.Join(wasmDir, "index.html")

	// Check if index.html already exists, don't overwrite it
	if _, err := os.Stat(htmlPath); err == nil {
		fmt.Println("index.html already exists, keeping existing version")
		return nil
	}

	fmt.Println("Creating new index.html (no existing file found)")
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Go Sailing!</title>
    <style>
        body {
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #001122, #003366);
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            color: white;
        }
        .header {
            text-align: center;
            margin-bottom: 20px;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
            color: #66ccff;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.5);
        }
        .header p {
            margin: 10px 0;
            color: #aaaaaa;
            font-size: 1.1em;
        }
        canvas {
            border: 3px solid #4488cc;
            border-radius: 8px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.3);
            margin: 20px;
        }
        .controls {
            background: rgba(0,0,0,0.3);
            padding: 20px;
            border-radius: 8px;
            text-align: center;
            max-width: 600px;
            margin: 20px;
        }
        .controls h3 {
            margin-top: 0;
            color: #66ccff;
            font-size: 1.3em;
        }
        .control-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-top: 15px;
        }
        .control-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 12px;
            background: rgba(255,255,255,0.1);
            border-radius: 4px;
            border-left: 3px solid #66ccff;
        }
        .key {
            background: #333;
            padding: 4px 8px;
            border-radius: 4px;
            font-family: monospace;
            font-size: 0.9em;
            color: #66ccff;
        }
        .loading {
            color: #66ccff;
            font-size: 1.2em;
            text-align: center;
            padding: 40px;
        }
        .error {
            color: #ff6666;
            background: rgba(255,0,0,0.1);
            padding: 20px;
            border-radius: 8px;
            border: 1px solid #ff6666;
        }
        @keyframes wave {
            0%, 100% { transform: rotate(0deg); }
            25% { transform: rotate(10deg); }
            75% { transform: rotate(-10deg); }
        }
        .wave {
            display: inline-block;
            animation: wave 2s infinite;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1><span class="wave">⛵</span> Go Sailing!</h1>
        <p>A sailing simulation built with Go and WebAssembly</p>
    </div>

    <div class="loading" id="loading">
        <div>Loading sailing game...</div>
        <div style="margin-top: 10px; font-size: 0.9em; color: #888;">
            Downloading and initializing WebAssembly module
        </div>
    </div>

    <div class="controls" style="display: none;" id="controls">
    </div>

    <div class="error" style="display: none;" id="error">
        <h3>⚠️ Failed to Load Game</h3>
        <p>There was an error loading the sailing game. Please check the browser console for details.</p>
        <div id="error-details" style="margin-top: 10px; font-family: monospace; font-size: 0.8em;"></div>
    </div>

    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();

        // Add some debug logging
        console.log('Starting WASM load...');

        WebAssembly.instantiateStreaming(fetch("sailing.wasm"), go.importObject)
            .then((result) => {
                console.log('WASM loaded successfully');
                document.getElementById('loading').style.display = 'none';
                document.getElementById('controls').style.display = 'block';

                // Run the game
                go.run(result.instance);
            })
            .catch((err) => {
                console.error('Failed to load WASM:', err);
                document.getElementById('loading').style.display = 'none';
                document.getElementById('error').style.display = 'block';
                document.getElementById('error-details').textContent = err.toString();
            });
    </script>
</body>
</html>`

	return os.WriteFile(htmlPath, []byte(html), 0644)
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)

	// Don't wait for the command to finish and ignore errors
	go exec.Command(cmd, args...).Run()
}
