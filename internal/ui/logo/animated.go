// Package logo renders a GOLAZO wordmark in a stylized way.
package logo

import (
	"math/rand"
	"strings"
)

// AnimationType defines the style of animation for the logo reveal.
type AnimationType int

const (
	// AnimationWave reveals characters left-to-right with staggered line delays (default)
	AnimationWave AnimationType = iota
	// AnimationTypewriter reveals character-by-character across all lines sequentially
	AnimationTypewriter
	// AnimationLineByLine reveals full lines one at a time
	AnimationLineByLine
	// AnimationCenterOut reveals from center of each line outward
	AnimationCenterOut
	// AnimationRandom reveals characters in random order
	AnimationRandom
)

// String returns the name of the animation type.
func (a AnimationType) String() string {
	switch a {
	case AnimationWave:
		return "Wave"
	case AnimationTypewriter:
		return "Typewriter"
	case AnimationLineByLine:
		return "LineByLine"
	case AnimationCenterOut:
		return "CenterOut"
	case AnimationRandom:
		return "Random"
	default:
		return "Unknown"
	}
}

// AllAnimationTypes returns all available animation types.
func AllAnimationTypes() []AnimationType {
	return []AnimationType{
		AnimationWave,
		AnimationTypewriter,
		AnimationLineByLine,
		AnimationCenterOut,
		AnimationRandom,
	}
}

// AnimatedLogo wraps a rendered logo and reveals it progressively.
// It uses logo.Render() internally and does not modify the render logic.
type AnimatedLogo struct {
	fullContent   string        // Complete output from logo.Render()
	lines         []string      // Split lines for reveal
	revealedCols  []int         // Per-line reveal progress
	revealedChars [][]bool      // Per-character reveal state (for random animation)
	charsPerTick  int           // Characters to reveal per tick (derived from duration)
	waveOffset    int           // Stagger between lines starting reveal (in chars)
	totalTicks    int           // Total ticks for animation
	currentTick   int           // Current tick count
	complete      bool          // Animation finished flag
	playCount     int           // Number of times animation has played
	maxPlays      int           // Max animations (0 = infinite, 1 = once)
	maxLineWidth  int           // Width of the longest line
	animationType AnimationType // Type of animation to use
	totalChars    int           // Total visible characters across all lines
	revealedCount int           // Count of revealed characters (for typewriter/random)
	lineWidths    []int         // Visible width of each line
}

// NewAnimatedLogo creates a new animated logo that wraps logo.Render().
// Uses AnimationWave as the default animation type.
// Parameters:
//   - version: version string to display
//   - compact: whether to render in compact mode
//   - opts: logo rendering options
//   - durationMs: total animation duration in milliseconds (e.g., 1000 for 1 second)
//   - maxPlays: number of times to play animation (1 = once, 0 = infinite)
func NewAnimatedLogo(version string, compact bool, opts Opts, durationMs int, maxPlays int) *AnimatedLogo {
	return NewAnimatedLogoWithType(version, compact, opts, durationMs, maxPlays, AnimationWave)
}

