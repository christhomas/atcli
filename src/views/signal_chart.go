package views

import (
	"atcli/src/services"
	"atcli/src/types"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SignalChart struct {
	signalChartView *tview.TextView
	eventBus        *services.EventBus
	app             *tview.Application
	stopped         bool
	signalCSQ       int // Current signal strength (CSQ value)
}

func NewSignalChart(title string, app *tview.Application, eventBus *services.EventBus) *SignalChart {
	signalChartView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)

	signalChartView.SetTextAlign(tview.AlignCenter)

	// Create the signal chart instance
	self := &SignalChart{
		signalChartView: signalChartView,
		eventBus:        eventBus,
		app:             app,
		stopped:         true, // Start in stopped state
		signalCSQ:       0,
	}

	signalChartView.
		SetTitle(fmt.Sprintf(" %s ", title)).
		SetBorder(true).
		SetBackgroundColor(tcell.ColorBlack)

	signalChartView.SetScrollable(false)
	signalChartView.SetChangedFunc(self.SetChanged)

	// Subscribe to modem responses, start and stop signal events
	eventBus.Subscribe(types.EventSerialResponse, self.handleModemResponse)
	eventBus.Subscribe(types.EventStopSignal, self.handleStopSignal)
	eventBus.Subscribe(types.EventStartSignal, self.handleStartSignal)

	// Set initial content
	self.signalChartView.SetText("[yellow]Signal monitoring inactive[white]\n\nUse /signal to start monitoring")

	return self
}

func (s *SignalChart) SetChanged() {
	s.eventBus.Publish(types.Event{Type: types.EventAppRedraw})
}

// monitorSignalStrength periodically queries the modem for signal strength
func (s *SignalChart) monitorSignalStrength() {
	// Wait a bit for the application to initialize
	time.Sleep(1 * time.Second)

	// Query signal strength every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Send the first query immediately
	s.querySignalStrength()

	for range ticker.C {
		if s.stopped {
			return
		}
		s.querySignalStrength()
	}
}

// querySignalStrength sends the AT+CSQ command to query signal strength
func (s *SignalChart) querySignalStrength() {
	// Send AT+CSQ command to the modem
	s.eventBus.Publish(types.Event{
		Type:    types.EventATModemCommand,
		Payload: "AT+CSQ",
	})
}

// handleModemResponse processes responses from the modem
func (s *SignalChart) handleModemResponse(event types.Event) {
	if s.stopped {
		return
	}

	if response, ok := event.Payload.(string); ok {
		// Skip the command echo (lines starting with "->")
		if strings.HasPrefix(response, "-> ") {
			return
		}

		// Check if this is a CSQ response
		if strings.Contains(response, "+CSQ:") {
			s.parseCSQResponse(response)
		}
	}
}

// parseCSQResponse extracts the signal strength from the CSQ response
func (s *SignalChart) parseCSQResponse(response string) {
	// Extract CSQ value using regex
	re := regexp.MustCompile(`\+CSQ:\s*(\d+),\s*(\d+)`)
	matches := re.FindStringSubmatch(response)

	if len(matches) >= 3 {
		csq, err := strconv.Atoi(matches[1])
		if err == nil {
			s.signalCSQ = csq
			s.updateSignalDisplay()
		}
	}
}

// updateSignalDisplay updates the signal strength display
func (s *SignalChart) updateSignalDisplay() {
	// CSQ values range from 0-31, where:
	// 0 = -113 dBm or less (worst)
	// 31 = -51 dBm or greater (best)
	// 99 = not known or not detectable

	var signalText string
	var signalBars string
	var signalQuality string
	var signalColor string

	if s.signalCSQ == 99 {
		signalText = "Signal not detectable"
		signalBars = ""
		signalQuality = "Unknown"
		signalColor = "[gray]"
	} else {
		// Calculate signal quality percentage (0-31 scale to 0-100%)
		percentage := int(float64(s.signalCSQ) / 31.0 * 100)

		// Calculate dBm value
		dbm := -113 + (2 * s.signalCSQ)

		// Determine signal quality text and color
		if percentage < 20 {
			signalQuality = "Very Poor"
			signalColor = "[red]"
		} else if percentage < 40 {
			signalQuality = "Poor"
			signalColor = "[orange]"
		} else if percentage < 60 {
			signalQuality = "Fair"
			signalColor = "[yellow]"
		} else if percentage < 80 {
			signalQuality = "Good"
			signalColor = "[lime]"
		} else {
			signalQuality = "Excellent"
			signalColor = "[green]"
		}

		// Create signal bars
		numBars := (s.signalCSQ * 10) / 31 // 0-10 bars
		if numBars < 1 {
			numBars = 1 // Always show at least one bar if we have a signal
		}

		signalBars = signalColor
		for i := 0; i < numBars; i++ {
			signalBars += "█"
		}
		for i := numBars; i < 10; i++ {
			signalBars += "░"
		}
		signalBars += "[white]"

		// Format the signal text
		signalText = fmt.Sprintf("Signal Strength: %s%d%%[white] (%d dBm)\nCSQ Value: %d/31",
			signalColor, percentage, dbm, s.signalCSQ)
	}

	// Update the display
	timestamp := time.Now().Format("15:04:05")
	displayText := fmt.Sprintf("\n%s\n\n%s\n\n%s%s[white]\n\nLast updated: %s",
		signalText, signalBars, signalColor, signalQuality, timestamp)

	// Update the text view
	s.signalChartView.SetText(displayText)

	// Publish an event to indicate the signal has been updated
	s.eventBus.Publish(types.Event{
		Type: types.EventSignalUpdated,
	})
}

func (s *SignalChart) GetName() string {
	return "signal"
}

func (s *SignalChart) GetComponent() tview.Primitive {
	return s.signalChartView
}

// Stop stops the signal monitoring
func (s *SignalChart) Stop() {
	if !s.stopped {
		s.stopped = true
		// Update the view to indicate monitoring is stopped
		s.signalChartView.SetText("[yellow]Signal monitoring stopped[white]\n\nUse /signal to restart monitoring")
	}
}

// handleStopSignal handles the stop signal event
func (s *SignalChart) handleStopSignal(event types.Event) {
	s.Stop()
}

// handleStartSignal handles the start signal event
func (s *SignalChart) handleStartSignal(event types.Event) {
	s.Start()
}

// Start starts the signal monitoring
func (s *SignalChart) Start() {
	if s.stopped {
		s.stopped = false
		// Set initial content
		s.signalChartView.SetText("[yellow]Initializing signal monitor...[white]\n\nWaiting for first signal reading...")
		// Start the signal monitoring loop
		go s.monitorSignalStrength()
	}
}

var _ types.ViewInterface = (*SignalChart)(nil)
