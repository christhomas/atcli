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

type GPSView struct {
	gpsView     *tview.TextView
	eventBus    *services.EventBus
	app         *tview.Application
	stopped     bool
	latitude    float64
	longitude   float64
	altitude    float64
	satellites  int
	lastUpdated time.Time
	utcTime     string // Store the last parsed UTC time from GPS
	date        string // Store the last parsed date (YYYYMMDD)
}

func NewGPSView(title string, app *tview.Application, eventBus *services.EventBus) *GPSView {
	gpsView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)

	gpsView.SetTextAlign(tview.AlignCenter)

	// Create the GPS view instance
	self := &GPSView{
		gpsView:     gpsView,
		eventBus:    eventBus,
		app:         app,
		stopped:     true, // Start in stopped state
		latitude:    0,
		longitude:   0,
		altitude:    0,
		satellites:  0,
		lastUpdated: time.Time{},
	}

	gpsView.
		SetTitle(fmt.Sprintf(" %s ", title)).
		SetBorder(true).
		SetBackgroundColor(tcell.ColorBlack)

	gpsView.SetScrollable(false)
	gpsView.SetChangedFunc(self.SetChanged)

	// Subscribe to modem responses, start and stop GPS events
	eventBus.Subscribe(types.EventSerialResponse, self.handleModemResponse)
	eventBus.Subscribe(types.EventStopGPS, self.handleStopGPS)
	eventBus.Subscribe(types.EventStartGPS, self.handleStartGPS)

	// Set initial content
	self.gpsView.SetText("[yellow]GPS monitoring inactive[white]\n\nUse /gps to start monitoring")

	return self
}

func (g *GPSView) SetChanged() {
	g.eventBus.Publish(types.Event{Type: types.EventAppRedraw})
}

// monitorGPS periodically queries the modem for GPS information
func (g *GPSView) monitorGPS() {
	// Wait a bit for the application to initialize
	time.Sleep(1 * time.Second)

	// Query GPS every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Send the first query immediately
	g.queryGPS()

	for range ticker.C {
		if g.stopped {
			return
		}
		g.queryGPS()
	}
}

// queryGPS sends the AT+CGPSINFO command to query GPS information
func (g *GPSView) queryGPS() {
	// Send AT+CGPSINFO command to the modem
	g.eventBus.Publish(types.Event{
		Type:    types.EventATModemCommand,
		Payload: "AT+CGPSINFO",
	})
}

// handleModemResponse processes responses from the modem
func (g *GPSView) handleModemResponse(event types.Event) {
	if g.stopped {
		return
	}

	if response, ok := event.Payload.(string); ok {
		// Skip the command echo (lines starting with "->")
		if strings.HasPrefix(response, "-> ") {
			return
		}

		// Check if this is a GPS response
		if strings.Contains(response, "+CGPSINFO:") {
			services.LogMessage(response)
			g.parseGPSResponse(response)
		}
	}
}

