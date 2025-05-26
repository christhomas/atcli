package cmd

import (
	"atcli/src/services"
	"atcli/src/types"

	"github.com/rivo/tview"
)

type HelpCommand struct {
	name        string
	description string
	cmdManager  *CommandManager
	eventBus    *services.EventBus
	app         *tview.Application
}

func NewHelpCommand(cmdManager *CommandManager, eventBus *services.EventBus, app *tview.Application) *HelpCommand {
	return &HelpCommand{
		name:        "help",
		description: "Show help and version information",
		cmdManager:  cmdManager,
		eventBus:    eventBus,
		app:         app,
	}
}

func (h *HelpCommand) GetName() string {
	return h.name
}

func (h *HelpCommand) GetDescription() string {
	return h.description
}

func (h *HelpCommand) Run(args []string) error {
	// Simply publish an event to switch to the help layout
	h.eventBus.Publish(types.Event{
		Type:    types.EventChangeLayout,
		Payload: "help",
	})

	return nil
}

// Ensure HelpCommand implements CommandInterface
var _ types.CommandInterface = (*HelpCommand)(nil)
