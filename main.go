package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

// systemPrompt defines the system prompt for the assistant.
// The content is embedded from prompt.txt at compile time.
//
//go:embed prompt.txt
var systemPrompt string

// addLineNumbers adds line numbers to the beginning of each line in the text
func addLineNumbers(text string, startLine *int) string {
	if text == "" {
		return text
	}

	lines := strings.Split(text, "\n")
	var result strings.Builder

	// Determine starting line number
	start := 1
	if startLine != nil {
		start = *startLine
	}

	for i, line := range lines {
		lineNum := start + i
		// Don't add line number to the last empty line if the text ends with \n
		if i == len(lines)-1 && line == "" {
			break
		}
		result.WriteString(fmt.Sprintf("%d: %s\n", lineNum, line))
	}

	// Remove the trailing newline if we added one
	output := result.String()
	if strings.HasSuffix(output, "\n") {
		output = output[:len(output)-1]
	}

	return output
}

// toolProviders holds the specific tool implementations for tool use
type toolProviders struct {
	Bash       BashTool
	TextEditor TextEditorTool
}

// completer provides auto-completion for common commands
var completer = readline.NewPrefixCompleter(
	readline.PcItem("/exit"),
	readline.PcItem("/quit"),
	readline.PcItem("/help"),
	readline.PcItem("/clear"),
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

func main() {
	// Define command-line flags
	var (
		modelName  = flag.String("model", "claude-4-sonnet", "Model to use (e.g., claude-sonnet-4-0, claude-3-5-sonnet-latest)")
		listModels = flag.Bool("list-models", false, "List available model names and exit")
		help       = flag.Bool("help", false, "Show help message")
	)

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Anthropic Claude Agent with Local Bash and Built-in Text Editor\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  ANTHROPIC_API_KEY    Anthropic API key (required)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                                   # Use default Claude 4 Sonnet model\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -model claude-sonnet-4-0          # Use Claude 4 Sonnet\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -model claude-4-opus              # Use Claude 4 Opus\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-models                      # Show available models\n", os.Args[0])
	}

	flag.Parse()

	// Handle help flag
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Handle list-models flag
	if *listModels {
		printAvailableModels()
		os.Exit(0)
	}

	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set ANTHROPIC_API_KEY environment variable")
		os.Exit(1)
	}

	// Instantiate tool providers
	tools := &toolProviders{
		Bash:       NewSimpleBashTool(),
		TextEditor: NewSimpleTextEditorTool(),
	}

	// Create Anthropic client
	client := NewAnthropicClient(apiKey, *modelName, systemPrompt, tools)

	// Initialize conversation
	conversation := NewConversation()

	// Create readline instance with history and editing support
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "You: ",
		HistoryFile:     ".gollum_history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		fmt.Printf("Error creating readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	fmt.Println("Anthropic Claude Agent with Local Bash and Built-in Text Editor")
	fmt.Printf("Using model: %s\n", *modelName)
	fmt.Println("Commands are executed locally on your machine")
	fmt.Printf("Text editor tool: %s\n", client.TextEditorToolName)
	fmt.Println("History is saved to .gollum_history")
	fmt.Println("Use Ctrl+R for reverse history search, Ctrl+C to interrupt")
	if systemPrompt != "" {
		fmt.Printf("System prompt: %s\n", systemPrompt)
	}
	fmt.Println("Type '/exit' to quit, '/help' for special commands")
	fmt.Println("---------------------------------------------------")

	for {
		userInput, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(userInput) == 0 {
				fmt.Println("\nGoodbye!")
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			fmt.Println("\nGoodbye!")
			break
		} else if err != nil {
			fmt.Printf("\nError reading input: %v\n", err)
			break
		}

		userInput = strings.TrimSpace(userInput)
		if userInput == "" {
			continue
		}
		
		// Handle special commands
		switch strings.ToLower(userInput) {
		case "/exit", "/quit":
			fmt.Println("Goodbye!")
			return
		case "/clear":
			readline.ClearScreen(rl)
			continue
		case "/help":
			fmt.Println("\nSpecial commands:")
			fmt.Println("  /exit, /quit - Exit the application")  
			fmt.Println("  /clear       - Clear the screen")
			fmt.Println("  /help        - Show this help")
			fmt.Println("\nKeyboard shortcuts:")
			fmt.Println("  Ctrl+R       - Reverse history search")
			fmt.Println("  Ctrl+C       - Interrupt current input")
			fmt.Println("  Ctrl+D       - Exit (EOF)")
			fmt.Println("  Up/Down      - Navigate history")
			continue
		}

		// Add user message
		conversation.AddUserMessage(userInput)

		// Loop to handle potential tool use
		for {
			ctx := context.Background()

			// Send message to Anthropic and get response
			toolUseBlocks, err := client.SendMessage(ctx, conversation)
			if err != nil {
				fmt.Printf("\nError: %v\n", err)
				break
			}

			// If there were tool uses, execute them and continue
			if len(toolUseBlocks) > 0 {
				client.ExecuteTools(toolUseBlocks, conversation)

				// Continue the conversation with tool results
				continue
			}

			// No tool use, break the loop
			break
		}

		fmt.Println()
	}
}


