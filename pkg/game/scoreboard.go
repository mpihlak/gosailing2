package game

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// RaceResult represents a single race completion record
type RaceResult struct {
	PlayerName      string    `json:"player_name"`
	RaceTimeSeconds float64   `json:"race_time_seconds"`
	SecondsLate     float64   `json:"seconds_late"`
	SpeedPercentage float64   `json:"speed_percentage"`
	MarkRounded     bool      `json:"mark_rounded"`
	Timestamp       time.Time `json:"timestamp"`
}

// LeaderboardEntry represents a formatted leaderboard entry for display
type LeaderboardEntry struct {
	Rank          int
	PlayerName    string
	RaceTime      string
	SecondsLate   string
	IsCurrentRace bool // Highlight the most recent race result
}

// Scoreboard manages the leaderboard display and player name input
type Scoreboard struct {
	// State management
	isVisible bool
	state     ScoreboardState

	// Player input
	playerName    string
	nameSubmitted bool

	// Leaderboard data
	leaderboard      []LeaderboardEntry
	currentRaceEntry *LeaderboardEntry // Current race entry (may be outside top 10)
	currentResult    *RaceResult

	// UI state
	cursorBlink bool
	lastBlink   time.Time
	submitError string
	isLoading   bool

	// Firebase integration (WASM only)
	firebase *FirebaseClient
}

type ScoreboardState int

const (
	StateEnterName ScoreboardState = iota
	StateDisplayLeaderboard
	StateError
)

// NewScoreboard creates a new scoreboard instance
func NewScoreboard() *Scoreboard {
	var firebase *FirebaseClient
	if IsWASM() {
		firebase = NewFirebaseClient()
	}

	return &Scoreboard{
		isVisible:        false,
		state:            StateEnterName,
		playerName:       "",
		leaderboard:      make([]LeaderboardEntry, 0),
		currentRaceEntry: nil,
		firebase:         firebase,
		lastBlink:        time.Now(),
	}
}

// Show displays the scoreboard with the given race result
func (s *Scoreboard) Show(result *RaceResult) {
	s.isVisible = true
	s.state = StateEnterName
	s.currentResult = result
	s.playerName = ""
	s.nameSubmitted = false
	s.submitError = ""
	s.isLoading = false
}

// Hide closes the scoreboard
func (s *Scoreboard) Hide() {
	s.isVisible = false
	s.state = StateEnterName
	s.playerName = ""
	s.nameSubmitted = false
	s.leaderboard = make([]LeaderboardEntry, 0)
	s.currentRaceEntry = nil
}

// IsVisible returns whether the scoreboard is currently displayed
func (s *Scoreboard) IsVisible() bool {
	return s.isVisible
}

// IsCapturingInput returns whether the scoreboard is currently capturing text input
func (s *Scoreboard) IsCapturingInput() bool {
	return s.isVisible && s.state == StateEnterName
}

// Update handles input and state updates
func (s *Scoreboard) Update() {
	if !s.isVisible {
		return
	}

	// Handle cursor blinking
	if time.Since(s.lastBlink) > 500*time.Millisecond {
		s.cursorBlink = !s.cursorBlink
		s.lastBlink = time.Now()
	}

	switch s.state {
	case StateEnterName:
		s.updateNameInput()
	case StateDisplayLeaderboard:
		s.updateLeaderboardDisplay()
	}
}

// updateNameInput handles player name entry
func (s *Scoreboard) updateNameInput() {
	// Handle text input
	inputChars := ebiten.AppendInputChars(nil)
	for _, ch := range inputChars {
		if len(s.playerName) < 20 && isValidNameChar(ch) {
			s.playerName += string(ch)
		}
	}

	// Handle backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(s.playerName) > 0 {
		s.playerName = s.playerName[:len(s.playerName)-1]
	}

	// Handle enter key to submit name
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(strings.TrimSpace(s.playerName)) > 0 {
		s.submitScore()
	}

	// Handle escape to skip submission (standalone mode)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if IsWASM() {
			// In WASM, show leaderboard without submitting
			s.loadLeaderboard()
		} else {
			// In standalone, just close
			s.Hide()
		}
	}
}

// updateLeaderboardDisplay handles leaderboard viewing
func (s *Scoreboard) updateLeaderboardDisplay() {
	// Handle escape or enter to close
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		s.Hide()
	}
}