// parseGPSResponse extracts the GPS information from the response
func (g *GPSView) parseGPSResponse(response string) {
	// Extract GPS coordinates using regex
	// Format: +CGPSINFO: <lat>,<N/S>,<lon>,<E/W>,<date>,<UTC time>,<alt>,<speed>,<course>
	re := regexp.MustCompile(`\+CGPSINFO:\s*([^,]*),([^,]*),([^,]*),([^,]*),([^,]*),([^,]*),([^,]*),([^,]*),([^,]*)`)
	matches := re.FindStringSubmatch(response)

	if len(matches) >= 10 {
		// Check if GPS data is available
		if matches[1] == "" || matches[3] == "" {
			g.updateGPSDisplay(false)
			return
		}

		// Parse latitude (format: ddmm.mmmm)
		latStr := matches[1]
		latDir := matches[2]
		lat, err := g.parseCoordinate(latStr)
		if err == nil {
			if latDir == "S" {
				lat = -lat
			}
			g.latitude = lat
		}

		// Parse longitude (format: dddmm.mmmm)
		lonStr := matches[3]
		lonDir := matches[4]
		lon, err := g.parseCoordinate(lonStr)
		if err == nil {
			if lonDir == "W" {
				lon = -lon
			}
			g.longitude = lon
		}
		// Parse date (format: DDMMYY)
		dateStr := matches[5]
		if len(dateStr) == 6 {
			g.date = fmt.Sprintf("%s-%s-%s", dateStr[0:2], dateStr[2:4], dateStr[4:6])
		} else {
			g.date = "N/A"
		}

		// Parse UTC time (format: hhmmss.ss as float)
		utcTimeStr := matches[6]
		utcTimeFloat, err := strconv.ParseFloat(utcTimeStr, 64)
		var utcTimeFormatted string
		if err == nil {
			utcInt := int(utcTimeFloat)
			hh := utcInt / 10000
			mm := (utcInt / 100) % 100
			ss := utcInt % 100
			utcTimeFormatted = fmt.Sprintf("%02d:%02d:%02d UTC", hh, mm, ss)
		} else {
			utcTimeFormatted = "N/A"
		}
		g.utcTime = utcTimeFormatted

		// Parse altitude (meters)
		altStr := matches[7]
		alt, err := strconv.ParseFloat(altStr, 64)
		if err == nil {
			g.altitude = alt
		}

		g.lastUpdated = time.Now()
		g.updateGPSDisplay(true)
	} else {
		g.updateGPSDisplay(false)
	}
}

// parseCoordinate converts NMEA format (ddmm.mmmm) to decimal degrees
func (g *GPSView) parseCoordinate(coord string) (float64, error) {
	if coord == "" {
		return 0, fmt.Errorf("empty coordinate")
	}

	// Find the decimal point position
	decimalPos := strings.Index(coord, ".")
	if decimalPos < 0 {
		return 0, fmt.Errorf("invalid coordinate format")
	}

	// For latitude: first 2 digits are degrees, rest is minutes
	// For longitude: first 3 digits are degrees, rest is minutes
	var degreeEndPos int
	if len(coord) >= 5 && decimalPos >= 3 { // Longitude (dddmm.mmmm)
		degreeEndPos = decimalPos - 2
	} else { // Latitude (ddmm.mmmm)
		degreeEndPos = decimalPos - 2
	}

	if degreeEndPos <= 0 {
		return 0, fmt.Errorf("invalid coordinate format")
	}

	// Extract degrees and minutes
	degrees, err := strconv.ParseFloat(coord[:degreeEndPos], 64)
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.ParseFloat(coord[degreeEndPos:], 64)
	if err != nil {
		return 0, err
	}

	// Convert to decimal degrees
	return degrees + (minutes / 60.0), nil
}

