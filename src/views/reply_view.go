package views

import (
	"atcli/src/services"
	"atcli/src/types"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ReplyView struct {
	eventBus     *services.EventBus
	replyView    *tview.TextView
	replyLineNum int
}

func NewReplyView(eventBus *services.EventBus, app *tview.Application, title string) *ReplyView {
	replyView := tview.NewTextView()
	replyView.SetDynamicColors(true).SetChangedFunc(func() { app.Draw() })
	replyView.SetTitle(title).SetBorder(true)
	replyView.SetBackgroundColor(tcell.ColorBlack)
	replyView.SetScrollable(true)

	self := &ReplyView{
		eventBus:     eventBus,
		replyView:    replyView,
		replyLineNum: 0,
	}

	replyView.SetInputCapture(self.SetInputCapture)

	eventBus.Subscribe(types.EventSerialError, self.SerialError)
	eventBus.Subscribe(types.EventSerialResponse, self.SerialResponse)

	return self
}

func (r *ReplyView) SetInputCapture(event *tcell.EventKey) *tcell.EventKey {
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
		r.eventBus.Publish(types.Event{Type: types.EventFocusInput})
		return event
	default:
		// For any other key, redirect to input field
		r.eventBus.Publish(types.Event{Type: types.EventFocusInput})
		return event
	}
}

func (r *ReplyView) GetName() string {
	return "reply"
}

func (r *ReplyView) GetComponent() tview.Primitive {
	return r.replyView
}

func (r *ReplyView) Append(text string) {
	r.replyView.Write([]byte(text))
	r.replyView.ScrollToEnd()
	r.replyLineNum++
}

func (r *ReplyView) SerialError(event types.Event) {
	r.Append("[red]Serial read error: " + event.Payload.(error).Error() + "\n")
}

func (r *ReplyView) SerialResponse(event types.Event) {
	clean := strings.TrimRight(event.Payload.(string), "\r\n")
	if clean != "" {
		r.Append("[" + fmt.Sprint(r.replyLineNum) + "] <- " + clean + "\n")
	}
}

var _ types.ViewInterface = (*ReplyView)(nil)
