package views

import (
	"atcli/src/types"
	"fmt"
	"time"

	"atcli/src/services"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type StatusBar struct {
	flex        *tview.Flex
	leftView    *tview.TextView
	rightView   *tview.TextView
	eventBus    *services.EventBus
	lastUTCTime string
	lastDate    string
	lastUpdated time.Time
	portName    string
	baudRate    int
}

func NewStatusBar(eventBus *services.EventBus) *StatusBar {
	left := tview.NewTextView().SetDynamicColors(true)
	left.SetBackgroundColor(tcell.ColorBlack)
	left.SetTextColor(tcell.ColorWhite)
	left.SetTextAlign(tview.AlignLeft)

	right := tview.NewTextView().SetDynamicColors(true)
	right.SetBackgroundColor(tcell.ColorBlack)
	right.SetTextColor(tcell.ColorWhite)
	right.SetTextAlign(tview.AlignRight)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(left, 0, 1, false).
		AddItem(right, 0, 1, false)

	s := &StatusBar{
		flex:      flex,
		leftView:  left,
		rightView: right,
		eventBus:  eventBus,
	}

	s.eventBus.Subscribe(types.EventUpdateTime, s.handleUpdateTime)
	go s.refreshTimer()

	return s
}

func (s *StatusBar) handleUpdateTime(event types.Event) {
	data, ok := event.Payload.(map[string]interface{})
	if !ok {
		return
	}
	utc, ok := data["utc"].(string)
	if !ok {
		return
	}
	date, _ := data["date"].(string)
	lastUpdated, ok := data["lastUpdated"].(time.Time)
	if !ok {
		return
	}
	s.lastUTCTime = utc
	s.lastDate = date
	s.lastUpdated = lastUpdated
	s.updateText()
}

func (s *StatusBar) refreshTimer() {
	for {
		time.Sleep(time.Second)
		s.updateText()
	}
}

func (s *StatusBar) updateText() {
	// Left: connection info
	left := fmt.Sprintf("[green]Connected to:[white] %s [green]Baud rate:[white] %d", s.portName, s.baudRate)
	s.leftView.SetText(left)

	// Right: time/date info
	right := ""
	if s.lastUTCTime != "" {
		secsAgo := int(time.Since(s.lastUpdated).Seconds())
		if s.lastDate != "" {
			right = fmt.Sprintf("[Status] %s %s (%ds ago)", s.lastDate, s.lastUTCTime, secsAgo)
		} else {
			right = fmt.Sprintf("[Status] %s (%ds ago)", s.lastUTCTime, secsAgo)
		}
	}
	s.rightView.SetText(right)
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
	return s.flex
}

func (s *StatusBar) setStatus() {
	s.leftView.SetText(fmt.Sprintf("[green]Connected to:[white] %s [green]Baud rate:[white] %d", s.portName, s.baudRate))
}

var _ types.ViewInterface = (*StatusBar)(nil)