// NewAnimatedLogoWithType creates a new animated logo with a specific animation type.
// Parameters:
//   - version: version string to display
//   - compact: whether to render in compact mode
//   - opts: logo rendering options
//   - durationMs: total animation duration in milliseconds (e.g., 1000 for 1 second)
//   - maxPlays: number of times to play animation (1 = once, 0 = infinite)
//   - animationType: the type of animation to use
func NewAnimatedLogoWithType(version string, compact bool, opts Opts, durationMs int, maxPlays int, animationType AnimationType) *AnimatedLogo {
	// Render the full logo once
	fullContent := Render(version, compact, opts)

	// Split into lines
	lines := strings.Split(fullContent, "\n")

	// Find max line width and total chars (for animation calculations)
	maxWidth := 0
	totalChars := 0
	lineWidths := make([]int, len(lines))
	for i, line := range lines {
		// Count visible characters (excluding ANSI codes)
		visibleWidth := visibleLength(line)
		lineWidths[i] = visibleWidth
		totalChars += visibleWidth
		if visibleWidth > maxWidth {
			maxWidth = visibleWidth
		}
	}

	// Calculate animation parameters
	// Tick interval is 70ms, so ticks in duration = durationMs / 70
	const tickIntervalMs = 70
	totalTicks := durationMs / tickIntervalMs
	if totalTicks < 1 {
		totalTicks = 1
	}

	// Chars per tick varies by animation type
	var charsPerTick int
	switch animationType {
	case AnimationTypewriter, AnimationRandom:
		// These reveal total chars over the duration
		charsPerTick = totalChars / totalTicks
		if charsPerTick < 1 {
			charsPerTick = 1
		}
	case AnimationLineByLine:
		// Reveal one line per few ticks
		linesPerTick := len(lines) / totalTicks
		if linesPerTick < 1 {
			charsPerTick = 1 // Will be used as lines per tick
		} else {
			charsPerTick = linesPerTick
		}
	default:
		// Wave and CenterOut use width-based calculation
		charsPerTick = maxWidth / totalTicks
		if charsPerTick < 1 {
			charsPerTick = 1
		}
	}

	// Wave offset: how many chars delay between each line starting
	// A smaller value creates a tighter wave, larger creates more stagger
	waveOffset := 1

	// Initialize reveal progress for each line
	revealedCols := make([]int, len(lines))

	// Initialize per-character reveal state for random animation
	revealedChars := make([][]bool, len(lines))
	for i, line := range lines {
		revealedChars[i] = make([]bool, visibleLength(line))
	}

	return &AnimatedLogo{
		fullContent:   fullContent,
		lines:         lines,
		revealedCols:  revealedCols,
		revealedChars: revealedChars,
		charsPerTick:  charsPerTick,
		waveOffset:    waveOffset,
		totalTicks:    totalTicks,
		currentTick:   0,
		complete:      false,
		playCount:     0,
		maxPlays:      maxPlays,
		maxLineWidth:  maxWidth,
		animationType: animationType,
		totalChars:    totalChars,
		revealedCount: 0,
		lineWidths:    lineWidths,
	}
}

// Tick advances the animation by one frame.
// The behavior depends on the animation type.
func (a *AnimatedLogo) Tick() {
	if a.complete {
		return
	}

	a.currentTick++

	switch a.animationType {
	case AnimationWave:
		a.tickWave()
	case AnimationTypewriter:
		a.tickTypewriter()
	case AnimationLineByLine:
		a.tickLineByLine()
	case AnimationCenterOut:
		a.tickCenterOut()
	case AnimationRandom:
		a.tickRandom()
	default:
		a.tickWave()
	}
}

// tickWave implements wave animation: left-to-right with staggered line delays.
func (a *AnimatedLogo) tickWave() {
	allComplete := true
	for i := range a.lines {
		// Wave delay: line i starts after i * waveOffset chars have been revealed on line 0
		lineDelay := i * a.waveOffset
		effectiveTick := a.currentTick - (lineDelay / max(a.charsPerTick, 1))

		if effectiveTick > 0 {
			targetChars := effectiveTick * a.charsPerTick
			lineWidth := a.lineWidths[i]

			if targetChars >= lineWidth {
				a.revealedCols[i] = lineWidth
			} else {
				a.revealedCols[i] = targetChars
				allComplete = false
			}
		} else {
			a.revealedCols[i] = 0
			allComplete = false
		}
	}

	if allComplete {
		a.complete = true
		a.playCount++
	}
}

// tickTypewriter implements typewriter animation: character-by-character sequentially.
func (a *AnimatedLogo) tickTypewriter() {
	// Reveal charsPerTick characters this tick
	toReveal := a.charsPerTick
	charIndex := 0

	for i := range a.lines {
		lineWidth := a.lineWidths[i]
		if a.revealedCols[i] >= lineWidth {
			charIndex += lineWidth
			continue
		}

		// This line still has characters to reveal
		remaining := lineWidth - a.revealedCols[i]
		if toReveal >= remaining {
			a.revealedCols[i] = lineWidth
			toReveal -= remaining
			charIndex += lineWidth
		} else {
			a.revealedCols[i] += toReveal
			return // Still animating
		}
	}

	// All characters revealed
	a.complete = true
	a.playCount++
}

