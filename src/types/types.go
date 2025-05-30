package types

import "github.com/rivo/tview"

// CommandInterface is the interface for any command object.
type CommandInterface interface {
	GetName() string
	GetDescription() string
	Run(args []string) error
}

// Command represents a slash command.
type Command struct {
	Name        string
	Description string
	Run         func(args []string) error
}

// ATCommandPayload is used for sending AT commands via the event bus with flow lock ownership.
type ATCommandPayload struct {
	Command string
	OwnerID string
}

// ATFlowStep represents a single step in a multi-step AT command flow.
type ATFlowStep struct {
	Command           string
	ExpectedResponses []string // All must be received before next step
}

type HistoryItem struct {
	Cmd   string
	Index int // Line index in the commands view
}

type ViewInterface interface {
	GetName() string
	GetComponent() tview.Primitive
}

type ViewMap map[string]ViewInterface

type LayoutInterface interface {
	GetName() string
	GetComponent() tview.Primitive
	OnLayoutChange()
}

type LayoutMap map[string]LayoutInterface

// EventType defines the type of event
type EventType string

// Common event types
const (
	EventAppRedraw       EventType = "app_redraw"
	EventAppFocus        EventType = "app_focus"
	EventAppShutdown     EventType = "app_shutdown"
	EventFocusInput      EventType = "focus_input"
	EventCommandSent     EventType = "command_sent"
	EventATModemCommand  EventType = "atmodem_command"
	EventATModemFlow     EventType = "atmodem_flow"
	EventCommandHistory  EventType = "command_history"
	EventInputSetCommand EventType = "input_set_command"
	EventReplyReceived   EventType = "reply_received"
	EventLogMessage      EventType = "log_message"
	EventChangeLayout    EventType = "change_layout"
	EventSignalUpdated   EventType = "signal_updated"
	EventGPSUpdated      EventType = "gps_updated"
	EventSerialError     EventType = "serial_error"
	EventSerialResponse  EventType = "serial_response"
	EventLayoutChange    EventType = "layout_change"
	EventStopSignal      EventType = "stop_signal"
	EventStartSignal     EventType = "start_signal"
	EventStopGPS         EventType = "stop_gps"
	EventStartGPS        EventType = "start_gps"
	EventUpdateTime      EventType = "update_time"
)

// Event represents an event in the system
type Event struct {
	Type    EventType
	Payload interface{}
}

// EventHandlerFunc is a function that handles an event
type EventHandlerFunc func(event Event)