// submitScore submits the current race result to Firebase (WASM only)
func (s *Scoreboard) submitScore() {
	name := strings.TrimSpace(s.playerName)
	if len(name) == 0 {
		return
	}

	s.currentResult.PlayerName = name
	s.currentResult.Timestamp = time.Now()

	if IsWASM() && s.firebase != nil {
		s.isLoading = true
		s.submitError = ""

		// Submit to Firebase
		s.firebase.SubmitScore(s.currentResult, func(success bool, err string) {
			s.isLoading = false
			if success {
				s.nameSubmitted = true
				s.loadLeaderboard()
			} else {
				s.submitError = err
			}
		})
	} else {
		// Standalone mode - just show local leaderboard
		s.nameSubmitted = true
		s.createLocalLeaderboard()
		s.state = StateDisplayLeaderboard
	}
}

// loadLeaderboard loads the leaderboard from Firebase (WASM only)
func (s *Scoreboard) loadLeaderboard() {
	if IsWASM() && s.firebase != nil {
		s.isLoading = true
		s.firebase.GetLeaderboard(func(results []RaceResult, err string) {
			s.isLoading = false
			if err != "" {
				s.submitError = err
				s.state = StateError
			} else {
				s.createLeaderboard(results)
				s.state = StateDisplayLeaderboard
			}
		})
	} else {
		s.createLocalLeaderboard()
		s.state = StateDisplayLeaderboard
	}
}

// createLeaderboard creates leaderboard entries from race results
func (s *Scoreboard) createLeaderboard(results []RaceResult) {
	// Filter completed races only
	completed := make([]RaceResult, 0)
	for _, result := range results {
		if result.MarkRounded {
			completed = append(completed, result)
		}
	}

	// Sort by race time (ascending)
	sort.Slice(completed, func(i, j int) bool {
		return completed[i].RaceTimeSeconds < completed[j].RaceTimeSeconds
	})

	// Find current race in the completed results
	var currentRaceResult *RaceResult
	var currentRaceRank int
	if s.currentResult != nil && s.currentResult.MarkRounded {
		for i, result := range completed {
			// Match by player name and exact race time (to identify the specific race)
			if result.PlayerName == s.currentResult.PlayerName &&
				fmt.Sprintf("%.2f", result.RaceTimeSeconds) == fmt.Sprintf("%.2f", s.currentResult.RaceTimeSeconds) {
				currentRaceResult = &result
				currentRaceRank = i + 1
				break
			}
		}
	}

	// Create display entries (top 10)
	s.leaderboard = make([]LeaderboardEntry, 0)
	maxEntries := 10
	if len(completed) < maxEntries {
		maxEntries = len(completed)
	}

	for i := 0; i < maxEntries; i++ {
		result := completed[i]

		// Format race time
		minutes := int(result.RaceTimeSeconds) / 60
		seconds := int(result.RaceTimeSeconds) % 60
		centiseconds := int((result.RaceTimeSeconds - float64(int(result.RaceTimeSeconds))) * 100)
		raceTimeStr := fmt.Sprintf("%02d:%02d.%02d", minutes, seconds, centiseconds)

		// Format seconds late
		lateStr := fmt.Sprintf("%.1f", result.SecondsLate)
		if result.SecondsLate < 0 {
			lateStr = "Early"
		}

		// Check if this is the current race
		isCurrentRace := currentRaceResult != nil &&
			result.PlayerName == currentRaceResult.PlayerName &&
			fmt.Sprintf("%.2f", result.RaceTimeSeconds) == fmt.Sprintf("%.2f", currentRaceResult.RaceTimeSeconds)

		entry := LeaderboardEntry{
			Rank:          i + 1,
			PlayerName:    result.PlayerName,
			RaceTime:      raceTimeStr,
			SecondsLate:   lateStr,
			IsCurrentRace: isCurrentRace,
		}

		s.leaderboard = append(s.leaderboard, entry)
	}

	// Create separate current race entry if it's outside top 10
	s.currentRaceEntry = nil
	if currentRaceResult != nil && currentRaceRank > 10 {
		minutes := int(currentRaceResult.RaceTimeSeconds) / 60
		seconds := int(currentRaceResult.RaceTimeSeconds) % 60
		centiseconds := int((currentRaceResult.RaceTimeSeconds - float64(int(currentRaceResult.RaceTimeSeconds))) * 100)
		raceTimeStr := fmt.Sprintf("%02d:%02d.%02d", minutes, seconds, centiseconds)

		lateStr := fmt.Sprintf("%.1f", currentRaceResult.SecondsLate)
		if currentRaceResult.SecondsLate < 0 {
			lateStr = "Early"
		}

		s.currentRaceEntry = &LeaderboardEntry{
			Rank:          currentRaceRank,
			PlayerName:    currentRaceResult.PlayerName,
			RaceTime:      raceTimeStr,
			SecondsLate:   lateStr,
			IsCurrentRace: true,
		}
	}
} // createLocalLeaderboard creates a local leaderboard for standalone mode
func (s *Scoreboard) createLocalLeaderboard() {
	if s.currentResult == nil {
		return
	}

	// Format current player's time
	minutes := int(s.currentResult.RaceTimeSeconds) / 60
	seconds := int(s.currentResult.RaceTimeSeconds) % 60
	centiseconds := int((s.currentResult.RaceTimeSeconds - float64(int(s.currentResult.RaceTimeSeconds))) * 100)
	raceTimeStr := fmt.Sprintf("%02d:%02d.%02d", minutes, seconds, centiseconds)

	lateStr := fmt.Sprintf("%.1f", s.currentResult.SecondsLate)
	if s.currentResult.SecondsLate < 0 {
		lateStr = "Early"
	}

	s.leaderboard = []LeaderboardEntry{
		{
			Rank:          1,
			PlayerName:    s.currentResult.PlayerName,
			RaceTime:      raceTimeStr,
			SecondsLate:   lateStr,
			IsCurrentRace: true,
		},
	}
	s.currentRaceEntry = nil // No separate entry needed for local mode
}