// tickLineByLine implements line-by-line animation: full lines appear one at a time.
func (a *AnimatedLogo) tickLineByLine() {
	// Calculate which line should be fully revealed based on tick progress
	linesPerTick := float64(len(a.lines)) / float64(a.totalTicks)
	targetLine := int(float64(a.currentTick) * linesPerTick)

	allComplete := true
	for i := range a.lines {
		if i <= targetLine {
			a.revealedCols[i] = a.lineWidths[i]
		} else {
			a.revealedCols[i] = 0
			allComplete = false
		}
	}

	if allComplete || targetLine >= len(a.lines)-1 {
		// Ensure all lines are revealed
		for i := range a.lines {
			a.revealedCols[i] = a.lineWidths[i]
		}
		a.complete = true
		a.playCount++
	}
}

// tickCenterOut implements center-out animation: reveals from center of each line outward.
func (a *AnimatedLogo) tickCenterOut() {
	allComplete := true
	for i := range a.lines {
		lineWidth := a.lineWidths[i]
		if lineWidth == 0 {
			continue
		}

		// Calculate how many chars from center should be revealed
		// Each tick reveals charsPerTick more (total spread, so /2 on each side)
		spread := a.currentTick * a.charsPerTick
		if spread >= lineWidth {
			a.revealedCols[i] = lineWidth
		} else {
			a.revealedCols[i] = spread
			allComplete = false
		}
	}

	if allComplete {
		a.complete = true
		a.playCount++
	}
}

// tickRandom implements random animation: characters appear in random order.
func (a *AnimatedLogo) tickRandom() {
	if a.revealedCount >= a.totalChars {
		a.complete = true
		a.playCount++
		return
	}

	// Reveal charsPerTick random unrevealed characters
	toReveal := a.charsPerTick
	attempts := 0
	maxAttempts := a.totalChars * 2 // Prevent infinite loop

	for toReveal > 0 && attempts < maxAttempts {
		attempts++

		// Pick a random line and character
		lineIdx := rand.Intn(len(a.lines))
		lineWidth := a.lineWidths[lineIdx]
		if lineWidth == 0 {
			continue
		}

		charIdx := rand.Intn(lineWidth)
		if !a.revealedChars[lineIdx][charIdx] {
			a.revealedChars[lineIdx][charIdx] = true
			a.revealedCount++
			toReveal--
		}
	}

	if a.revealedCount >= a.totalChars {
		a.complete = true
		a.playCount++
	}
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// View returns the current animation frame.
// When complete, returns the full logo content.
func (a *AnimatedLogo) View() string {
	if a.complete {
		return a.fullContent
	}

	switch a.animationType {
	case AnimationCenterOut:
		return a.viewCenterOut()
	case AnimationRandom:
		return a.viewRandom()
	default:
		return a.viewLeftToRight()
	}
}

// viewLeftToRight renders for Wave, Typewriter, and LineByLine animations.
func (a *AnimatedLogo) viewLeftToRight() string {
	var result strings.Builder
	for i, line := range a.lines {
		if i > 0 {
			result.WriteString("\n")
		}

		revealed := a.revealedCols[i]
		if revealed <= 0 {
			// Line not started yet - output empty space to maintain layout
			result.WriteString(strings.Repeat(" ", a.lineWidths[i]))
			continue
		}

		// Truncate the line to revealed chars (handling ANSI codes)
		truncated := truncateToVisible(line, revealed)
		remaining := a.lineWidths[i] - revealed
		if remaining > 0 {
			// Pad with spaces to maintain layout
			truncated += strings.Repeat(" ", remaining)
		}
		result.WriteString(truncated)
	}

	return result.String()
}

// viewCenterOut renders center-out animation where characters reveal from center.
func (a *AnimatedLogo) viewCenterOut() string {
	var result strings.Builder
	for i, line := range a.lines {
		if i > 0 {
			result.WriteString("\n")
		}

		lineWidth := a.lineWidths[i]
		if lineWidth == 0 {
			continue
		}

		spread := a.revealedCols[i]
		if spread >= lineWidth {
			result.WriteString(line)
			continue
		}

		// Calculate center and what to reveal
		center := lineWidth / 2
		leftStart := center - spread/2
		rightEnd := center + (spread+1)/2

		if leftStart < 0 {
			leftStart = 0
		}
		if rightEnd > lineWidth {
			rightEnd = lineWidth
		}

		// Build the partial line: spaces + revealed center + spaces
		leftPad := strings.Repeat(" ", leftStart)
		revealed := truncateVisibleRange(line, leftStart, rightEnd)
		rightPad := strings.Repeat(" ", lineWidth-rightEnd)

		result.WriteString(leftPad)
		result.WriteString(revealed)
		result.WriteString(rightPad)
	}

	return result.String()
}

// viewRandom renders random animation where individual characters are revealed.
func (a *AnimatedLogo) viewRandom() string {
	var result strings.Builder
	for i, line := range a.lines {
		if i > 0 {
			result.WriteString("\n")
		}

		lineWidth := a.lineWidths[i]
		if lineWidth == 0 {
			continue
		}

		// Build line character by character based on reveal state
		result.WriteString(renderWithMask(line, a.revealedChars[i]))
	}

	return result.String()
}

// truncateVisibleRange extracts visible characters from start to end index.
func truncateVisibleRange(s string, start, end int) string {
	if start >= end {
		return ""
	}

	var result strings.Builder
	visibleCount := 0
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			if visibleCount >= start && visibleCount < end {
				result.WriteRune(r)
			}
			continue
		}
		if inEscape {
			if visibleCount >= start && visibleCount < end {
				result.WriteRune(r)
			}
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		if visibleCount >= start && visibleCount < end {
			result.WriteRune(r)
		}
		visibleCount++

		if visibleCount >= end {
			break
		}
	}

	// Reset ANSI at the end
	if result.Len() > 0 {
		result.WriteString("\x1b[0m")
	}

	return result.String()
}

