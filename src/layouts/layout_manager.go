package layouts

import (
	"atcli/src/services"
	"atcli/src/types"

	"github.com/rivo/tview"
)

type LayoutManager struct {
	pages         *tview.Pages
	layouts       types.LayoutMap
	currentLayout string
	eventBus      *services.EventBus
}

func NewLayoutManager(app *tview.Application, eventBus *services.EventBus) *LayoutManager {
	pages := tview.NewPages()

	lm := &LayoutManager{
		pages:         pages,
		layouts:       make(types.LayoutMap),
		currentLayout: "",
		eventBus:      eventBus,
	}

	// Subscribe to screen change events
	eventBus.Subscribe(types.EventChangeLayout, lm.handleScreenChanged)
	// Subscribe to layout change events
	eventBus.Subscribe(types.EventLayoutChange, lm.handleLayoutChange)

	// Set the pages component as the root
	app.SetRoot(pages, true)

	return lm
}

// Register adds a new layout to the manager with the given name
func (l *LayoutManager) Register(layout types.LayoutInterface, isVisible bool) {
	name := layout.GetName()

	l.layouts[name] = layout

	// Add the page but don't show it yet (visible=true, primitive is loaded but not shown)
	l.pages.AddPage(name, layout.GetComponent(), true, isVisible)

	if isVisible {
		l.currentLayout = name
	}
}

// handleScreenChanged handles screen change events
func (l *LayoutManager) handleScreenChanged(event types.Event) {
	if layoutName, ok := event.Payload.(string); ok {
		if layout, exists := l.layouts[layoutName]; exists {
			// Switch to the page
			l.pages.SwitchToPage(layoutName)
			l.currentLayout = layoutName

			// Call OnLayoutChange on the layout
			layout.OnLayoutChange()
		}
	}
}

// handleLayoutChange handles layout change events
func (l *LayoutManager) handleLayoutChange(event types.Event) {
	// Call OnLayoutChange on the current layout
	if layout, exists := l.layouts[l.currentLayout]; exists {
		layout.OnLayoutChange()
	}
}

// GetCurrentLayout returns the name of the currently active layout
func (l *LayoutManager) GetCurrentLayout() string {
	return l.currentLayout
}
