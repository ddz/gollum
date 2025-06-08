package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// systemPrompt defines the system prompt for the assistant.
// Modify this string to customize the agent's behavior and personality.
// Leave empty ("") to use Claude's default behavior.
//
// Examples:
// const systemPrompt = "You are a helpful coding assistant. Be concise and practical."
// const systemPrompt = "You are an expert Linux system administrator."
// const systemPrompt = "You are a helpful AI assistant that specializes in data analysis. Always explain your reasoning step by step."
const systemPrompt = ""

// getModelFromString converts a model string to the appropriate anthropic.Model constant
func getModelFromString(modelStr string) anthropic.Model {
	switch modelStr {
	// Claude 4 models
	case "claude-sonnet-4-0", "claude-4-sonnet":
		return anthropic.ModelClaudeSonnet4_0
	case "claude-sonnet-4-20250514", "claude-4-sonnet-20250514":
		return anthropic.ModelClaude4Sonnet20250514
	case "claude-opus-4-0", "claude-4-opus":
		return anthropic.ModelClaudeOpus4_0
	case "claude-opus-4-20250514", "claude-4-opus-20250514":
		return anthropic.ModelClaude4Opus20250514

	// Claude 3.7 models
	case "claude-3-7-sonnet-latest", "claude-3.7-sonnet-latest":
		return anthropic.ModelClaude3_7SonnetLatest
	case "claude-3-7-sonnet-20250219", "claude-3.7-sonnet-20250219":
		return anthropic.ModelClaude3_7Sonnet20250219

	// Claude 3.5 models
	case "claude-3-5-sonnet-latest", "claude-3.5-sonnet-latest":
		return anthropic.ModelClaude3_5SonnetLatest
	case "claude-3-5-sonnet-20241022", "claude-3.5-sonnet-20241022":
		return anthropic.ModelClaude3_5Sonnet20241022
	case "claude-3-5-sonnet-20240620", "claude-3.5-sonnet-20240620":
		return anthropic.ModelClaude_3_5_Sonnet_20240620
	case "claude-3-5-haiku-latest", "claude-3.5-haiku-latest":
		return anthropic.ModelClaude3_5HaikuLatest
	case "claude-3-5-haiku-20241022", "claude-3.5-haiku-20241022":
		return anthropic.ModelClaude3_5Haiku20241022

	// Claude 3 models
	case "claude-3-opus-latest", "claude-3-opus":
		return anthropic.ModelClaude3OpusLatest
	case "claude-3-opus-20240229":
		return anthropic.ModelClaude_3_Opus_20240229
	case "claude-3-sonnet-20240229":
		return anthropic.ModelClaude_3_Sonnet_20240229
	case "claude-3-haiku-20240307":
		return anthropic.ModelClaude_3_Haiku_20240307

	// Claude 2 models
	case "claude-2.1":
		return anthropic.ModelClaude_2_1
	case "claude-2.0":
		return anthropic.ModelClaude_2_0

	default:
		// Return the raw string as a Model - this allows for future models
		// that may not be in our mapping yet
		return anthropic.Model(modelStr)
	}
}

