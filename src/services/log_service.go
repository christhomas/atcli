package services

import (
	"atcli/src/types"
	"fmt"
	"sync"
	"time"
)

var logMutex sync.Mutex
var logEventBus *EventBus

// InitLogService initializes the log service with the event bus
func InitLogService(eventBus *EventBus) {
	logEventBus = eventBus
}

// LogMessage adds a message to the log system
// It formats the message with a timestamp and publishes it to the event bus
// so that any component listening for log messages can display them
func LogMessage(message string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	timestamp := time.Now().Format("15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	// Print to standard output for debugging
	// fmt.Println(formattedMessage)

	// Publish to event bus if available
	if logEventBus != nil {
		logEventBus.Publish(types.Event{
			Type:    types.EventLogMessage,
			Payload: formattedMessage,
		})
	}
}
