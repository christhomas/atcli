package services

import (
	"atcli/src/types"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rivo/tview"
	"go.bug.st/serial"
)

type SerialPort struct {
	port     serial.Port
	eventBus *EventBus
}

func NewSerialPort(eventBus *EventBus, app *tview.Application, portName string, baudRate int) *SerialPort {
	mode := serial.Mode{
		BaudRate: baudRate,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, &mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open serial port %s: %v\n", portName, err)
		os.Exit(1)
	}

	// Handle Ctrl+C gracefully
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		app.Stop()
		os.Exit(0)
	}()

	self := &SerialPort{
		port:     port,
		eventBus: eventBus,
	}

	// Goroutine to read from serial and update repliesView
	go self.Read()

	eventBus.Subscribe(types.EventATModemCommand, self.Write)

	return self
}

func (s *SerialPort) Close() {
	s.port.Close()
}

func (s *SerialPort) Read() {
	// FIXME: Investigate whether this lock can deadlock the app, or whether go behaves nicely
	mu := sync.Mutex{}

	buf := make([]byte, 256)
	partial := ""
	for {
		n, err := s.port.Read(buf)
		if err != nil {
			mu.Lock()
			s.eventBus.Publish(types.Event{Type: types.EventSerialError, Payload: err})
			mu.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}
		if n > 0 {
			mu.Lock()
			incoming := partial + string(buf[:n])
			lines := strings.SplitAfter(incoming, "\n")
			for i, line := range lines {
				if i == len(lines)-1 && !strings.HasSuffix(line, "\n") {
					partial = line // Save incomplete line
					break
				}

				s.eventBus.Publish(types.Event{Type: types.EventSerialResponse, Payload: "<- " + line})
				partial = ""
			}
			mu.Unlock()
		}
	}
}

func (s *SerialPort) Write(event types.Event) {
	mu := sync.Mutex{}

	// Get the command to send
	command := event.Payload.(string)

	// Send the command to the serial port
	_, err := s.port.Write([]byte(command + "\r\n"))

	if err != nil {
		mu.Lock()
		s.eventBus.Publish(types.Event{Type: types.EventSerialError, Payload: err})
		mu.Unlock()
	} else {
		mu.Lock()
		// Echo the command with the prefix to the command view
		s.eventBus.Publish(types.Event{Type: types.EventSerialResponse, Payload: "-> " + command})
		mu.Unlock()
	}
}
