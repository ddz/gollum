package main

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
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

// getTextEditorToolForModel returns the appropriate text editor tool configuration for the given model
func getTextEditorToolForModel(model anthropic.Model) anthropic.BetaToolUnionParam {
	modelStr := string(model)

	// Claude 4 models use text_editor_20250429
	if strings.Contains(modelStr, "claude-4") ||
		strings.Contains(modelStr, "claude-sonnet-4") ||
		strings.Contains(modelStr, "claude-opus-4") {
		return anthropic.BetaToolUnionParam{
			OfTextEditor20250429: &anthropic.BetaToolTextEditor20250429Param{
				Name: "str_replace_based_edit_tool",
			},
		}
	}

	// Claude 3.7 and Claude 3.5 models use text_editor_20250124
	if strings.Contains(modelStr, "claude-3-7") ||
		strings.Contains(modelStr, "claude-3.7") ||
		strings.Contains(modelStr, "claude-3-5") ||
		strings.Contains(modelStr, "claude-3.5") {
		return anthropic.BetaToolUnionParam{
			OfTextEditor20250124: &anthropic.BetaToolTextEditor20250124Param{
				Name: "str_replace_editor",
			},
		}
	}

	// Claude 3 and Claude 2 models use text_editor_20241022
	// This includes claude-3-*, claude-2-*, etc.
	return anthropic.BetaToolUnionParam{
		OfTextEditor20241022: &anthropic.BetaToolTextEditor20241022Param{
			Name: "str_replace_editor",
		},
	}
}

// getTextEditorToolName returns the tool name for the given model
func getTextEditorToolName(model anthropic.Model) string {
	modelStr := string(model)

	// Claude 4 models use str_replace_based_edit_tool
	if strings.Contains(modelStr, "claude-4") ||
		strings.Contains(modelStr, "claude-sonnet-4") ||
		strings.Contains(modelStr, "claude-opus-4") {
		return "str_replace_based_edit_tool"
	}

	// Claude 3.7, 3.5, 3, and 2 models use str_replace_editor
	return "str_replace_editor"
}
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
	fmt.Println("Claude 4 models (text_editor_20250429):")
	fmt.Println("  claude-sonnet-4-0, claude-4-sonnet")
	fmt.Println("  claude-sonnet-4-20250514, claude-4-sonnet-20250514")
	fmt.Println("  claude-opus-4-0, claude-4-opus")
	fmt.Println("  claude-opus-4-20250514, claude-4-opus-20250514")
	fmt.Println("\nClaude 3.7 models (text_editor_20250124):")
	fmt.Println("  claude-3-7-sonnet-latest, claude-3.7-sonnet-latest")
	fmt.Println("  claude-3-7-sonnet-20250219, claude-3.7-sonnet-20250219")
	fmt.Println("\nClaude 3.5 models (text_editor_20250124):")
	fmt.Println("  claude-3-5-sonnet-latest, claude-3.5-sonnet-latest (default)")
	fmt.Println("  claude-3-5-sonnet-20241022, claude-3.5-sonnet-20241022")
	fmt.Println("  claude-3-5-sonnet-20240620, claude-3.5-sonnet-20240620")
	fmt.Println("  claude-3-5-haiku-latest, claude-3.5-haiku-latest")
	fmt.Println("  claude-3-5-haiku-20241022, claude-3.5-haiku-20241022")
	fmt.Println("\nClaude 3 models (text_editor_20241022):")
	fmt.Println("  claude-3-opus-latest, claude-3-opus")
	fmt.Println("  claude-3-opus-20240229")
	fmt.Println("  claude-3-sonnet-20240229")
	fmt.Println("  claude-3-haiku-20240307")
	fmt.Println("\nClaude 2 models (text_editor_20241022):")
	fmt.Println("  claude-2.1")
	fmt.Println("  claude-2.0")
	fmt.Println("\nYou can also specify any model name directly (for future models).")
	fmt.Println("Text editor tool versions are automatically selected based on model compatibility.")
}

