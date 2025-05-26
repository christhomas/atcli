package layouts

import (
	"atcli/src/services"
	"atcli/src/types"
	"atcli/src/views"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type GPSLayout struct {
	layout      tview.Primitive
	leftPanel   *tview.Flex
	commandView tview.Primitive
	logView     tview.Primitive
	logViewRef  *views.LogView // Reference to the actual LogView for accessing its methods
	eventBus    *services.EventBus
}

func NewGPSLayout(viewManager *views.ViewManager, eventBus *services.EventBus) *GPSLayout {
	// Get references to the views we'll need
	commandView := viewManager.GetView("command").GetComponent()
	gpsView := viewManager.GetView("gps").GetComponent()
	logViewRef := viewManager.GetView("log").(*views.LogView) // Get the actual LogView reference
	logView := logViewRef.GetComponent()                      // Get the Primitive component

	// Left panel: vertical flex for commandView and logView
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add both views to the left panel, but set logView's proportion to 0 initially to hide it
	leftPanel.AddItem(commandView, 0, 1, false)
	leftPanel.AddItem(logView, 0, 0, false) // Initially hidden with proportion 0

	// Right panel: GPS view
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(gpsView, 0, 1, false)

	// Horizontal split for the two panels
	panelsFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	panelsFlex.AddItem(leftPanel, 0, 1, false)
	panelsFlex.AddItem(rightPanel, 0, 1, false)
	panelsFlex.SetBackgroundColor(tcell.ColorBlack)

	// GPS screen: input, panels, status
	gpsScreen := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(viewManager.GetView("input").GetComponent(), 1, 0, true).
		AddItem(panelsFlex, 0, 1, false).
		AddItem(viewManager.GetView("statusbar").GetComponent(), 1, 0, false)
	gpsScreen.SetBackgroundColor(tcell.ColorBlack)

	// Create the GPS layout
	gpsLayout := &GPSLayout{
		layout:      gpsScreen,
		leftPanel:   leftPanel,
		commandView: commandView,
		logView:     logView,
		logViewRef:  logViewRef,
		eventBus:    eventBus,
	}

	return gpsLayout
}

func (g *GPSLayout) GetName() string {
	return "gps"
}

func (g *GPSLayout) GetComponent() tview.Primitive {
	return g.layout
}

// OnLayoutChange is called when the layout changes or becomes active
func (g *GPSLayout) OnLayoutChange() {
	// Check if the log view should be visible based on its state
	if g.logViewRef.IsVisible() {
		// Show the log view without changing its visibility state
		g.leftPanel.ResizeItem(g.commandView, 0, 2)
		g.leftPanel.ResizeItem(g.logView, 0, 1)
	} else {
		// Hide the log view without changing its visibility state
		g.leftPanel.ResizeItem(g.commandView, 0, 1)
		g.leftPanel.ResizeItem(g.logView, 0, 0)
	}
}

var _ types.LayoutInterface = (*GPSLayout)(nil)
