package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"atcli/src/cmd"
	"atcli/src/layouts"
	"atcli/src/services"
	"atcli/src/types"
	"atcli/src/views"
)

// Version information
var (
	Version   = "0.1"
	BuildTime = "not set"
	GitCommit = "not set"
)

func main() {
	version := flag.Bool("version", false, "Print version information and exit")
	portName := flag.String("port", "/dev/serial0", "Serial port to use")
	baudRate := flag.Int("baud", 115200, "Baud rate")
	flag.Parse()

	if *version {
		fmt.Printf("ATCLI - AT Command Line Interface\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	app := tview.NewApplication()
	app.EnableMouse(true)

	// Initialise the event bus to send messages between components
	eventBus := services.NewEventBus()

	// Initialize the log service with the event bus
	services.InitLogService(eventBus)

	viewManager := views.NewViewManager()
	layoutManager := layouts.NewLayoutManager(app, eventBus)

	inputField := views.NewInputField(eventBus, "Command: ", tcell.ColorBlue)
	viewManager.Register(inputField)

	commandView := views.NewCommandView(eventBus, app, "Sent Commands", tcell.ColorBlack)
	viewManager.Register(commandView)

	replyView := views.NewReplyView(eventBus, app, "Modem Replies")
	viewManager.Register(replyView)

	signalView := views.NewSignalChart("Signal Strength", app, eventBus)
	viewManager.Register(signalView)

	// Create and register the GPS view
	gpsView := views.NewGPSView("GPS Location", app, eventBus)
	viewManager.Register(gpsView)

	// Create and register the log view
	logView := views.NewLogView(app, eventBus)
	viewManager.Register(logView)

	statusBar := views.NewStatusBar()
	viewManager.Register(statusBar)
	statusBar.SetPortName(*portName)
	statusBar.SetBaudRate(*baudRate)

	// Create command manager and register commands
	cmdManager := cmd.NewCommandManager(eventBus)
	cmdManager.RegisterCommand(cmd.NewHelpCommand(cmdManager, eventBus, app))
	cmdManager.RegisterCommand(cmd.NewQuitCommand(eventBus))
	cmdManager.RegisterCommand(cmd.NewATModemCommand(eventBus))
	cmdManager.RegisterCommand(cmd.NewSignalCommand(eventBus))
	cmdManager.RegisterCommand(cmd.NewLogCommand(eventBus, logView))
	cmdManager.RegisterCommand(cmd.NewGPSCommand(eventBus))

	layoutManager.Register(layouts.NewHomeLayout(viewManager, eventBus), true)
	layoutManager.Register(layouts.NewSignalChartLayout(viewManager, eventBus), false)
	layoutManager.Register(layouts.NewGPSLayout(viewManager, eventBus), false)
	layoutManager.Register(layouts.NewHelpLayout(eventBus, cmdManager), false)

	serialPort := services.NewSerialPort(eventBus, app, *portName, *baudRate)
	defer serialPort.Close()

	eventBus.Subscribe(types.EventAppRedraw, func(event types.Event) {
		app.Draw()
	})

	eventBus.Subscribe(types.EventAppFocus, func(event types.Event) {
		app.SetFocus(event.Payload.(tview.Primitive))
	})

	eventBus.Subscribe(types.EventAppShutdown, func(event types.Event) {
		// Clean up resources before exiting
		serialPort.Close()

		// Stop the application
		app.Stop()
	})

	// Set up root and focus input field
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
