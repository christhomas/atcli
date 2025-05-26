package layouts

import (
	"atcli/src/services"
	"atcli/src/types"
	"atcli/src/views"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SignalChartLayout struct {
	layout      *tview.Flex
	leftPanel   *tview.Flex
	commandView tview.Primitive
	logView     tview.Primitive
	logViewRef  *views.LogView // Reference to the actual LogView for accessing its methods
	eventBus    *services.EventBus
}

func NewSignalChartLayout(viewManager *views.ViewManager, eventBus *services.EventBus) *SignalChartLayout {
	// Get references to the views we'll need
	commandView := viewManager.GetView("command").GetComponent()
	signalView := viewManager.GetView("signal").GetComponent()
	logViewRef := viewManager.GetView("log").(*views.LogView) // Get the actual LogView reference
	logView := logViewRef.GetComponent()                      // Get the Primitive component

	// Left panel: vertical flex for commandView and logView
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add both views to the left panel, but set logView's proportion to 0 initially to hide it
	leftPanel.AddItem(commandView, 0, 1, false)
	leftPanel.AddItem(logView, 0, 0, false) // Initially hidden with proportion 0

	// Right panel: signal chart with frame
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(signalView, 0, 1, false)

	// Horizontal split for the two panels
	panelsFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	panelsFlex.AddItem(leftPanel, 0, 1, false)
	panelsFlex.AddItem(rightPanel, 0, 1, false)
	panelsFlex.SetBackgroundColor(tcell.ColorBlack)

	// Signal screen: input, panels, status
	signalScreen := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(viewManager.GetView("input").GetComponent(), 1, 0, true).
		AddItem(panelsFlex, 0, 1, false).
		AddItem(viewManager.GetView("statusbar").GetComponent(), 1, 0, false)
	signalScreen.SetBackgroundColor(tcell.ColorBlack)

	// Create the signal layout
	signalLayout := &SignalChartLayout{
		layout:      signalScreen,
		leftPanel:   leftPanel,
		commandView: commandView,
		logView:     logView,
		logViewRef:  logViewRef,
		eventBus:    eventBus,
	}

	return signalLayout
}

func (s *SignalChartLayout) GetName() string {
	return "signal"
}

func (s *SignalChartLayout) GetComponent() tview.Primitive {
	return s.layout
}

// OnLayoutChange is called when the layout changes or becomes active
func (s *SignalChartLayout) OnLayoutChange() {
	// Check if the log view should be visible based on its state
	if s.logViewRef.IsVisible() {
		// Show the log view without changing its visibility state
		s.leftPanel.ResizeItem(s.commandView, 0, 2)
		s.leftPanel.ResizeItem(s.logView, 0, 1)
	} else {
		// Hide the log view without changing its visibility state
		s.leftPanel.ResizeItem(s.commandView, 0, 1)
		s.leftPanel.ResizeItem(s.logView, 0, 0)
	}
}

var _ types.LayoutInterface = (*SignalChartLayout)(nil)
