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

	flowLock  sync.Mutex // Ensures only one flow or command at a time
	flowOwner string     // Owner ID for re-entrant lock
	flowCond  *sync.Cond // For waiting on flow lock
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
	self.flowCond = sync.NewCond(&self.flowLock)

	// Goroutine to read from serial and update repliesView
	go self.Read()

	// The caller is responsible for acquiring the flow lock and setting OwnerID in the payload for flows.
	// For single commands, lock acquisition should be done by the caller as well if needed.
	eventBus.Subscribe(types.EventATModemCommand, self.Write)

	// Subscribe to EventATModemFlow for running multi-step flows
	// The payload is expected to be []services.ATFlowStep for now
	// In the future, this can be extended to a struct with more metadata
	eventBus.Subscribe(types.EventATModemFlow, func(event types.Event) {
		go self.RunFlow(event)
	})

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

				s.eventBus.Publish(types.Event{Type: types.EventSerialResponse, Payload: strings.TrimSpace(line)})
				partial = ""
			}
			mu.Unlock()
		}
	}
}

// Write writes to the serial port. The event.Payload should be ATCommandPayload, with Command and OwnerID fields.
func (s *SerialPort) Write(event types.Event) {
	mu := sync.Mutex{}
	payload, ok := event.Payload.(types.ATCommandPayload)
	if !ok {
		// fallback for legacy string payloads
		command, ok := event.Payload.(string)
		if !ok {
			s.eventBus.Publish(types.Event{Type: types.EventSerialError, Payload: fmt.Errorf("invalid AT command payload")})
			return
		}
		payload = types.ATCommandPayload{Command: command, OwnerID: ""}
	}
	ownerID := payload.OwnerID
	command := payload.Command

	// Check flow lock ownership
	s.flowLock.Lock()
	if s.flowOwner != "" && s.flowOwner != ownerID {
		s.flowLock.Unlock()
		mu.Lock()
		s.eventBus.Publish(types.Event{Type: types.EventSerialError, Payload: fmt.Errorf("flow lock held by another owner: ownerId=%s", s.flowOwner)})
		mu.Unlock()
		return
	}
	s.flowLock.Unlock()

	// Send the command to the serial port
	_, err := s.port.Write([]byte(command + "\r\n"))

	// If you want to debug your requests uncomment this
	LogMessage(fmt.Sprintf("-> %s", command))

	if err != nil {
		mu.Lock()
		s.eventBus.Publish(types.Event{Type: types.EventSerialError, Payload: err})
		mu.Unlock()
	} else {
		mu.Lock()
		// Echo the command with the prefix to the command view
		s.eventBus.Publish(types.Event{Type: types.EventSerialResponse, Payload: strings.TrimSpace(command)})
		mu.Unlock()
	}
}

// RunFlow executes a multi-step AT command flow with flow lock management.
// The event.Payload must be []services.ATFlowStep.
func (s *SerialPort) RunFlow(event types.Event) {
	ownerID := fmt.Sprintf("flow-%d", time.Now().UnixNano())
	if err := s.AcquireFlowLock(ownerID, 120*time.Second); err != nil {
		s.eventBus.Publish(types.Event{Type: types.EventSerialError, Payload: fmt.Errorf("could not acquire flow lock for flow: %w", err)})
		return
	}
	defer s.ReleaseFlowLock(ownerID)

	steps, ok := event.Payload.([]types.ATFlowStep)
	if !ok {
		s.eventBus.Publish(types.Event{Type: types.EventSerialError, Payload: fmt.Errorf("invalid flow payload: expected []ATFlowStep")})
		return
	}

	// FIXME: This is not working logic, but it actually does work
	// FIXME: However, stepping through the logic, I can clearly see how it's
	// FIXME: Pretty awful and doesn't really work. It just accidentally works

	for _, step := range steps {
		responsesCh := make(chan string, len(step.ExpectedResponses))
		received := make(map[string]bool)

		// Handler for serial responses
		handler := func(ev types.Event) {
			resp, ok := ev.Payload.(string)
			if !ok {
				return
			}
			// Ignore command echo
			if strings.HasPrefix(resp, "-> ") {
				return
			}
			if len(resp) == 0 {
				return
			}
			for _, expected := range step.ExpectedResponses {
				LogMessage("expected: '" + expected + "', resp: '" + resp + "'")
				if resp != expected {
					return
				}

				responsesCh <- expected
				received[expected] = true
			}
		}

		// Subscribe
		s.eventBus.Subscribe(types.EventSerialResponse, handler)

		payload := types.ATCommandPayload{Command: step.Command, OwnerID: ownerID}
		s.Write(types.Event{Type: types.EventATModemCommand, Payload: payload})

		// Wait for all expected responses or timeout
		allReceived := false
		waitTimeout := time.After(3 * time.Second)
		for count := 0; count < len(step.ExpectedResponses); {
			select {
			case <-responsesCh:
				count++
			case <-waitTimeout:
				s.eventBus.Publish(types.Event{Type: types.EventSerialError, Payload: fmt.Errorf("timeout waiting for response to '%s'", step.Command)})
				allReceived = false
				goto unsubscribe
			}
		}
		allReceived = true

	unsubscribe:
		// Unsubscribe handler
		s.eventBus.Unsubscribe(types.EventSerialResponse, handler)
		if !allReceived {
			break // abort flow on timeout
		}
	}
}

// AcquireFlowLock tries to acquire the flow lock for the given ownerID, with a timeout. Returns error if not acquired.
func (s *SerialPort) AcquireFlowLock(ownerID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	s.flowLock.Lock()
	defer s.flowLock.Unlock()
	for s.flowOwner != "" && s.flowOwner != ownerID {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return fmt.Errorf("timeout acquiring flow lock")
		}
		s.flowCond.Wait()
	}
	s.flowOwner = ownerID
	return nil
}

// ReleaseFlowLock releases the flow lock if held by the given ownerID.
func (s *SerialPort) ReleaseFlowLock(ownerID string) {
	s.flowLock.Lock()
	if s.flowOwner == ownerID {
		s.flowOwner = ""
		s.flowCond.Broadcast()
	}
	s.flowLock.Unlock()
}
