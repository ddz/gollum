package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Conversation represents the conversation history
type Conversation struct {
	messages []anthropic.BetaMessageParam
}

// NewConversation creates a new conversation
func NewConversation() *Conversation {
	return &Conversation{
		messages: []anthropic.BetaMessageParam{},
	}
}

// AddUserMessage adds a user message to the conversation history
func (c *Conversation) AddUserMessage(content string) {
	c.messages = append(c.messages,
		anthropic.NewBetaUserMessage(
			anthropic.NewBetaTextBlock(content)))
}

// AddAssistantMessage adds an assistant message to the conversation history
func (c *Conversation) AddAssistantMessage(message anthropic.BetaMessage) {
	c.messages = append(c.messages, message.ToParam())
}

// AddToolResults adds tool results to the conversation history
func (c *Conversation) AddToolResults(results []anthropic.BetaContentBlockParamUnion) {
	c.messages = append(c.messages,
		anthropic.NewBetaUserMessage(results...))
}

// AnthropicClient wraps the Anthropic SDK client and provides high-level methods
type AnthropicClient struct {
	client               *anthropic.Client
	model                anthropic.Model
	TextEditorToolName   string
	systemPrompt         string
	tools                *toolProviders
}

// NewAnthropicClient creates a new Anthropic client with the specified configuration
func NewAnthropicClient(apiKey, modelName, systemPrompt string, tools *toolProviders) *AnthropicClient {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	model := getModelFromString(modelName)
	textEditorToolName := getTextEditorToolName(model)

	return &AnthropicClient{
		client:               &client,
		model:                model,
		TextEditorToolName:   textEditorToolName,
		systemPrompt:         systemPrompt,
		tools:                tools,
	}
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

	// Claude 3.7 and Claude 3.5 Sonnet models use text_editor_20250124
	if strings.Contains(modelStr, "claude-3-7") ||
		strings.Contains(modelStr, "claude-3.7") ||
		(strings.Contains(modelStr, "claude-3-5") && strings.Contains(modelStr, "sonnet")) ||
		(strings.Contains(modelStr, "claude-3.5") && strings.Contains(modelStr, "sonnet")) {
		return anthropic.BetaToolUnionParam{
			OfTextEditor20250124: &anthropic.BetaToolTextEditor20250124Param{
				Name: "str_replace_editor",
			},
		}
	}

	// Fallback for any unknown models - use text_editor_20250124 as default
	return anthropic.BetaToolUnionParam{
		OfTextEditor20250124: &anthropic.BetaToolTextEditor20250124Param{
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

	// Claude 3.7 and 3.5 Sonnet models use str_replace_editor
	return "str_replace_editor"
}

// getModelFromString converts a model string to an anthropic.Model
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

	// Claude 3.5 Sonnet models (Haiku models removed)
	case "claude-3-5-sonnet-latest", "claude-3.5-sonnet-latest":
		return anthropic.ModelClaude3_5SonnetLatest
	case "claude-3-5-sonnet-20241022", "claude-3.5-sonnet-20241022":
		return anthropic.ModelClaude3_5Sonnet20241022
	case "claude-3-5-sonnet-20240620", "claude-3.5-sonnet-20240620":
		return anthropic.ModelClaude_3_5_Sonnet_20240620

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
	fmt.Println("  claude-sonnet-4-0, claude-4-sonnet (default)")
	fmt.Println("  claude-sonnet-4-20250514, claude-4-sonnet-20250514")
	fmt.Println("  claude-opus-4-0, claude-4-opus")
	fmt.Println("  claude-opus-4-20250514, claude-4-opus-20250514")
	fmt.Println("\nClaude 3.7 models (text_editor_20250124):")
	fmt.Println("  claude-3-7-sonnet-latest, claude-3.7-sonnet-latest")
	fmt.Println("  claude-3-7-sonnet-20250219, claude-3.7-sonnet-20250219")
	fmt.Println("\nClaude 3.5 Sonnet models (text_editor_20250124):")
	fmt.Println("  claude-3-5-sonnet-latest, claude-3.5-sonnet-latest")
	fmt.Println("  claude-3-5-sonnet-20241022, claude-3.5-sonnet-20241022")
	fmt.Println("  claude-3-5-sonnet-20240620, claude-3.5-sonnet-20240620")
	fmt.Println("\nYou can also specify any model name directly (for future models).")
	fmt.Println("Text editor tool versions are automatically selected based on model compatibility.")
	fmt.Println("\nNote: Only models Claude 3.5 Sonnet and later are supported due to")
	fmt.Println("text editor and bash tool requirements.")
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

// SendMessage sends a message to the Anthropic API and handles the streaming response
func (ac *AnthropicClient) SendMessage(ctx context.Context, conversation *Conversation) ([]toolUseInfo, error) {
	// Get the appropriate text editor tool for the selected model
	textEditorTool := getTextEditorToolForModel(ac.model)

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

	// Build the message parameters
	params := anthropic.BetaMessageNewParams{
		Model:     ac.model,
		MaxTokens: 1024,
		Messages:  conversation.messages,
		Tools:     toolParams,
		Betas: []anthropic.AnthropicBeta{
			anthropic.AnthropicBetaComputerUse2025_01_24,
		},
	}

	// Add system prompt if provided
	if ac.systemPrompt != "" {
		params.System = []anthropic.BetaTextBlockParam{
			{
				Text: ac.systemPrompt,
				Type: "text",
			},
		}
	}

	stream := ac.client.Beta.Messages.NewStreaming(ctx, params)

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
			return toolUseBlocks, fmt.Errorf("error accumulating message: %v", err)
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
					fmt.Printf("\n[Preparing to execute bash command locally...]\n")
				} else if current.Name == ac.TextEditorToolName {
					fmt.Printf("\n[Preparing to execute text editor command...]\n")
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
			if inToolUse && (current.Name == "bash" || current.Name == ac.TextEditorToolName) {
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
		return toolUseBlocks, fmt.Errorf("stream error: %v", err)
	}

	// Add assistant message to conversation
	conversation.AddAssistantMessage(message)

	return toolUseBlocks, nil
}

// ExecuteTools executes the provided tool use blocks and adds results to conversation
func (ac *AnthropicClient) ExecuteTools(toolUseBlocks []toolUseInfo, conversation *Conversation) {
	var results []anthropic.BetaContentBlockParamUnion

	fmt.Println("\n[Executing tool commands...]")

	// Process each tool use
	for _, toolUse := range toolUseBlocks {
		if toolUse.Name == "bash" {
			toolUseResult := ac.onBashToolUse(toolUse)
			results = append(results, toolUseResult)
		} else if toolUse.Name == "str_replace_editor" || toolUse.Name == "str_replace_based_edit_tool" {
			toolUseResult := ac.onTextEditorToolUse(toolUse)
			results = append(results, toolUseResult)
		}
	}

	// Add tool results to conversation
	conversation.AddToolResults(results)
}

// onBashToolUse handles bash tool execution
func (ac *AnthropicClient) onBashToolUse(toolUse toolUseInfo) anthropic.BetaContentBlockParamUnion {
	// Create tool result
	var toolResult anthropic.BetaContentBlockParamUnion

	// Parse the command from the input
	var input struct {
		Command string `json:"command"`
		Restart bool   `json:"restart"`
	}
	err := json.Unmarshal(toolUse.Input, &input)
	if err != nil {
		fmt.Printf("\nError parsing bash command: %v\n", err)
		return anthropic.NewBetaToolResultBlock(
			toolUse.ID,
			fmt.Sprintf("Error parsing command: %v", err),
			true, // isError
		)
	}

	if input.Restart {
		fmt.Printf("\n Restarting bash session...")
		message, err := ac.tools.Bash.Restart()

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
	stdout, stderr, err := ac.tools.Bash.ExecuteCommand(input.Command)
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

// onTextEditorToolUse handles text editor tool execution
func (ac *AnthropicClient) onTextEditorToolUse(toolUse toolUseInfo) anthropic.BetaContentBlockParamUnion {
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
			startVal := "start"
			endVal := "end"
			if start != nil {
				startVal = fmt.Sprintf("%d", *start)
			}
			if end != nil {
				endVal = fmt.Sprintf("%d", *end)
			}
			fmt.Printf(" (lines %s-%s)", startVal, endVal)
		}
		fmt.Println()

		rawOutput, execErr := ac.tools.TextEditor.View(input.Path, start, end)
		if execErr == nil {
			// Add line numbers to the output
			output = addLineNumbers(rawOutput, start)
		}

	case "str_replace":
		fmt.Printf("\n[%s] String replace in: %s\n", toolName, input.Path)
		fmt.Printf("  Replacing: %q\n", input.OldStr)
		fmt.Printf("  With: %q\n", input.NewStr)

		execErr = ac.tools.TextEditor.StringReplace(input.Path, input.OldStr, input.NewStr)
		if execErr == nil {
			output = "String replacement completed successfully"
		}

	case "create":
		fmt.Printf("\n[%s] Creating file: %s\n", toolName, input.Path)

		execErr = ac.tools.TextEditor.Create(input.Path, input.FileText)
		if execErr == nil {
			output = fmt.Sprintf("File %s created successfully", input.Path)
		}

	case "insert":
		if input.InsertLine == nil {
			execErr = fmt.Errorf("insert_line is required for insert command")
		} else {
			fmt.Printf("\n[%s] Inserting text in: %s (after line %d)\n",
				toolName, input.Path, *input.InsertLine)

			execErr = ac.tools.TextEditor.Insert(input.Path, *input.InsertLine, input.NewText)
			if execErr == nil {
				output = "Text insertion completed successfully"
			}
		}

	case "undo_edit":
		fmt.Printf("\n[%s] Undoing last edit in: %s\n", toolName, input.Path)

		execErr = ac.tools.TextEditor.UndoEdit(input.Path)
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
