package layouts

import (
	"atcli/src/cmd"
	"atcli/src/services"
	"atcli/src/types"
	"fmt"
	"strings"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const version = "0.1"

// HelpLayout represents a help screen
type HelpLayout struct {
	content  tview.Primitive
	eventBus *services.EventBus
}

// NewHelpLayout creates a new help layout
func NewHelpLayout(eventBus *services.EventBus, commandManager *cmd.CommandManager) *HelpLayout {
	// Create the help message
	helpMsg := fmt.Sprintf("atcli - version %s", version)
	helpMsg += "A modern AT command terminal.\n\n"
	helpMsg += "Available commands:\n"

	// Add commands to the help message if we have a command manager
	commands := commandManager.ListCommands()
	for _, cmd := range commands {
		helpMsg += fmt.Sprintf("/%s - %s\n", cmd.Name, cmd.Description)
	}

	// Function to return to home screen
	returnToHome := func() {
		eventBus.Publish(types.Event{
			Type:    types.EventChangeLayout,
			Payload: "home",
		})
	}

	lines := strings.Count(helpMsg, "\n") + 1
	buttons := 1
	padding := 2
	modalHeight := lines + buttons + padding + 10

	modal := tview.NewModal().
		SetText(helpMsg).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			returnToHome()
		}).
		SetTitle(" Help ")
	services.LogMessage("Showing popup")
	services.LogMessage(helpMsg)

	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			returnToHome()
			return nil // Consume the event
		}
		return event // Pass other keys through
	})

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(modal, modalHeight, 1, true).
				AddItem(nil, 0, 1, false),
			0, 1, true).
		AddItem(nil, 0, 1, false)

	return &HelpLayout{
		content:  flex,
		eventBus: eventBus,
	}
}

func (h *HelpLayout) GetName() string {
	return "help"
}

// GetComponent returns the help screen content
func (h *HelpLayout) GetComponent() tview.Primitive {
	return h.content
}

// OnLayoutChange is called when the layout changes or becomes active
func (h *HelpLayout) OnLayoutChange() {
	// Help layout doesn't need to do anything special when layout changes
}

var _ types.LayoutInterface = (*HelpLayout)(nil)
