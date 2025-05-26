package cmd

import (
	"atcli/src/services"
	"atcli/src/types"
	"strings"
)

type CommandManager struct {
	commands map[string]*types.Command
	eventBus *services.EventBus
}

func NewCommandManager(eventBus *services.EventBus) *CommandManager {
	cm := &CommandManager{
		commands: map[string]*types.Command{},
		eventBus: eventBus,
	}

	// Subscribe to command events once during initialization
	eventBus.Subscribe(types.EventCommandSent, cm.handleCommandSent)

	return cm
}

// RegisterCommand registers any CommandInterface as a command.
func (c *CommandManager) RegisterCommand(obj types.CommandInterface) {
	cmd := &types.Command{
		Name:        obj.GetName(),
		Description: obj.GetDescription(),
		Run:         obj.Run,
	}

	c.commands[cmd.Name] = cmd
}

func (c *CommandManager) handleCommandSent(event types.Event) {
	request := event.Payload.(string)

	if !strings.HasPrefix(request, "/") {
		request = "/atmodem " + request
	}

	c.handleSlashCommand(request)
}

func (c *CommandManager) handleSlashCommand(request string) {
	// Remove the leading slash
	commandText := strings.TrimPrefix(request, "/")

	// Split into command name and arguments
	parts := strings.Fields(commandText)
	if len(parts) == 0 {
		return
	}

	commandName := parts[0]
	args := parts[1:]

	// Find the command in registered commands
	if cmd, ok := c.commands[commandName]; ok {
		// Execute the command with arguments
		err := cmd.Run(args)
		if err != nil {
			// Publish error event if needed
			c.eventBus.Publish(types.Event{
				Type:    types.EventLogMessage,
				Payload: "Error executing command: " + err.Error(),
			})
		}
	} else {
		// Command not found
		c.eventBus.Publish(types.Event{
			Type:    types.EventLogMessage,
			Payload: "Unknown command: /" + commandName,
		})
	}
}

// // Execute parses and runs a slash command string.
// func Execute(input string) error {
// 	if !strings.HasPrefix(input, "/") {
// 		return nil // Not a command
// 	}
// 	parts := strings.Fields(input[1:])
// 	if len(parts) == 0 {
// 		return nil
// 	}
// 	cmd, ok := commands[parts[0]]
// 	if !ok {
// 		return nil // Unknown command
// 	}
// 	return cmd.Run(parts[1:])
// }

// ExecuteCommand handles slash commands and UI popups.
// Returns true if the input was a slash command.
// func ExecuteCommand(app *tview.Application, setActiveScreen func(string), input string) bool {
// 	// Special handling for help command with UI modal
// 	if input == "/help" || strings.HasPrefix(input, "/help ") {
// 		modal := tview.NewModal().
// 			SetText("atcli v0.1\nA modern AT command terminal.\nType /help for this message.").
// 			AddButtons([]string{"OK"}).
// 			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
// 				setActiveScreen("home")
// 			})
// 		app.SetRoot(modal, true)
// 		return true
// 	}

// 	// For all other commands, execute them in a goroutine to prevent UI blocking
// 	if strings.HasPrefix(input, "/") {
// 		parts := strings.Fields(input[1:])
// 		if len(parts) == 0 {
// 			return true
// 		}

// 		cmd, ok := commands[parts[0]]
// 		if !ok {
// 			return true // Unknown command
// 		}

// 		// Execute command in a goroutine to avoid blocking the UI
// 		go func() {
// 			_ = cmd.Run(parts[1:])
// 		}()
// 		return true
// 	}

// 	return false
// }

// FIXME: when EventCommandSent it published
// TODO: also need to deal with slash command
// if strings.HasPrefix(userInput, "/") {
// 	cmd.ExecuteCommand(app, setActiveScreen, userInput)
// 	return
// }

// ListCommands returns all registered commands.
func (c *CommandManager) ListCommands() []*types.Command {
	list := make([]*types.Command, 0, len(c.commands))
	for _, c := range c.commands {
		list = append(list, c)
	}
	return list
}