// printAvailableModels prints the list of supported model names
func printAvailableModels() {
	fmt.Println("\nSupported model names:")
	fmt.Println("Claude 4 models:")
	fmt.Println("  claude-sonnet-4-0, claude-4-sonnet")
	fmt.Println("  claude-sonnet-4-20250514, claude-4-sonnet-20250514")
	fmt.Println("  claude-opus-4-0, claude-4-opus")
	fmt.Println("  claude-opus-4-20250514, claude-4-opus-20250514")
	fmt.Println("\nClaude 3.7 models:")
	fmt.Println("  claude-3-7-sonnet-latest, claude-3.7-sonnet-latest")
	fmt.Println("  claude-3-7-sonnet-20250219, claude-3.7-sonnet-20250219")
	fmt.Println("\nClaude 3.5 models:")
	fmt.Println("  claude-3-5-sonnet-latest, claude-3.5-sonnet-latest (default)")
	fmt.Println("  claude-3-5-sonnet-20241022, claude-3.5-sonnet-20241022")
	fmt.Println("  claude-3-5-sonnet-20240620, claude-3.5-sonnet-20240620")
	fmt.Println("  claude-3-5-haiku-latest, claude-3.5-haiku-latest")
	fmt.Println("  claude-3-5-haiku-20241022, claude-3.5-haiku-20241022")
	fmt.Println("\nClaude 3 models:")
	fmt.Println("  claude-3-opus-latest, claude-3-opus")
	fmt.Println("  claude-3-opus-20240229")
	fmt.Println("  claude-3-sonnet-20240229")
	fmt.Println("  claude-3-haiku-20240307")
	fmt.Println("\nClaude 2 models:")
	fmt.Println("  claude-2.1")
	fmt.Println("  claude-2.0")
	fmt.Println("\nYou can also specify any model name directly (for future models).")
}

// toolProviders holds the specific tool implementations for tool use
type toolProviders struct {
	Bash BashTool
}

// toolUseInfo holds information about a tool use block
type toolUseInfo struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// currentToolUse holds the state of a tool use being accumulated
type currentToolUse struct {
	ID    string
	Name  string
	Input string
}

func main() {
	// Define command-line flags
	var (
		modelName  = flag.String("model", "claude-3-5-sonnet-latest", "Model to use (e.g., claude-sonnet-4-0, claude-3-5-sonnet-latest)")
		listModels = flag.Bool("list-models", false, "List available model names and exit")
		help       = flag.Bool("help", false, "Show help message")
	)

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Anthropic Claude Agent with Local Bash Execution\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  ANTHROPIC_API_KEY    Anthropic API key (required)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                                   # Use default Claude 3.5 Sonnet model\n", os.Args[0])
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

	// Convert model name to appropriate constant
	selectedModel := getModelFromString(*modelName)

	// Create client
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	// Instantiate tool providers
	tools := &toolProviders{
		Bash: NewSimpleBashTool(),
	}

	toolParams := []anthropic.BetaToolUnionParam{
		// Use the built-in Bash20250124 tool
		anthropic.BetaToolUnionParam{
			OfBashTool20250124: &anthropic.BetaToolBash20250124Param{
				Name: "bash",
			},
		},
	}

	// Initialize conversation
	messages := []anthropic.BetaMessageParam{}

	// Create a scanner for user input
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Anthropic Claude Agent with Local Bash Execution")
	fmt.Printf("Using model: %s\n", *modelName)
	fmt.Println("Commands are executed locally on your machine")
	if systemPrompt != "" {
		fmt.Printf("System prompt: %s\n", systemPrompt)
	}
	fmt.Println("Type 'exit' to quit")
	fmt.Println("---------------------------------------------------")

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}

		userInput := scanner.Text()
		if strings.ToLower(userInput) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Add user message
		messages = append(messages,
			anthropic.NewBetaUserMessage(
				anthropic.NewBetaTextBlock(userInput)))

		// Loop to handle potential tool use
		for {
			// Create streaming request with beta for bash tool
			ctx := context.Background()

			// Build the message parameters
			// Using the selected model from command line
			params := anthropic.BetaMessageNewParams{
				Model:     selectedModel,
				MaxTokens: 1024,
				Messages:  messages,
				Tools:     toolParams,
				Betas: []anthropic.AnthropicBeta{
					anthropic.AnthropicBetaComputerUse2025_01_24,
				},
			}

			// Add system prompt if provided
			if systemPrompt != "" {
				params.System = []anthropic.BetaTextBlockParam{
					{
						Text: systemPrompt,
						Type: "text",
					},
				}
			}

			stream := client.Beta.Messages.NewStreaming(ctx, params)

			fmt.Print("\nAssistant: ")

			message := anthropic.BetaMessage{}
			toolUseBlocks := []toolUseInfo{}

			current := currentToolUse{}
			inToolUse := false

			// Process the stream
			for stream.Next() {
				event := stream.Current()
				err := message.Accumulate(event)
				if err != nil {
					fmt.Printf("\nError accumulating message: %v\n", err)
					break
				}

				// Handle different event types
				switch event := event.AsAny().(type) {
				case anthropic.BetaRawContentBlockStartEvent:
					// Check what type of content block this is
					if event.ContentBlock.Type == "tool_use" {
						inToolUse = true
						current.ID = event.ContentBlock.ID
						current.Name = event.ContentBlock.Name
						current.Input = ""
						if current.Name == "bash" {
							fmt.Printf(
								"\n[Preparing to execute bash command locally...]\n")
						}
					}

				case anthropic.BetaRawContentBlockDeltaEvent:
					// Handle text deltas
					if delta := event.Delta; delta.Type == "text_delta" {
						fmt.Print(delta.Text)
					} else if delta.Type == "input_json_delta" && inToolUse {
						// Accumulate tool input
						current.Input += delta.PartialJSON
					}

				case anthropic.BetaRawContentBlockStopEvent:
					if inToolUse && current.Name == "bash" {
						// Parse and store the tool use
						toolUseBlocks = append(toolUseBlocks, toolUseInfo{
							ID:    current.ID,
							Name:  current.Name,
							Input: json.RawMessage(current.Input),
						})
						inToolUse = false
					}
				}
			}

			// Check for stream errors
			if err := stream.Err(); err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				break
			}

			// Add assistant message to history
			messages = append(messages, message.ToParam())

			// If there were bash tool uses, execute them and continue
			if len(toolUseBlocks) > 0 {
				toolUseResults := onToolUse(tools, toolUseBlocks)
				messages = append(messages,
					anthropic.NewBetaUserMessage(toolUseResults...))

				// Continue the conversation with tool results
				continue
			}

			// No tool use, break the loop
			break
		}

		fmt.Println()
	}
}

