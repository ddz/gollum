package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
)

// CommandAction represents different special command actions
type CommandAction int

const (
	ActionContinue CommandAction = iota // Continue the input loop
	ActionExit                          // Exit the application
	ActionClear                         // Clear screen
	ActionNew                           // Start new conversation
	ActionHelp                          // Show help
	ActionUnknown                       // Unknown command
)

// CommandResult represents the result of processing a special command
type CommandResult struct {
	Action  CommandAction
	Message string // Optional message to display
}

// CommandHandler is a function type for handling special commands
type CommandHandler func() CommandResult

// UserInputHandler encapsulates readline functionality for user input handling
type UserInputHandler struct {
	rl       *readline.Instance
	handlers map[string]CommandHandler // Map of command name to handler function
}

// completer provides auto-completion for common commands
var completer = readline.NewPrefixCompleter(
	readline.PcItem("/exit"),
	readline.PcItem("/quit"),
	readline.PcItem("/help"),
	readline.PcItem("/clear"),
	readline.PcItem("/new"),
)

// filterInput filters input runes
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// NewUserInputHandler creates a new UserInputHandler with readline configuration
func NewUserInputHandler() (*UserInputHandler, error) {
	// Create readline instance with history and editing support
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "~~> ",
		HistoryFile:     ".gollum_history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating readline: %w", err)
	}

	handler := &UserInputHandler{
		rl:       rl,
		handlers: make(map[string]CommandHandler),
	}

	// Register built-in commands
	handler.registerBuiltinCommands()

	// Update auto-completion with registered commands
	handler.UpdateAutoComplete()

	return handler, nil
}

// Close closes the readline instance
func (h *UserInputHandler) Close() error {
	return h.rl.Close()
}

// ReadLine reads a line of input from the user
// Returns the input string and an error if any
func (h *UserInputHandler) ReadLine() (string, error) {
	userInput, err := h.rl.Readline()
	if err != nil {
		return userInput, err
	}
	return strings.TrimSpace(userInput), nil
}

// ClearScreen clears the terminal screen
func (h *UserInputHandler) ClearScreen() {
	readline.ClearScreen(h.rl)
}

// RegisterCommand registers a handler for a special command
// commandName should not include the leading '/' - it will be added automatically
func (h *UserInputHandler) RegisterCommand(commandName string, handler CommandHandler) {
	h.handlers[strings.ToLower(commandName)] = handler
}

// UnregisterCommand removes a handler for a special command
func (h *UserInputHandler) UnregisterCommand(commandName string) {
	delete(h.handlers, strings.ToLower(commandName))
}

// GetRegisteredCommands returns a list of all registered command names (without the '/' prefix)
func (h *UserInputHandler) GetRegisteredCommands() []string {
	commands := make([]string, 0, len(h.handlers))
	for cmd := range h.handlers {
		commands = append(commands, cmd)
	}
	return commands
}

// UpdateAutoComplete updates the auto-completion based on registered commands
func (h *UserInputHandler) UpdateAutoComplete() {
	items := make([]readline.PrefixCompleterInterface, 0, len(h.handlers))
	for cmd := range h.handlers {
		items = append(items, readline.PcItem("/"+cmd))
	}

	// Create new completer with current commands
	newCompleter := readline.NewPrefixCompleter(items...)
	h.rl.Config.AutoComplete = newCompleter
}

// IsInterruptError checks if the error is a readline interrupt error
func IsInterruptError(err error) bool {
	return err == readline.ErrInterrupt
}

// IsEOFError checks if the error is an EOF error
func IsEOFError(err error) bool {
	return err == io.EOF
}

