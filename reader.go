package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

// CommandHandler is a function type for handling special commands
type CommandHandler func(w io.Writer) error

// BuiltinCommand represents a built-in command with its metadata and handler
type BuiltinCommand struct {
	Name        string
	Description string
	Handler     CommandHandler
}

// Reader encapsulates readline functionality for user input handling
type Reader struct {
	rl           *readline.Instance
	handlers     map[string]CommandHandler // Map of command name to handler function
	descriptions map[string]string         // Map of command name to description
}

// completer starts empty and is populated dynamically based on registered command handlers
var completer = readline.NewPrefixCompleter()

// builtinCommands defines all built-in commands with their descriptions and handlers
var builtinCommands = []BuiltinCommand{
	{
		Name:        "exit",
		Description: "Exit the application",
		Handler: func(w io.Writer) error {
			fmt.Fprintln(w, "Goodbye!")
			return io.EOF // Signal exit
		},
	},
	{
		Name:        "quit",
		Description: "Exit the application",
		Handler: func(w io.Writer) error {
			fmt.Fprintln(w, "Goodbye!")
			return io.EOF // Signal exit
		},
	},
	{
		Name:        "clear",
		Description: "Clear the screen",
		Handler: func(w io.Writer) error {
			// Note: this will be set to the actual reader's ClearScreen method during registration
			return nil
		},
	},
	{
		Name:        "help",
		Description: "Show this help",
		Handler: func(w io.Writer) error {
			// Note: this will be set to the actual help implementation during registration
			return nil
		},
	},
}

// filterInput filters input runes
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// NewReader creates a new Reader with readline configuration
func NewReader() (*Reader, error) {
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

	handler := &Reader{
		rl:           rl,
		handlers:     make(map[string]CommandHandler),
		descriptions: make(map[string]string),
	}

	// Register built-in commands
	handler.registerBuiltinCommands()

	return handler, nil
}

// Close closes the readline instance
func (r *Reader) Close() error {
	return r.rl.Close()
}

// ReadLine reads a line of input from the user
// Returns the input string and an error if any
func (r *Reader) ReadLine() (string, error) {
	userInput, err := r.rl.Readline()
	if err != nil {
		return userInput, err
	}
	return strings.TrimSpace(userInput), nil
}

// ClearScreen clears the terminal screen
func (r *Reader) ClearScreen() {
	readline.ClearScreen(r.rl)
}

// RegisterCommand registers a handler for a special command
// commandName should not include the leading '/' - it will be added automatically
// description is a human-readable description of what the command does
// The auto-completer is automatically updated after registration
func (r *Reader) RegisterCommand(commandName string, description string, handler CommandHandler) {
	lowercaseName := strings.ToLower(commandName)
	r.handlers[lowercaseName] = handler
	r.descriptions[lowercaseName] = description
	r.UpdateAutoComplete()
}

// UnregisterCommand removes a handler for a special command
// The auto-completer is automatically updated after unregistration
func (r *Reader) UnregisterCommand(commandName string) {
	lowercaseName := strings.ToLower(commandName)
	delete(r.handlers, lowercaseName)
	delete(r.descriptions, lowercaseName)
	r.UpdateAutoComplete()
}

// GetRegisteredCommands returns a list of all registered command names (without the '/' prefix)
func (r *Reader) GetRegisteredCommands() []string {
	commands := make([]string, 0, len(r.handlers))
	for cmd := range r.handlers {
		commands = append(commands, cmd)
	}
	return commands
}

// UpdateAutoComplete updates the auto-completion based on registered commands
func (r *Reader) UpdateAutoComplete() {
	items := make([]readline.PrefixCompleterInterface, 0, len(r.handlers))
	for cmd := range r.handlers {
		items = append(items, readline.PcItem("/"+cmd))
	}

	// Create new completer with current commands
	newCompleter := readline.NewPrefixCompleter(items...)
	r.rl.Config.AutoComplete = newCompleter
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
func (r *Reader) registerBuiltinCommands() {
	// Register each built-in command from the global array
	for _, cmd := range builtinCommands {
		switch cmd.Name {
		case "clear":
			// Clear command needs access to the reader's ClearScreen method
			r.RegisterCommand(cmd.Name, cmd.Description, func(w io.Writer) error {
				r.ClearScreen()
				return nil
			})
		case "help":
			// Help command needs access to the reader's state for dynamic help
			r.RegisterCommand(cmd.Name, cmd.Description, func(w io.Writer) error {
				return r.generateHelp(w)
			})
		default:
			// Use the handler directly from the command definition
			r.RegisterCommand(cmd.Name, cmd.Description, cmd.Handler)
		}
	}

	// Note: 'new' command is not registered here - it should be registered
	// by the main application with access to the conversation context
}

// generateHelp generates the help text dynamically based on registered commands
func (r *Reader) generateHelp(w io.Writer) error {
	// Get all registered commands for dynamic help
	commands := r.GetRegisteredCommands()

	fmt.Fprintln(w, "\nSpecial commands:")

	// Display registered commands with their descriptions
	for _, cmd := range commands {
		if desc, exists := r.descriptions[cmd]; exists {
			fmt.Fprintf(w, "  /%s - %s\n", cmd, desc)
		} else {
			// Fallback for commands without descriptions (shouldn't happen)
			fmt.Fprintf(w, "  /%s - No description available\n", cmd)
		}
	}

	fmt.Fprintln(w, "\nKeyboard shortcuts:")
	fmt.Fprintln(w, "  Ctrl+R       - Reverse history search")
	fmt.Fprintln(w, "  Ctrl+C       - Interrupt current input")
	fmt.Fprintln(w, "  Ctrl+D       - Exit (EOF)")
	fmt.Fprintln(w, "  Up/Down      - Navigate history")

	return nil
}

// UserInput reads input from the user and processes any special commands
// Returns:
// - userInput: the input to process (never empty unless error)
// - err: any error that occurred, including EOF for clean exit
func (r *Reader) UserInput() (userInput string, err error) {
	for {
		// Read input from user
		input, err := r.ReadLine()
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
		processedInput, err := r.ProcessInput(input)
		if err != nil {
			if err == io.EOF {
				// Exit requested by command
				return "", err
			}
			return "", err // Return error from command processing
		}

		// If we have actual user input (not a special command), return it
		if processedInput != "" {
			return processedInput, nil
		}

		// processedInput was empty (special command processed), continue loop
	}
}

// ProcessInput processes input and handles special commands
// Returns the processed input (empty string if it was a special command)
// Returns an error if command processing fails, including io.EOF for exit
func (r *Reader) ProcessInput(input string) (processedInput string, err error) {
	if input == "" {
		return "", nil
	}

	// Check if it's a special command
	if !strings.HasPrefix(input, "/") {
		return input, nil // Not a special command, return as-is
	}

	// Process the special command
	err = r.ProcessSpecialCommand(input)
	if err != nil {
		return "", err // Return error from command processing (including io.EOF for exit)
	}

	// Special command processed successfully, return empty string
	return "", nil
}

// ProcessSpecialCommand processes special commands that start with '/'
// Returns an error if command processing fails, or io.EOF if exit is requested
func (r *Reader) ProcessSpecialCommand(input string) error {
	if !strings.HasPrefix(input, "/") {
		return nil
	}

	// Remove the '/' prefix and convert to lowercase
	commandName := strings.ToLower(strings.TrimPrefix(input, "/"))

	// Look up the handler for this command
	if handler, exists := r.handlers[commandName]; exists {
		return handler(os.Stdout)
	}

	// Unknown command
	fmt.Printf("Unknown command: %s\nType '/help' to see available commands.\n", input)
	return nil
}