// updateGPSDisplay updates the GPS display
func (g *GPSView) updateGPSDisplay(hasData bool) {
	var displayText string

	if !hasData {
		displayText = "\n[yellow]Waiting for GPS signal...[white]\n\nMake sure the GPS antenna is connected\nand has a clear view of the sky."
	} else {
		// Format coordinates for display
		latDir := "N"
		if g.latitude < 0 {
			latDir = "S"
			g.latitude = -g.latitude
		}

		lonDir := "E"
		if g.longitude < 0 {
			lonDir = "W"
			g.longitude = -g.longitude
		}

		// Calculate degrees, minutes, seconds for latitude
		latDeg := int(g.latitude)
		latMin := int((g.latitude - float64(latDeg)) * 60)
		latSec := (g.latitude - float64(latDeg) - float64(latMin)/60) * 3600

		// Calculate degrees, minutes, seconds for longitude
		lonDeg := int(g.longitude)
		lonMin := int((g.longitude - float64(lonDeg)) * 60)
		lonSec := (g.longitude - float64(lonDeg) - float64(lonMin)/60) * 3600

		// Format the coordinates in different formats
		decimalFormat := fmt.Sprintf("\n[green]Decimal Degrees:[white]\nLatitude: %.6f° %s\nLongitude: %.6f° %s",
			g.latitude, latDir, g.longitude, lonDir)

		dmsFormat := fmt.Sprintf("\n[green]Degrees, Minutes, Seconds:[white]\nLatitude: %d° %d' %.2f\" %s\nLongitude: %d° %d' %.2f\" %s",
			latDeg, latMin, latSec, latDir, lonDeg, lonMin, lonSec, lonDir)

		altitudeInfo := ""
		if g.altitude != 0 {
			altitudeInfo = fmt.Sprintf("\n\n[green]Altitude:[white] %.1f meters", g.altitude)
		}

		// Google Maps link
		mapsLink := fmt.Sprintf("\n\n[blue]Google Maps:[white]\nhttps://maps.google.com/?q=%.6f,%.6f", g.latitude, g.longitude)

		// Update timestamp
		timestamp := g.lastUpdated.Format("15:04:05")
		timeInfo := fmt.Sprintf("\n\nLast updated: %s", timestamp)

		// Combine all information
		utcInfo := ""
		if g.utcTime != "" {
			utcInfo = fmt.Sprintf("\n\n[green]GPS UTC Time:[white] %s", g.utcTime)
		}
		dateInfo := ""
		if g.date != "" {
			dateInfo = fmt.Sprintf("\n[green]GPS Date:[white] %s", g.date)
		}
		displayText = decimalFormat + "\n" + dmsFormat + altitudeInfo + mapsLink + utcInfo + dateInfo + timeInfo
	}

	// Update the text view
	g.gpsView.SetText(displayText)

	// Publish an event to indicate the GPS has been updated
	g.eventBus.Publish(types.Event{
		Type: types.EventGPSUpdated,
	})

	// Emit EventUpdateTime for status bar with UTC time, date, and last update
	if g.utcTime != "" && g.utcTime != "N/A" {
		g.eventBus.Publish(types.Event{
			Type: types.EventUpdateTime,
			Payload: map[string]interface{}{
				"utc":         g.utcTime,
				"date":        g.date,
				"lastUpdated": g.lastUpdated,
			},
		})
	}
}

func (g *GPSView) GetName() string {
	return "gps"
}

func (g *GPSView) GetComponent() tview.Primitive {
	return g.gpsView
}

// handleStopGPS handles the stop GPS event
func (g *GPSView) handleStopGPS(event types.Event) {
	g.Stop()
}

// handleStartGPS handles the start GPS event
func (g *GPSView) handleStartGPS(event types.Event) {
	g.Start()
}

// Start starts the GPS monitoring
func (g *GPSView) Start() {
	// GPS initialization flow
	initFlow := []types.ATFlowStep{
		{Command: "AT+CGNSSPWR=0", ExpectedResponses: []string{"OK"}},
		{Command: "AT+CGNSSPWR=1", ExpectedResponses: []string{"OK", "+CGNSSPWR: READY!"}},
		{Command: "AT+CGNSSTST=1", ExpectedResponses: []string{"OK"}},
		{Command: "AT+CGNSSPORTSWITCH=0,1", ExpectedResponses: []string{"OK"}},
	}
	g.eventBus.Publish(types.Event{
		Type:    types.EventATModemFlow,
		Payload: initFlow,
	})

	if g.stopped {
		g.stopped = false
		// Set initial content
		g.gpsView.SetText("[yellow]GPS monitoring active[white]\n\nWaiting for GPS data...")
	}
	// Start the GPS monitoring loop
	go g.monitorGPS()
}

// Stop stops the GPS monitoring
func (g *GPSView) Stop() {
	// GPS stop flow
	stopFlow := []types.ATFlowStep{
		{Command: "AT+CGNSSTST=0", ExpectedResponses: []string{"OK"}},
		{Command: "AT+CGNSSPORTSWITCH=0,0", ExpectedResponses: []string{"OK"}},
		{Command: "AT+CGNSSPWR=0", ExpectedResponses: []string{"OK"}},
	}
	g.eventBus.Publish(types.Event{
		Type:    types.EventATModemFlow,
		Payload: stopFlow,
	})

	if !g.stopped {
		g.stopped = true
		// Update the view to indicate monitoring is stopped
		g.gpsView.SetText("[yellow]GPS monitoring stopped[white]\n\nUse /gps to restart monitoring")
	}
}

var _ types.ViewInterface = (*GPSView)(nil)
