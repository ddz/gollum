package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

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

// executeBashCommand executes a bash command locally and returns the output
func executeBashCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nError output:\n" + stderr.String()
	}

	if err != nil {
		return output, fmt.Errorf("command failed: %v\n%s", err, output)
	}

	return output, nil
}

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set ANTHROPIC_API_KEY environment variable")
		os.Exit(1)
	}

	// Create client
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	// Use the built-in Bash20250124 tool
	bashTool := anthropic.BetaToolUnionParam{
		OfBashTool20250124: &anthropic.BetaToolBash20250124Param{
			Name: "bash",
		},
	}

	// Initialize conversation
	messages := []anthropic.BetaMessageParam{}

	// Create a scanner for user input
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Anthropic Agent with Local Bash Execution")
	fmt.Println("Commands are executed locally on your machine")
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
			stream := client.Beta.Messages.NewStreaming(ctx,
				anthropic.BetaMessageNewParams{
					Model:     anthropic.ModelClaude3_5Sonnet20241022,
					MaxTokens: 1024,
					Messages:  messages,
					Tools:     []anthropic.BetaToolUnionParam{bashTool},
					Betas: []anthropic.AnthropicBeta{
						anthropic.AnthropicBetaComputerUse2025_01_24,
					},
				})

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
				fmt.Println("\n[Executing bash commands locally...]")

				// Process each tool use
				for _, toolUse := range toolUseBlocks {
					if toolUse.Name == "bash" {
						// Parse the command from the input
						var input struct {
							Command string `json:"command"`
						}
						err := json.Unmarshal(toolUse.Input, &input)
						if err != nil {
							fmt.Printf(
								"\nError parsing bash command: %v\n", err)
							continue
						}

						fmt.Printf("\n$ %s\n", input.Command)

						// Execute the command locally
						output, err := executeBashCommand(input.Command)

						// Create tool result
						var toolResultContent anthropic.BetaContentBlockParamUnion

						if err != nil {
							fmt.Printf("Error: %v\n", err)
							toolResultContent = anthropic.NewBetaToolResultBlock(
								toolUse.ID,
								fmt.Sprintf("Error executing command: %v\nOutput: %s",
									err, output),
								true, // isError
							)
						} else {
							fmt.Print(output)
							if !strings.HasSuffix(output, "\n") {
								fmt.Println()
							}
							toolResultContent = anthropic.NewBetaToolResultBlock(
								toolUse.ID,
								output,
								false, // isError
							)
						}

						// Add tool result to messages
						messages = append(messages,
							anthropic.NewBetaUserMessage(toolResultContent))
					}
				}

				// Continue the conversation with tool results
				continue
			}

			// No tool use, break the loop
			break
		}

		fmt.Println()
	}
}
