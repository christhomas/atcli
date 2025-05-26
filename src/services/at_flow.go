package services

import (
	"atcli/src/types"
	"fmt"
	"time"
)

// ATFlowRunner runs a sequence of ATFlowSteps, sending commands and waiting for expected responses.
type ATFlowRunner struct {
	eventBus        *EventBus
	steps           []types.ATFlowStep
	responseTimeout time.Duration
	logf            func(string)
}

// NewATFlowRunner creates a new ATFlowRunner.
func NewATFlowRunner(eventBus *EventBus, steps []types.ATFlowStep, timeout time.Duration, logf func(string)) *ATFlowRunner {
	return &ATFlowRunner{
		eventBus:        eventBus,
		steps:           steps,
		responseTimeout: timeout,
		logf:            logf,
	}
}

// Run executes the AT command flow. Returns true if all steps succeed, false otherwise.
func (r *ATFlowRunner) Run() bool {
	responsesCh := make(chan string, 10)
	stopCh := make(chan struct{})

	// Subscribe to serial responses
	handler := func(event types.Event) {
		if resp, ok := event.Payload.(string); ok {
			responsesCh <- resp
		}
	}
	r.eventBus.Subscribe(types.EventSerialResponse, handler)

	succeeded := true

	for i, step := range r.steps {
		r.logf(fmt.Sprintf("[ATFlow] Step %d: Sending '%s'", i+1, step.Command))
		r.eventBus.Publish(types.Event{
			Type:    types.EventATModemCommand,
			Payload: step.Command,
		})

		// Wait for all expected responses
		expected := make(map[string]bool)
		for _, exp := range step.ExpectedResponses {
			expected[exp] = false
		}

		deadline := time.Now().Add(r.responseTimeout)
		stepSucceeded := true
		for len(expected) > 0 {
			select {
			case resp := <-responsesCh:
				for exp := range expected {
					if exp != "" && containsResponse(resp, exp) {
						r.logf(fmt.Sprintf("[ATFlow] Step %d: Got expected response: '%s'", i+1, exp))
						delete(expected, exp)
					}
				}
			case <-time.After(time.Until(deadline)):
				// Timeout
				for exp := range expected {
					r.logf(fmt.Sprintf("[ATFlow] Step %d: Timeout waiting for response: '%s'", i+1, exp))
				}
				stepSucceeded = false
				break
			}
			if !stepSucceeded {
				break
			}
		}
		if !stepSucceeded {
			succeeded = false
			break
		}
	}
	close(stopCh)
	return succeeded
}

// containsResponse checks if the response contains the expected substring (case-insensitive).
func containsResponse(resp, exp string) bool {
	return len(exp) > 0 && len(resp) > 0 && (resp == exp || containsIgnoreCase(resp, exp))
}

func containsIgnoreCase(a, b string) bool {
	return len(a) >= len(b) && (a == b || (len(a) > len(b) && containsIgnoreCase(a[1:], b))) || (len(a) >= len(b) && (a[:len(b)] == b || a[len(a)-len(b):] == b))
}