func onToolUse(tools *toolProviders, toolUseBlocks []toolUseInfo) []anthropic.BetaContentBlockParamUnion {
	var results []anthropic.BetaContentBlockParamUnion

	fmt.Println("\n[Executing bash commands locally...]")

	// Process each tool use
	for _, toolUse := range toolUseBlocks {
		if toolUse.Name == "bash" {
			toolUseResult := onBashToolUse(tools.Bash, toolUse)
			// Add tool result to messages
			results = append(results, toolUseResult)
		}
	}

	return results
}

func onBashToolUse(bashTool BashTool, toolUse toolUseInfo) anthropic.BetaContentBlockParamUnion {
	// Create tool result
	var toolResult anthropic.BetaContentBlockParamUnion

	// Parse the command from the input
	var input struct {
		Command string `json:"command"`
		Restart bool   `json:"restart"`
	}
	err := json.Unmarshal(toolUse.Input, &input)
	if err != nil {
		fmt.Printf(
			"\nError parsing bash command: %v\n", err)
		return toolResult
	}

	if input.Restart {
		fmt.Printf("\n Restarting bash session...")
		message, err := bashTool.Restart()

		// No actual need to restart we don't support sessions yet
		toolResult = anthropic.NewBetaToolResultBlock(
			toolUse.ID,
			message,
			err != nil, // isError
		)

		return toolResult
	}

	fmt.Printf("\n$ %s\n", input.Command)

	// Execute the command locally
	stdout, stderr, err := bashTool.ExecuteCommand(input.Command)
	output := fmt.Sprintf("<stdout>%s</stdout><stderr>%s</stderr>",
		stdout, stderr)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	toolResult = anthropic.NewBetaToolResultBlock(
		toolUse.ID,
		output,
		err != nil, // isError
	)

	return toolResult
}