// registerBuiltinCommands registers the default built-in special commands
func (h *UserInputHandler) registerBuiltinCommands() {
	// Register exit/quit commands
	exitHandler := func() CommandResult {
		return CommandResult{
			Action:  ActionExit,
			Message: "Goodbye!",
		}
	}
	h.RegisterCommand("exit", exitHandler)
	h.RegisterCommand("quit", exitHandler)

	// Register clear command
	h.RegisterCommand("clear", func() CommandResult {
		h.ClearScreen()
		return CommandResult{Action: ActionContinue}
	})

	// Note: 'new' command is not registered here - it should be registered
	// by the main application with access to the conversation context

	// Register help command
	h.RegisterCommand("help", func() CommandResult {
		// Get all registered commands for dynamic help
		commands := h.GetRegisteredCommands()

		helpMsg := `
Special commands:`

		// Add built-in commands with descriptions
		builtinCommands := map[string]string{
			"exit":  "Exit the application",
			"quit":  "Exit the application",
			"clear": "Clear the screen",
			"new":   "Start a new conversation",
			"help":  "Show this help",
		}

		for _, cmd := range commands {
			if desc, exists := builtinCommands[cmd]; exists {
				helpMsg += fmt.Sprintf("\n  /%s - %s", cmd, desc)
			} else {
				helpMsg += fmt.Sprintf("\n  /%s - Custom command", cmd)
			}
		}

		helpMsg += `

Keyboard shortcuts:
  Ctrl+R       - Reverse history search
  Ctrl+C       - Interrupt current input
  Ctrl+D       - Exit (EOF)
  Up/Down      - Navigate history`

		return CommandResult{
			Action:  ActionHelp,
			Message: helpMsg,
		}
	})
}

// UserInput reads input from the user and processes any special commands
// Returns:
// - userInput: the input to process (never empty unless error)
// - err: any error that occurred, including EOF for clean exit
func (h *UserInputHandler) UserInput() (userInput string, err error) {
	for {
		// Read input from user
		input, err := h.ReadLine()
		if IsInterruptError(err) {
			if len(input) == 0 {
				fmt.Println("\nGoodbye!")
				return "", io.EOF // Use EOF to signal clean exit
			} else {
				continue // Ignore interrupted input, try again
			}
		} else if IsEOFError(err) {
			fmt.Println("\nGoodbye!")
			return "", err // Return EOF for clean exit
		} else if err != nil {
			return "", err // Return the error for handling by caller
		}

		// Process the input (handle special commands)
		shouldContinue, processedInput := h.ProcessInput(input)
		if !shouldContinue {
			fmt.Println("Goodbye!")
			return "", io.EOF // Use EOF to signal clean exit
		}

		// If we have actual user input (not a special command), return it
		if processedInput != "" {
			return processedInput, nil
		}

		// processedInput was empty (special command processed), continue loop
	}
}

// Returns true if the application should continue, false if it should exit
// Returns the processed input (empty string if it was a special command)
func (h *UserInputHandler) ProcessInput(input string) (shouldContinue bool, processedInput string) {
	if input == "" {
		return true, ""
	}

	// Check if it's a special command
	if !strings.HasPrefix(input, "/") {
		return true, input // Not a special command, return as-is
	}

	// Process the special command
	cmdResult := h.ProcessSpecialCommand(input)

	switch cmdResult.Action {
	case ActionExit:
		if cmdResult.Message != "" {
			fmt.Println(cmdResult.Message)
		}
		return false, "" // Signal to exit
	case ActionClear:
		return true, "" // Continue with empty input (command processed)
	case ActionHelp, ActionUnknown:
		if cmdResult.Message != "" {
			fmt.Println(cmdResult.Message)
		}
		return true, "" // Continue with empty input (command processed)
	case ActionContinue:
		return true, "" // Command processed, continue with empty input
	default:
		return true, input // Unknown action, treat as regular input
	}
}

// ProcessSpecialCommand processes special commands that start with '/'
// Returns CommandResult indicating what action should be taken
func (h *UserInputHandler) ProcessSpecialCommand(input string) CommandResult {
	if !strings.HasPrefix(input, "/") {
		return CommandResult{Action: ActionContinue}
	}

	// Remove the '/' prefix and convert to lowercase
	commandName := strings.ToLower(strings.TrimPrefix(input, "/"))

	// Look up the handler for this command
	if handler, exists := h.handlers[commandName]; exists {
		return handler()
	}

	// Unknown command
	return CommandResult{
		Action:  ActionUnknown,
		Message: fmt.Sprintf("Unknown command: %s\nType '/help' to see available commands.", input),
	}
}
