package ui

import (
	"math/rand"
	"strings"
	"time"

	"github.com/0xjuanma/golazo/internal/constants"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// SpinnerTickInterval is the unified tick rate for all spinners (70ms ≈ 14 fps).
// This balances smooth animation with keyboard responsiveness.
const SpinnerTickInterval = 70 * time.Millisecond

// TickMsg is the unified message type for all spinner updates.
// Only ONE tick chain should exist at any time to prevent message queue flooding.
type TickMsg struct{}

// SpinnerTick returns a command that generates a TickMsg after the standard interval.
// This is the ONLY function that should create spinner ticks - ensures single tick chain.
func SpinnerTick() tea.Cmd {
	return tea.Tick(SpinnerTickInterval, func(time.Time) tea.Msg {
		return TickMsg{}
	})
}

// RandomCharSpinner is a custom spinner that displays a wave of random characters.
// Note: Spinners do NOT self-tick. The app manages the tick chain centrally.
type RandomCharSpinner struct {
	charPool   []rune // Pool of characters to choose from
	display    []rune // Currently displayed characters (wave buffer)
	width      int
	startColor colorful.Color // Gradient start color (cyan)
	endColor   colorful.Color // Gradient end color (red)
}

// NewRandomCharSpinner creates a new random character spinner.
func NewRandomCharSpinner() *RandomCharSpinner {
	// Extended Latin character set with subtle symbols for smooth, sophisticated animation
	// Includes: uppercase, lowercase, European accented letters, numbers, subtle symbols
	charPool := []rune(
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" + // Basic Latin
			"ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþÿ" + // Extended Latin
			"0123456789" + // Numbers
			"×÷±≈∞≠√" + // Mathematical
			"→←↑↓↔" + // Arrows
			"€£¥$" + // Currency
			"·•°§", // Clean punctuation
	)

	// Create gradient: cyan to red (high energy theme)
	startColor, _ := colorful.Hex(constants.GradientStartColor) // Bright cyan
	endColor, _ := colorful.Hex(constants.GradientEndColor)     // Bright red

	width := 20

	// Initialize display buffer with random characters
	display := make([]rune, width)
	for i := range display {
		display[i] = charPool[rand.Intn(len(charPool))]
	}

	return &RandomCharSpinner{
		charPool:   charPool,
		display:    display,
		width:      width,
		startColor: startColor,
		endColor:   endColor,
	}
}

// Tick advances the spinner animation - randomizes all characters for trendy effect.
// Does NOT return a tick command - the app manages the tick chain.
func (r *RandomCharSpinner) Tick() {
	// Ensure display buffer matches width
	if len(r.display) != r.width {
		r.display = make([]rune, r.width)
	}

	// Randomize all characters each tick for dynamic, trendy effect
	for i := range r.display {
		r.display[i] = r.charPool[rand.Intn(len(r.charPool))]
	}
}

// View renders the spinner with gradient colors.
func (r *RandomCharSpinner) View() string {
	if r.width <= 0 {
		r.width = 20
	}

	// Ensure display buffer exists
	if len(r.display) == 0 {
		r.display = make([]rune, r.width)
		for i := range r.display {
			r.display[i] = r.charPool[rand.Intn(len(r.charPool))]
		}
	}

	// Apply gradient to each character
	var result strings.Builder
	for i, char := range r.display {
		ratio := float64(i) / float64(r.width-1)
		color := r.startColor.BlendLab(r.endColor, ratio)
		hexColor := color.Hex()
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor))
		result.WriteString(charStyle.Render(string(char)))
	}

	return result.String()
}

// SetWidth sets the width of the spinner and resizes the display buffer.
func (r *RandomCharSpinner) SetWidth(width int) {
	if width == r.width {
		return
	}
	r.width = width

	// Resize display buffer with random characters
	r.display = make([]rune, width)
	for i := range r.display {
		r.display[i] = r.charPool[rand.Intn(len(r.charPool))]
	}
}
