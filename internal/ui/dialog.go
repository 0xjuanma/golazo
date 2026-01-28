// Package ui provides terminal user interface components for golazo.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Dialog sizing constants (30% larger for better readability).
const (
	DefaultDialogMaxWidth  = 104
	DefaultDialogMaxHeight = 39
)

// DialogAction represents an action returned by a dialog after handling a message.
type DialogAction any

// DialogActionClose signals that the dialog should be closed.
type DialogActionClose struct{}

// Dialog is a component that can be displayed as an overlay on top of the UI.
type Dialog interface {
	// ID returns the unique identifier of the dialog.
	ID() string
	// Update processes a message and returns the updated dialog and an optional action.
	Update(msg tea.Msg) (Dialog, DialogAction)
	// View renders the dialog content within the specified dimensions.
	View(width, height int) string
}

// DialogOverlay manages multiple dialogs as an overlay stack.
type DialogOverlay struct {
	dialogs []Dialog
}

// NewDialogOverlay creates a new DialogOverlay instance.
func NewDialogOverlay() *DialogOverlay {
	return &DialogOverlay{
		dialogs: []Dialog{},
	}
}

// HasDialogs checks if there are any active dialogs.
func (o *DialogOverlay) HasDialogs() bool {
	return len(o.dialogs) > 0
}

// ContainsDialog checks if a dialog with the specified ID exists.
func (o *DialogOverlay) ContainsDialog(dialogID string) bool {
	for _, dialog := range o.dialogs {
		if dialog.ID() == dialogID {
			return true
		}
	}
	return false
}

// OpenDialog adds a new dialog to the stack.
func (o *DialogOverlay) OpenDialog(dialog Dialog) {
	o.dialogs = append(o.dialogs, dialog)
}

// CloseDialog removes the dialog with the specified ID from the stack.
func (o *DialogOverlay) CloseDialog(dialogID string) {
	for i, dialog := range o.dialogs {
		if dialog.ID() == dialogID {
			o.dialogs = append(o.dialogs[:i], o.dialogs[i+1:]...)
			return
		}
	}
}

// CloseFrontDialog closes the front (topmost) dialog in the stack.
func (o *DialogOverlay) CloseFrontDialog() {
	if len(o.dialogs) == 0 {
		return
	}
	o.dialogs = o.dialogs[:len(o.dialogs)-1]
}

// FrontDialog returns the front (topmost) dialog, or nil if there are no dialogs.
func (o *DialogOverlay) FrontDialog() Dialog {
	if len(o.dialogs) == 0 {
		return nil
	}
	return o.dialogs[len(o.dialogs)-1]
}

// Update handles message routing to the front dialog.
// Returns the action from the dialog, if any.
func (o *DialogOverlay) Update(msg tea.Msg) DialogAction {
	if len(o.dialogs) == 0 {
		return nil
	}

	idx := len(o.dialogs) - 1
	dialog := o.dialogs[idx]

	updatedDialog, action := dialog.Update(msg)
	o.dialogs[idx] = updatedDialog

	return action
}

// View renders the overlay with the front dialog centered.
func (o *DialogOverlay) View(width, height int) string {
	if len(o.dialogs) == 0 {
		return ""
	}

	dialog := o.dialogs[len(o.dialogs)-1]
	dialogView := dialog.View(width, height)

	return centerDialog(dialogView, width, height)
}

// centerDialog centers a dialog view within the given dimensions.
func centerDialog(view string, width, height int) string {
	viewWidth, viewHeight := lipgloss.Size(view)

	// Calculate padding to center the dialog
	padLeft := (width - viewWidth) / 2
	padTop := (height - viewHeight) / 2

	if padLeft < 0 {
		padLeft = 0
	}
	if padTop < 0 {
		padTop = 0
	}

	return lipgloss.NewStyle().
		PaddingLeft(padLeft).
		PaddingTop(padTop).
		Width(width).
		Height(height).
		Render(view)
}

// DialogSize calculates appropriate dialog dimensions based on content and screen size.
func DialogSize(screenWidth, screenHeight, contentWidth, contentHeight int) (width, height int) {
	// Use 80% of screen or content size, whichever is smaller
	maxWidth := screenWidth * 80 / 100
	maxHeight := screenHeight * 80 / 100

	width = min(contentWidth, maxWidth)
	height = min(contentHeight, maxHeight)

	// Apply absolute maximums
	if width > DefaultDialogMaxWidth {
		width = DefaultDialogMaxWidth
	}
	if height > DefaultDialogMaxHeight {
		height = DefaultDialogMaxHeight
	}

	return width, height
}
