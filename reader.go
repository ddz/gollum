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

// command represents a command with its metadata and handler
type command struct {
	Name        string
	Description string
	Handler     CommandHandler
}

// Reader encapsulates readline functionality for user input handling
type Reader struct {
	rl       *readline.Instance
	cmds map[string]command // Map of command name to command struct
}

// completer starts empty and is populated dynamically based on registered command handlers
var completer = readline.NewPrefixCompleter()

// builtinCommands defines all built-in commands with their descriptions and handlers
var builtinCommands = []command{
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
		rl:       rl,
		cmds: make(map[string]command),
	}

	// Register built-in commands
	handler.registerBuiltinCommands()

	return handler, nil
}

// Close closes the readline instance
func (r *Reader) Close() error {
	return r.rl.Close()
}

// clearScreen clears the terminal screen
func (r *Reader) clearScreen() {
	readline.ClearScreen(r.rl)
}

// RegisterCommand registers a handler for a special command
// commandName should not include the leading '/' - it will be added automatically
// description is a human-readable description of what the command does
// The auto-completer is automatically updated after registration
func (r *Reader) RegisterCommand(commandName string, description string, handler CommandHandler) {
	lowercaseName := strings.ToLower(commandName)
	r.cmds[lowercaseName] = command{
		Name:        lowercaseName,
		Description: description,
		Handler:     handler,
	}
	
	r.updateAutoComplete()
}

// commands returns a list of all registered command names (without the '/' prefix)
func (r *Reader) commands() []string {
	commands := make([]string, 0, len(r.cmds))
	for cmd := range r.cmds {
		commands = append(commands, cmd)
	}
	return commands
}

// updateAutoComplete updates the auto-completion based on registered commands
func (r *Reader) updateAutoComplete() {
	items := make([]readline.PrefixCompleterInterface, 0, len(r.cmds))
	for cmd := range r.cmds {
		items = append(items, readline.PcItem("/"+cmd))
	}

	// Create new completer with current commands
	newCompleter := readline.NewPrefixCompleter(items...)
	r.rl.Config.AutoComplete = newCompleter
}

// registerBuiltinCommands registers the default built-in special commands
func (r *Reader) registerBuiltinCommands() {
	// Register each built-in command from the global array
	for _, cmd := range builtinCommands {
		switch cmd.Name {
		case "clear":
			// Clear command needs access to the reader's ClearScreen method
			r.RegisterCommand(cmd.Name, cmd.Description, func(w io.Writer) error {
				r.clearScreen()
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
	commands := r.commands()

	fmt.Fprintln(w, "\nSpecial commands:")

	// Display registered commands with their descriptions
	for _, cmdName := range commands {
		if cmd, exists := r.cmds[cmdName]; exists {
			fmt.Fprintf(w, "  /%-12s\t- %s\n", cmdName, cmd.Description)
		} else {
			// Fallback for commands without descriptions (shouldn't happen)
			fmt.Fprintf(w, "  /%s - No description available\n", cmdName)
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
		input, err := r.rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err != nil {
			return "", err
		}

		// Trim whitespace from input
		input = strings.TrimSpace(input)
		
		// Check if it's a special command
		if isSpecialCommand(input) {
			// Process the input (handle special commands)
			err := r.handleSpecialCommand(input)
			if err != nil {
				return "", err
			}
		} else {
			// It's user input, return it
			return input, nil
		}
	}
}

func isSpecialCommand(input string) bool {
	return strings.HasPrefix(input, "/")
}

// handleSpecialCommand processes input and handles special commands
// Returns an error if command processing fails including io.EOF for exit
func (r *Reader) handleSpecialCommand(input string) error {
	// Remove the '/' prefix and convert to lowercase
	commandName := strings.ToLower(strings.TrimPrefix(input, "/"))

	// Look up the handler for this command
	if cmd, exists := r.cmds[commandName]; exists {
		return cmd.Handler(os.Stdout)
	}

	// Unknown command
	fmt.Printf("Unknown command: %s\nType '/help' to see available commands.\n", input)
	return nil
}
