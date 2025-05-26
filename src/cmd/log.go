package cmd

import (
	"atcli/src/services"
	"atcli/src/types"
	"atcli/src/views"
	"fmt"
	"sync"
	"time"
)

// LogCommand implements CommandInterface for /log
// It manages the UI panel for error logging
type LogCommand struct {
	eventBus    *services.EventBus
	name        string
	description string
	active      bool
	mutex       sync.Mutex
	logView     *views.LogView
}

// NewLogCommand creates a new log command
func NewLogCommand(eventBus *services.EventBus, logView *views.LogView) *LogCommand {
	return &LogCommand{
		eventBus:    eventBus,
		name:        "log",
		description: "Show error logging panel. Usage: /log, /log off, or /log close",
		active:      false,
		logView:     logView,
	}
}

// GetName returns the command name
func (l *LogCommand) GetName() string {
	return l.name
}

// GetDescription returns the command description
func (l *LogCommand) GetDescription() string {
	return l.description
}

// Run executes the log command
func (l *LogCommand) Run(args []string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Handle off or close parameters to deactivate logging
	if len(args) > 0 && (args[0] == "off" || args[0] == "close") {
		// Always set to inactive regardless of current state
		l.active = false
	} else {
		// Toggle log view state
		l.active = !l.active
	}

	l.logView.SetVisible(l.active)

	// Notify layouts that they need to update their UI
	l.eventBus.Publish(types.Event{
		Type: types.EventLayoutChange,
	})

	// Add a timestamp to show when logging was activated/deactivated
	timestamp := time.Now().Format("15:04:05")
	if l.active {
		services.LogMessage(fmt.Sprintf("[yellow]Logging activated at %s[white]", timestamp))
	} else {
		services.LogMessage(fmt.Sprintf("[yellow]Logging deactivated at %s[white]", timestamp))
	}

	return nil
}

var _ types.CommandInterface = (*LogCommand)(nil)
