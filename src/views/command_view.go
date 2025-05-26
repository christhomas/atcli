package views

import (
	"atcli/src/services"
	"atcli/src/types"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CommandView struct {
	eventBus         *services.EventBus
	commandView      *tview.TextView
	commandHistory   []types.HistoryItem
	historyIndex     int
	currentHighlight int
}

func NewCommandView(eventBus *services.EventBus, app *tview.Application, title string, color tcell.Color) *CommandView {
	commandView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)

	self := &CommandView{
		eventBus:         eventBus,
		commandView:      commandView,
		commandHistory:   []types.HistoryItem{},
		historyIndex:     -1,
		currentHighlight: -1,
	}

	commandView.
		SetTitle(title).
		SetBorder(true).
		SetBackgroundColor(color)

	commandView.SetScrollable(true)
	commandView.SetInputCapture(self.SetInputCapture)
	commandView.SetChangedFunc(self.SetChanged)

	eventBus.Subscribe(types.EventCommandSent, self.handleCommandSent)
	eventBus.Subscribe(types.EventCommandHistory, self.handleCommandHistory)

	return self
}

func (c *CommandView) GetName() string {
	return "command"
}

func (c *CommandView) GetComponent() tview.Primitive {
	return c.commandView
}

func (c *CommandView) SetChanged() {
	c.eventBus.Publish(types.Event{Type: types.EventAppRedraw})
}

// Set up key handlers for the panels to allow scrolling but redirect typing to input
func (c *CommandView) SetInputCapture(event *tcell.EventKey) *tcell.EventKey {
	// Allow navigation keys for scrolling
	switch event.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyHome, tcell.KeyEnd:
		return event
	case tcell.KeyRune:
		// Check if it's a mouse event (mouse events come through as KeyRune with special runes)
		if event.Rune() == 0 {
			// This is likely a mouse event, allow it for text selection
			return event
		}
		// For other rune keys, redirect to input field
		c.eventBus.Publish(types.Event{Type: types.EventFocusInput})
		return event
	default:
		// For any other key, redirect to input field
		c.eventBus.Publish(types.Event{Type: types.EventFocusInput})
		return event
	}
}

func (c *CommandView) handleCommandSent(event types.Event) {
	// Get current line count for history indexing
	lineCount := strings.Count(c.commandView.GetText(true), "\n")

	// Add to command history with its line position
	c.commandHistory = append(c.commandHistory, types.HistoryItem{Cmd: event.Payload.(string), Index: lineCount})
	c.historyIndex = -1 // Reset history index

	// Clear any existing highlight
	if c.currentHighlight >= 0 {
		c.commandView.Highlight("")
		c.currentHighlight = -1
	}

	c.commandView.Write([]byte(event.Payload.(string) + "\n"))
	c.commandView.ScrollToEnd()
}

func (c *CommandView) handleCommandHistory(event types.Event) {
	direction, ok := event.Payload.(int)
	if !ok {
		return
	}

	// Going up in history (-1)
	if direction < 0 {
		if len(c.commandHistory) == 0 {
			return
		}

		if c.historyIndex < len(c.commandHistory)-1 {
			c.historyIndex++
		}

		// Clear previous highlight
		if c.currentHighlight >= 0 {
			c.commandView.Highlight("")
		}

		// Get the history item
		item := c.commandHistory[len(c.commandHistory)-1-c.historyIndex]

		// Publish event to set input field text
		c.eventBus.Publish(types.Event{
			Type:    types.EventInputSetCommand,
			Payload: item.Cmd,
		})

		// Highlight the command in the left panel
		c.commandView.Highlight(fmt.Sprintf("%d", item.Index))
		c.currentHighlight = item.Index
	} else if direction > 0 { // Going down in history (+1)
		if c.historyIndex > 0 {
			// Clear previous highlight
			if c.currentHighlight >= 0 {
				c.commandView.Highlight("")
			}

			c.historyIndex--

			// Get the history item
			item := c.commandHistory[len(c.commandHistory)-1-c.historyIndex]

			// Publish event to set input field text
			c.eventBus.Publish(types.Event{
				Type:    types.EventInputSetCommand,
				Payload: item.Cmd,
			})

			// Highlight the command in the left panel
			c.commandView.Highlight(fmt.Sprintf("%d", item.Index))
			c.currentHighlight = item.Index
		} else if c.historyIndex == 0 {
			// Clear highlight when returning to empty input
			if c.currentHighlight >= 0 {
				c.commandView.Highlight("")
				c.currentHighlight = -1
			}
			c.historyIndex = -1

			// Publish event to clear input field text
			c.eventBus.Publish(types.Event{
				Type:    types.EventInputSetCommand,
				Payload: "",
			})
		}
	}
}

var _ types.ViewInterface = (*CommandView)(nil)
