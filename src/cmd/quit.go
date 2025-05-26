package cmd

import (
	"atcli/src/services"
	"atcli/src/types"
)

// QuitCommand handles the /quit command to safely exit the application
type QuitCommand struct {
	eventBus    *services.EventBus
	name        string
	description string
}

// NewQuitCommand creates a new quit command
func NewQuitCommand(eventBus *services.EventBus) *QuitCommand {
	return &QuitCommand{
		eventBus:    eventBus,
		name:        "quit",
		description: "Safely exit the application",
	}
}

// GetName returns the name of the command
func (q *QuitCommand) GetName() string {
	return q.name
}

// GetDescription returns the description of the command
func (q *QuitCommand) GetDescription() string {
	return q.description
}

// Run executes the quit command
func (q *QuitCommand) Run(args []string) error {
	// Publish the shutdown event
	q.eventBus.Publish(types.Event{
		Type: types.EventAppShutdown,
	})

	return nil
}

// Ensure QuitCommand implements CommandInterface
var _ types.CommandInterface = (*QuitCommand)(nil)
