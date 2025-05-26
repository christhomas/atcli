package cmd

import (
	"atcli/src/services"
	"atcli/src/types"
)

// SignalCommand implements CommandInterface for /signal command
// It switches to the signal layout
type SignalCommand struct {
	eventBus    *services.EventBus
	name        string
	description string
}

// NewSignalCommand creates a new signal command
func NewSignalCommand(eventBus *services.EventBus) *SignalCommand {
	return &SignalCommand{
		eventBus:    eventBus,
		name:        "signal",
		description: "Show signal strength chart",
	}
}

// GetName returns the name of the command
func (s *SignalCommand) GetName() string {
	return s.name
}

// GetDescription returns the description of the command
func (s *SignalCommand) GetDescription() string {
	return s.description
}

// Run executes the signal command
func (s *SignalCommand) Run(args []string) error {
	services.LogMessage("[blue]Signal command started[white]")

	// Check if we have arguments
	if len(args) > 0 {
		// Handle the 'close' argument to return to home page
		if args[0] == "close" {
			services.LogMessage("[blue]Closing signal view and returning to home[white]")
			// First send the stop signal event to stop the monitoring
			s.eventBus.Publish(types.Event{
				Type: types.EventStopSignal,
			})
			// Then change the layout back to home
			s.eventBus.Publish(types.Event{
				Type:    types.EventChangeLayout,
				Payload: "home",
			})
			return nil
		}
	}

	// Switch to signal screen using the event bus
	s.eventBus.Publish(types.Event{
		Type:    types.EventChangeLayout,
		Payload: "signal",
	})
	
	// Start the signal monitoring
	s.eventBus.Publish(types.Event{
		Type: types.EventStartSignal,
	})

	return nil
}

// Ensure SignalCommand implements CommandInterface
var _ types.CommandInterface = (*SignalCommand)(nil)
