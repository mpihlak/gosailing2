//go:build js && wasm

package game

// IsWASM returns true when running in WebAssembly environment
func IsWASM() bool {
	return true
}