// toolProviders holds the specific tool implementations for tool use
type toolProviders struct {
	Bash       BashTool
	TextEditor TextEditorTool
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
		fmt.Fprintf(os.Stderr, "Anthropic Claude Agent with Local Bash and Built-in Text Editor\n\n")
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
		Bash:       NewSimpleBashTool(),
		TextEditor: NewSimpleTextEditorTool(),
	}

	// Get the appropriate text editor tool for the selected model
	textEditorTool := getTextEditorToolForModel(selectedModel)
	textEditorToolName := getTextEditorToolName(selectedModel)

	toolParams := []anthropic.BetaToolUnionParam{
		// Use the built-in Bash20250124 tool
		anthropic.BetaToolUnionParam{
			OfBashTool20250124: &anthropic.BetaToolBash20250124Param{
				Name: "bash",
			},
		},
		// Use the appropriate text editor tool for the model
		textEditorTool,
	}

	// Initialize conversation
	messages := []anthropic.BetaMessageParam{}

	// Create a scanner for user input
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Anthropic Claude Agent with Local Bash and Built-in Text Editor")
	fmt.Printf("Using model: %s\n", *modelName)
	fmt.Println("Commands are executed locally on your machine")
	fmt.Printf("Text editor tool: %s\n", textEditorToolName)
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
						} else if current.Name == textEditorToolName {
							fmt.Printf(
								"\n[Preparing to execute text editor command...]\n")
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
					if inToolUse && (current.Name == "bash" ||
						current.Name == textEditorToolName) {
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

	fmt.Println("\n[Executing tool commands...]")

	// Process each tool use
	for _, toolUse := range toolUseBlocks {
		if toolUse.Name == "bash" {
			toolUseResult := onBashToolUse(tools.Bash, toolUse)
			// Add tool result to messages
			results = append(results, toolUseResult)
		} else if toolUse.Name == "str_replace_editor" || toolUse.Name == "str_replace_based_edit_tool" {
			toolUseResult := onTextEditorToolUse(tools.TextEditor, toolUse)
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

func onTextEditorToolUse(textEditor TextEditorTool, toolUse toolUseInfo) anthropic.BetaContentBlockParamUnion {
	// Create tool result
	var toolResult anthropic.BetaContentBlockParamUnion

	// The built-in text editor tool has a different schema than our custom one
	// It primarily focuses on string replacement operations
	var input struct {
		// Built-in text editor fields
		Command   string `json:"command"`
		Path      string `json:"path"`
		OldStr    string `json:"old_str"`
		NewStr    string `json:"new_str"`
		ViewRange []int  `json:"view_range,omitempty"`

		// Legacy custom fields (for backward compatibility)
		Start      *int   `json:"start"`
		End        *int   `json:"end"`
		FileText   string `json:"file_text"`
		InsertLine *int   `json:"insert_line"`
		NewText    string `json:"new_text"`
	}
	err := json.Unmarshal(toolUse.Input, &input)
	if err != nil {
		fmt.Printf("\nError parsing text editor command: %v\n", err)
		toolResult = anthropic.NewBetaToolResultBlock(
			toolUse.ID,
			fmt.Sprintf("Error parsing command: %v", err),
			true, // isError
		)
		return toolResult
	}

	var output string
	var execErr error

	// Use the actual tool name from the tool use for logging
	toolName := toolUse.Name

	switch input.Command {
	case "view":
		fmt.Printf("\n[%s] Viewing: %s", toolName, input.Path)

		// Handle view_range from built-in tool
		var start, end *int
		if len(input.ViewRange) >= 1 {
			start = &input.ViewRange[0]
		}
		if len(input.ViewRange) >= 2 {
			end = &input.ViewRange[1]
		}
		// Fall back to legacy start/end fields
		if start == nil && input.Start != nil {
			start = input.Start
		}
		if end == nil && input.End != nil {
			end = input.End
		}

		if start != nil || end != nil {
			fmt.Printf(" (lines %v-%v)", start, end)
		}
		fmt.Println()

		rawOutput, execErr := textEditor.View(input.Path, start, end)
		if execErr == nil {
			// Add line numbers to the output
			output = addLineNumbers(rawOutput, start)
		}

	case "str_replace":
		fmt.Printf("\n[%s] String replace in: %s\n", toolName, input.Path)
		fmt.Printf("  Replacing: %q\n", input.OldStr)
		fmt.Printf("  With: %q\n", input.NewStr)

		execErr = textEditor.StringReplace(input.Path, input.OldStr, input.NewStr)
		if execErr == nil {
			output = "String replacement completed successfully"
		}

	case "create":
		fmt.Printf("\n[%s] Creating file: %s\n", toolName, input.Path)

		execErr = textEditor.Create(input.Path, input.FileText)
		if execErr == nil {
			output = fmt.Sprintf("File %s created successfully", input.Path)
		}

	case "insert":
		if input.InsertLine == nil {
			execErr = fmt.Errorf("insert_line is required for insert command")
		} else {
			fmt.Printf("\n[%s] Inserting text in: %s (after line %d)\n",
				toolName, input.Path, *input.InsertLine)

			execErr = textEditor.Insert(input.Path, *input.InsertLine, input.NewText)
			if execErr == nil {
				output = "Text insertion completed successfully"
			}
		}

	case "undo_edit":
		fmt.Printf("\n[%s] Undoing last edit in: %s\n", toolName, input.Path)

		execErr = textEditor.UndoEdit(input.Path)
		if execErr == nil {
			output = "Undo completed successfully"
		}

	default:
		execErr = fmt.Errorf("unknown text editor command: %s", input.Command)
	}

	if execErr != nil {
		fmt.Printf("Error: %s\n", execErr)
		toolResult = anthropic.NewBetaToolResultBlock(
			toolUse.ID,
			fmt.Sprintf("Error: %v", execErr),
			true, // isError
		)
	} else {
		toolResult = anthropic.NewBetaToolResultBlock(
			toolUse.ID,
			output,
			false, // isError
		)
	}

	return toolResult
}
