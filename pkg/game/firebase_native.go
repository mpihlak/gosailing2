//go:build !js || !wasm

package game

// FirebaseClient stub for non-WASM builds
type FirebaseClient struct{}

// NewFirebaseClient creates a stub Firebase client for non-WASM builds
func NewFirebaseClient() *FirebaseClient {
	return &FirebaseClient{}
}

// Initialize is a no-op for non-WASM builds
func (fc *FirebaseClient) Initialize() {}

// SubmitScore is a no-op for non-WASM builds
func (fc *FirebaseClient) SubmitScore(result *RaceResult, callback func(bool, string)) {
	callback(false, "Firebase not available in standalone mode")
}

// GetLeaderboard is a no-op for non-WASM builds
func (fc *FirebaseClient) GetLeaderboard(callback func([]RaceResult, string)) {
	callback(nil, "Firebase not available in standalone mode")
}
