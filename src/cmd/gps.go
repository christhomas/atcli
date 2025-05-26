package cmd

import (
	"atcli/src/services"
	"atcli/src/types"
)

// GPSCommand implements CommandInterface for /gps command
// It switches to the GPS layout
type GPSCommand struct {
	eventBus    *services.EventBus
	name        string
	description string
}

// NewGPSCommand creates a new GPS command
func NewGPSCommand(eventBus *services.EventBus) *GPSCommand {
	return &GPSCommand{
		eventBus:    eventBus,
		name:        "gps",
		description: "Show GPS location information",
	}
}

// GetName returns the name of the command
func (g *GPSCommand) GetName() string {
	return g.name
}

// GetDescription returns the description of the command
func (g *GPSCommand) GetDescription() string {
	return g.description
}

// Run executes the GPS command
func (g *GPSCommand) Run(args []string) error {
	// Check if we have arguments
	if len(args) > 0 {
		// Handle the 'close' argument to return to home page
		if args[0] == "close" {
			// First send the stop GPS event to stop the monitoring
			g.eventBus.Publish(types.Event{
				Type: types.EventStopGPS,
			})
			// Then change the layout back to home
			g.eventBus.Publish(types.Event{
				Type:    types.EventChangeLayout,
				Payload: "home",
			})
			return nil
		}
	}

	// Switch to GPS screen using the event bus
	g.eventBus.Publish(types.Event{
		Type:    types.EventChangeLayout,
		Payload: "gps",
	})
	
	// Start the GPS monitoring
	g.eventBus.Publish(types.Event{
		Type: types.EventStartGPS,
	})

	return nil
}

// Ensure GPSCommand implements CommandInterface
var _ types.CommandInterface = (*GPSCommand)(nil)
