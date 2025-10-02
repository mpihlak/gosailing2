//go:build js && wasm

package game

import (
	"syscall/js"
	"time"
)

// FirebaseClient handles Firebase Firestore operations in WASM
type FirebaseClient struct {
	firestore js.Value
	isReady   bool
}

// NewFirebaseClient creates a new Firebase client for WASM
func NewFirebaseClient() *FirebaseClient {
	return &FirebaseClient{
		isReady: false,
	}
}

// Initialize sets up the Firebase connection
func (fc *FirebaseClient) Initialize() {
	// Get Firebase Firestore instance from JavaScript
	firebase := js.Global().Get("firebase")
	if firebase.IsUndefined() {
		return
	}

	fc.firestore = firebase.Call("firestore")
	fc.isReady = true
}

// SubmitScore submits a race result to Firestore
func (fc *FirebaseClient) SubmitScore(result *RaceResult, callback func(bool, string)) {
	if !fc.isReady {
		fc.Initialize()
	}

	if !fc.isReady {
		callback(false, "Firebase not initialized")
		return
	}

	// Convert result to JavaScript object
	resultData := map[string]interface{}{
		"player_name":       result.PlayerName,
		"race_time_seconds": result.RaceTimeSeconds,
		"seconds_late":      result.SecondsLate,
		"speed_percentage":  result.SpeedPercentage,
		"mark_rounded":      result.MarkRounded,
		"timestamp":         result.Timestamp.Unix(),
	}

	// Create JavaScript object
	jsData := js.ValueOf(resultData)

	// Submit to Firestore collection "race_results"
	collection := fc.firestore.Call("collection", "race_results")

	// Create success callback
	successCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		callback(true, "")
		return nil
	})
	defer successCallback.Release()

	// Create error callback
	errorCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errorMsg := "Failed to submit score"
		if len(args) > 0 && !args[0].IsUndefined() {
			if message := args[0].Get("message"); !message.IsUndefined() {
				errorMsg = message.String()
			}
		}
		callback(false, errorMsg)
		return nil
	})
	defer errorCallback.Release()

	// Add document
	promise := collection.Call("add", jsData)
	promise.Call("then", successCallback)
	promise.Call("catch", errorCallback)
}

// GetLeaderboard retrieves the top race results from Firestore
func (fc *FirebaseClient) GetLeaderboard(callback func([]RaceResult, string)) {
	if !fc.isReady {
		fc.Initialize()
	}

	if !fc.isReady {
		callback(nil, "Firebase not initialized")
		return
	}

	// Query Firestore for race results, ordered by race time, limited to 50
	collection := fc.firestore.Call("collection", "race_results")
	query := collection.Call("where", "mark_rounded", "==", true)
	query = query.Call("orderBy", "race_time_seconds", "asc")
	query = query.Call("limit", 50)

	// Create success callback
	successCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			callback(nil, "No data received")
			return nil
		}

		querySnapshot := args[0]
		results := make([]RaceResult, 0)

		// Process each document
		docs := querySnapshot.Get("docs")
		docsLength := docs.Get("length").Int()

		for i := 0; i < docsLength; i++ {
			doc := docs.Index(i)
			data := doc.Call("data")

			// Extract data from JavaScript object
			result := RaceResult{
				PlayerName:      getStringValue(data, "player_name"),
				RaceTimeSeconds: getFloatValue(data, "race_time_seconds"),
				SecondsLate:     getFloatValue(data, "seconds_late"),
				SpeedPercentage: getFloatValue(data, "speed_percentage"),
				MarkRounded:     getBoolValue(data, "mark_rounded"),
				Timestamp:       time.Unix(int64(getFloatValue(data, "timestamp")), 0),
			}

			results = append(results, result)
		}

		callback(results, "")
		return nil
	})
	defer successCallback.Release()

	// Create error callback
	errorCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errorMsg := "Failed to load leaderboard"
		if len(args) > 0 && !args[0].IsUndefined() {
			if message := args[0].Get("message"); !message.IsUndefined() {
				errorMsg = message.String()
			}
		}
		callback(nil, errorMsg)
		return nil
	})
	defer errorCallback.Release()

	// Execute query
	promise := query.Call("get")
	promise.Call("then", successCallback)
	promise.Call("catch", errorCallback)
}

// Helper functions to safely extract values from JavaScript objects
func getStringValue(jsObj js.Value, key string) string {
	val := jsObj.Get(key)
	if val.IsUndefined() || val.IsNull() {
		return ""
	}
	return val.String()
}

func getFloatValue(jsObj js.Value, key string) float64 {
	val := jsObj.Get(key)
	if val.IsUndefined() || val.IsNull() {
		return 0.0
	}
	return val.Float()
}

func getBoolValue(jsObj js.Value, key string) bool {
	val := jsObj.Get(key)
	if val.IsUndefined() || val.IsNull() {
		return false
	}
	return val.Bool()
}