// Draw renders the scoreboard overlay
func (s *Scoreboard) Draw(screen *ebiten.Image) {
	if !s.isVisible {
		return
	}

	// Draw semi-transparent overlay
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 200}, false)

	switch s.state {
	case StateEnterName:
		s.drawNameEntry(screen)
	case StateDisplayLeaderboard:
		s.drawLeaderboard(screen)
	case StateError:
		s.drawError(screen)
	}
}

// drawNameEntry draws the player name entry screen
func (s *Scoreboard) drawNameEntry(screen *ebiten.Image) {
	bounds := screen.Bounds()
	centerX := bounds.Dx() / 2
	centerY := bounds.Dy() / 2

	// Title
	title := "🏆 Race Complete! 🏆"
	ebitenutil.DebugPrintAt(screen, title, centerX-80, centerY-120)

	// Race time display
	if s.currentResult != nil {
		minutes := int(s.currentResult.RaceTimeSeconds) / 60
		seconds := int(s.currentResult.RaceTimeSeconds) % 60
		centiseconds := int((s.currentResult.RaceTimeSeconds - float64(int(s.currentResult.RaceTimeSeconds))) * 100)
		timeText := fmt.Sprintf("Your Time: %02d:%02d.%02d", minutes, seconds, centiseconds)
		ebitenutil.DebugPrintAt(screen, timeText, centerX-70, centerY-90)
	}

	// Name entry prompt
	prompt := "Enter your name:"
	ebitenutil.DebugPrintAt(screen, prompt, centerX-60, centerY-40)

	// Input field background
	fieldX := centerX - 100
	fieldY := centerY - 10
	fieldWidth := 200
	fieldHeight := 25

	vector.DrawFilledRect(screen, float32(fieldX), float32(fieldY), float32(fieldWidth), float32(fieldHeight), color.RGBA{255, 255, 255, 255}, false)
	vector.StrokeRect(screen, float32(fieldX), float32(fieldY), float32(fieldWidth), float32(fieldHeight), 2, color.RGBA{100, 100, 100, 255}, false)

	// Player name text
	nameText := s.playerName
	if s.cursorBlink {
		nameText += "|"
	}
	ebitenutil.DebugPrintAt(screen, nameText, fieldX+5, fieldY+5)

	// Instructions
	var instructions string
	if IsWASM() {
		instructions = "Press ENTER to submit • ESC to view leaderboard only"
	} else {
		instructions = "Press ENTER to continue • ESC to skip"
	}
	ebitenutil.DebugPrintAt(screen, instructions, centerX-130, centerY+40)

	// Loading indicator
	if s.isLoading {
		ebitenutil.DebugPrintAt(screen, "Submitting score...", centerX-60, centerY+70)
	}

	// Error message
	if s.submitError != "" {
		ebitenutil.DebugPrintAt(screen, "Error: "+s.submitError, centerX-80, centerY+70)
	}
}

