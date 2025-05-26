package views

import (
	"atcli/src/services"
	"atcli/src/types"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type InputField struct {
	eventBus   *services.EventBus
	inputField *tview.InputField
}

func NewInputField(eventBus *services.EventBus, label string, color tcell.Color) *InputField {
	inputField := tview.NewInputField().SetLabel(label).SetFieldWidth(0)

	self := &InputField{
		eventBus:   eventBus,
		inputField: inputField,
	}

	inputField.SetBackgroundColor(color)
	inputField.SetDoneFunc(self.SetDoneFunc)
	inputField.SetInputCapture(self.SetInputCapture)

	eventBus.Subscribe(types.EventFocusInput, self.handleFocusInput)
	eventBus.Subscribe(types.EventInputSetCommand, self.handleSetCommand)

	return self
}

func (i *InputField) GetName() string {
	return "input"
}

func (i *InputField) GetComponent() tview.Primitive {
	return i.inputField
}

// Input handler: send command to serial and echo in left panel
func (i *InputField) SetDoneFunc(key tcell.Key) {
	if key != tcell.KeyEnter {
		return
	}

	userInput := i.inputField.GetText()
	i.inputField.SetText("")

	if userInput == "" {
		return
	}

	i.eventBus.Publish(types.Event{
		Type:    types.EventCommandSent,
		Payload: userInput,
	})
}

// Handle up/down keys for command history
func (i *InputField) SetInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyUp:
		i.eventBus.Publish(types.Event{Type: types.EventCommandHistory, Payload: -1})
	case tcell.KeyDown:
		i.eventBus.Publish(types.Event{Type: types.EventCommandHistory, Payload: 1})
	}
	return event
}

func (i *InputField) handleFocusInput(event types.Event) {
	i.eventBus.Publish(types.Event{
		Type:    types.EventAppFocus,
		Payload: i.inputField,
	})
}

// handleSetCommand sets the text of the input field based on a command history selection
func (i *InputField) handleSetCommand(event types.Event) {
	if text, ok := event.Payload.(string); ok {
		// First clear the field
		i.inputField.SetText("")
		// Then set the text - this should position the cursor at the end
		i.inputField.SetText(text)
	}
}

var _ types.ViewInterface = (*InputField)(nil)