// renderWithMask renders a line showing only characters where mask[i] is true.
func renderWithMask(s string, mask []bool) string {
	var result strings.Builder
	visibleIdx := 0
	inEscape := false
	var pendingEscape strings.Builder

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			pendingEscape.Reset()
			pendingEscape.WriteRune(r)
			continue
		}
		if inEscape {
			pendingEscape.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
				// Only write escape if next visible char is revealed
				if visibleIdx < len(mask) && mask[visibleIdx] {
					result.WriteString(pendingEscape.String())
				}
			}
			continue
		}

		if visibleIdx < len(mask) && mask[visibleIdx] {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
		visibleIdx++
	}

	// Reset ANSI at the end
	result.WriteString("\x1b[0m")

	return result.String()
}

// IsComplete returns whether the animation has finished.
func (a *AnimatedLogo) IsComplete() bool {
	return a.complete
}

// Reset resets the animation state for potential replay.
func (a *AnimatedLogo) Reset() {
	a.currentTick = 0
	a.complete = false
	a.revealedCount = 0
	for i := range a.revealedCols {
		a.revealedCols[i] = 0
	}
	// Reset random reveal state
	for i := range a.revealedChars {
		for j := range a.revealedChars[i] {
			a.revealedChars[i][j] = false
		}
	}
}

// GetAnimationType returns the current animation type.
func (a *AnimatedLogo) GetAnimationType() AnimationType {
	return a.animationType
}

// visibleLength returns the number of visible characters in a string,
// excluding ANSI escape codes.
func visibleLength(s string) int {
	length := 0
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		length++
	}

	return length
}

// truncateToVisible truncates a string to n visible characters,
// preserving ANSI escape codes.
func truncateToVisible(s string, n int) string {
	if n <= 0 {
		return ""
	}

	var result strings.Builder
	visibleCount := 0
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		if visibleCount >= n {
			break
		}
		result.WriteRune(r)
		visibleCount++
	}

	// Reset ANSI at the end to prevent color bleeding
	if visibleCount > 0 {
		result.WriteString("\x1b[0m")
	}

	return result.String()
}