// drawLeaderboard draws the leaderboard display
func (s *Scoreboard) drawLeaderboard(screen *ebiten.Image) {
	bounds := screen.Bounds()
	centerX := bounds.Dx() / 2
	startY := 100

	// Title
	title := "🏆 LEADERBOARD 🏆"
	ebitenutil.DebugPrintAt(screen, title, centerX-80, startY-30)

	// Headers
	headerY := startY + 20
	ebitenutil.DebugPrintAt(screen, "Rank", centerX-180, headerY)
	ebitenutil.DebugPrintAt(screen, "Name", centerX-120, headerY)
	ebitenutil.DebugPrintAt(screen, "Time", centerX-20, headerY)
	ebitenutil.DebugPrintAt(screen, "Late", centerX+60, headerY)

	// Draw line under headers
	lineY := float32(headerY + 15)
	vector.StrokeLine(screen, float32(centerX-190), lineY, float32(centerX+120), lineY, 1, color.RGBA{255, 255, 255, 255}, false)

	// Leaderboard entries
	for i, entry := range s.leaderboard {
		entryY := startY + 50 + (i * 25)

		// Highlight current race
		if entry.IsCurrentRace {
			highlightY := float32(entryY - 2)
			vector.DrawFilledRect(screen, float32(centerX-195), highlightY, 320, 20, color.RGBA{173, 216, 230, 150}, false)
		}

		// Draw entry data
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", entry.Rank), centerX-180, entryY)

		// Truncate long names
		displayName := entry.PlayerName
		if len(displayName) > 12 {
			displayName = displayName[:12] + "..."
		}
		ebitenutil.DebugPrintAt(screen, displayName, centerX-120, entryY)
		ebitenutil.DebugPrintAt(screen, entry.RaceTime, centerX-20, entryY)
		ebitenutil.DebugPrintAt(screen, entry.SecondsLate, centerX+60, entryY)
	}

	// Draw separator and current race entry if it's outside top 10
	if s.currentRaceEntry != nil {
		separatorY := startY + 50 + (len(s.leaderboard) * 25) + 10

		// Draw separator dots
		ebitenutil.DebugPrintAt(screen, "...", centerX-10, separatorY)

		// Draw current race entry
		entryY := separatorY + 20

		// Highlight current race with light blue background
		highlightY := float32(entryY - 2)
		vector.DrawFilledRect(screen, float32(centerX-195), highlightY, 320, 20, color.RGBA{173, 216, 230, 150}, false)

		// Draw entry data
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", s.currentRaceEntry.Rank), centerX-180, entryY)

		// Truncate long names
		displayName := s.currentRaceEntry.PlayerName
		if len(displayName) > 12 {
			displayName = displayName[:12] + "..."
		}
		ebitenutil.DebugPrintAt(screen, displayName, centerX-120, entryY)
		ebitenutil.DebugPrintAt(screen, s.currentRaceEntry.RaceTime, centerX-20, entryY)
		ebitenutil.DebugPrintAt(screen, s.currentRaceEntry.SecondsLate, centerX+60, entryY)
	} // Instructions
	var instructions string
	if IsWASM() {
		instructions = "Press ENTER or ESC to continue • Data saved online"
	} else {
		instructions = "Press ENTER or ESC to continue • Local data only"
	}
	ebitenutil.DebugPrintAt(screen, instructions, centerX-140, bounds.Dy()-50)
}

// drawError draws the error screen
func (s *Scoreboard) drawError(screen *ebiten.Image) {
	bounds := screen.Bounds()
	centerX := bounds.Dx() / 2
	centerY := bounds.Dy() / 2

	// Error message
	ebitenutil.DebugPrintAt(screen, "⚠️ Error loading leaderboard", centerX-100, centerY-30)
	ebitenutil.DebugPrintAt(screen, s.submitError, centerX-100, centerY)
	ebitenutil.DebugPrintAt(screen, "Press ESC to continue", centerX-70, centerY+30)
}

// isValidNameChar checks if a character is valid for player names
func isValidNameChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == ' ' || ch == '-' || ch == '_'
}
