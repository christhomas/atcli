package views

import (
	"atcli/src/types"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type StatusBar struct {
	statusBar *tview.TextView
	portName  string
	baudRate  int
}

func NewStatusBar() *StatusBar {
	statusBar := tview.NewTextView().SetDynamicColors(true)
	statusBar.SetBackgroundColor(tcell.ColorBlack)
	statusBar.SetTextColor(tcell.ColorWhite)

	return &StatusBar{statusBar: statusBar}
}

func (s *StatusBar) SetPortName(portName string) {
	s.portName = portName
	s.setStatus()
}

func (s *StatusBar) SetBaudRate(baudRate int) {
	s.baudRate = baudRate
	s.setStatus()
}

func (s *StatusBar) GetName() string {
	return "statusbar"
}

func (s *StatusBar) GetComponent() tview.Primitive {
	return s.statusBar
}

func (s *StatusBar) setStatus() {
	s.statusBar.SetText(fmt.Sprintf("[green]Connected to:[white] %s [green]Baud rate:[white] %d", s.portName, s.baudRate))
}

var _ types.ViewInterface = (*StatusBar)(nil)
