package views

import "atcli/src/types"

type ViewManager struct {
	viewRegistry types.ViewMap
}

func NewViewManager() *ViewManager {
	return &ViewManager{viewRegistry: types.ViewMap{}}
}

func (v *ViewManager) Register(view types.ViewInterface) {
	v.viewRegistry[view.GetName()] = view
}

func (v *ViewManager) GetView(name string) types.ViewInterface {
	if view, exists := v.viewRegistry[name]; exists {
		return view
	}

	panic("View not found: " + name)
}
