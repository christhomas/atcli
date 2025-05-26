package layouts

import (
	"atcli/src/services"
	"atcli/src/types"
	"atcli/src/views"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type HomeLayout struct {
	layout      *tview.Flex
	leftPanel   *tview.Flex
	commandView *views.CommandView
	logView     *views.LogView
	eventBus    *services.EventBus
}

func NewHomeLayout(viewManager *views.ViewManager, eventBus *services.EventBus) *HomeLayout {
	// Get references to the views we'll need
	commandView := viewManager.GetView("command").(*views.CommandView)
	replyView := viewManager.GetView("reply").(*views.ReplyView)
	logView := viewManager.GetView("log").(*views.LogView)

	// Left panel: vertical flex for commandView and logView
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add both views to the left panel, but set logView's proportion to 0 initially to hide it
	leftPanel.AddItem(commandView.GetComponent(), 0, 1, false)
	leftPanel.AddItem(logView.GetComponent(), 0, 0, false) // Initially hidden with proportion 0

	// Right panel: vertical flex for replies
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(replyView.GetComponent(), 0, 1, false)

	// Horizontal split for the two panels
	panelsFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	panelsFlex.AddItem(leftPanel, 0, 1, false)
	panelsFlex.AddItem(rightPanel, 0, 1, false)
	panelsFlex.SetBackgroundColor(tcell.ColorBlack)

	// Home screen: input, panels, status
	homeScreen := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(viewManager.GetView("input").GetComponent(), 1, 0, true).
		AddItem(panelsFlex, 0, 1, false).
		AddItem(viewManager.GetView("statusbar").GetComponent(), 1, 0, false)
	homeScreen.SetBackgroundColor(tcell.ColorBlack)

	// Create the home layout
	homeLayout := &HomeLayout{
		layout:      homeScreen,
		leftPanel:   leftPanel,
		commandView: commandView,
		logView:     logView,
		eventBus:    eventBus,
	}

	return homeLayout
}

func (h *HomeLayout) GetName() string {
	return "home"
}

func (h *HomeLayout) GetComponent() tview.Primitive {
	return h.layout
}

// OnLayoutChange is called when the layout changes or becomes active
func (h *HomeLayout) OnLayoutChange() {
	// Check if the log view should be visible based on its state
	if h.logView.IsVisible() {
		// Show the log view without changing its visibility state
		h.leftPanel.ResizeItem(h.commandView.GetComponent(), 0, 2)
		h.leftPanel.ResizeItem(h.logView.GetComponent(), 0, 1)
	} else {
		// Hide the log view without changing its visibility state
		h.leftPanel.ResizeItem(h.commandView.GetComponent(), 0, 1)
		h.leftPanel.ResizeItem(h.logView.GetComponent(), 0, 0)
	}
}

var _ types.LayoutInterface = (*HomeLayout)(nil)
