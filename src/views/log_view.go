package views

import (
	"atcli/src/services"
	"atcli/src/types"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LogView struct {
	view     *tview.TextView
	eventBus *services.EventBus
	app      *tview.Application
	visible  bool
}

func NewLogView(app *tview.Application, eventBus *services.EventBus) *LogView {
	// Create log view for error logging
	logView := tview.NewTextView().SetDynamicColors(true).SetChangedFunc(func() { app.Draw() })

	logView.SetTitle("Log Messages").SetBorder(true)
	logView.SetBackgroundColor(tcell.ColorBlack)
	logView.SetScrollable(true)
	logView.SetWordWrap(true)

	logView.SetText("Log view initialized. Use /log to toggle this view.\n")

	// Create the log view instance
	view := &LogView{
		view:     logView,
		eventBus: eventBus,
		app:      app,
	}

	// Subscribe to log messages
	eventBus.Subscribe(types.EventLogMessage, view.handleLogMessage)

	return view
}

// handleLogMessage processes log messages from the event bus
func (l *LogView) handleLogMessage(event types.Event) {
	if message, ok := event.Payload.(string); ok {
		// Log to standard output for debugging
		// log.Println("Log message received:", message)

		// Directly write to the text view without QueueUpdateDraw
		// This avoids potential deadlocks in the UI thread
		current := l.view.GetText(false) // Get current text without tags
		newText := current + message + "\n"
		l.view.SetText(newText)
		l.view.ScrollToEnd()
	}
}

func (l *LogView) GetName() string {
	return "log"
}

func (l *LogView) GetComponent() tview.Primitive {
	return l.view
}

// IsVisible returns whether the log view is currently visible
func (l *LogView) IsVisible() bool {
	return l.visible
}

// SetVisible sets the visibility state of the log view
func (l *LogView) SetVisible(visible bool) {
	l.visible = visible
}

var _ types.ViewInterface = (*LogView)(nil)
