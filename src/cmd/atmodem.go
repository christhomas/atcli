package cmd

import (
	"atcli/src/services"
	"atcli/src/types"
)

type ATModemCommand struct {
	eventBus    *services.EventBus
	name        string
	description string
}

func NewATModemCommand(eventBus *services.EventBus) *ATModemCommand {
	return &ATModemCommand{
		eventBus:    eventBus,
		name:        "atmodem",
		description: "Send AT command to modem",
	}
}

func (a *ATModemCommand) GetName() string {
	return a.name
}

func (a *ATModemCommand) GetDescription() string {
	return a.description
}

func (a *ATModemCommand) Run(args []string) error {
	// Join all arguments into a single command string
	command := ""
	if len(args) > 0 {
		for _, arg := range args {
			if command != "" {
				command += " "
			}
			command += arg
		}
	}

	// Publish the command to the serial port
	a.eventBus.Publish(types.Event{
		Type:    types.EventATModemCommand,
		Payload: command,
	})

	return nil
}

var _ types.CommandInterface = (*ATModemCommand)(nil)
