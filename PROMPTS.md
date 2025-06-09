# LLM Prompts Used to Build This Project

This document contains all the LLM prompts identified from the git commit history that were used to generate code and features for this project. The project appears to be built incrementally using various LLM assistants (primarily Goose with Claude models).

## System Prompt

From `prompt.txt`:

```
You are Gollum, a monster from Middle-earth who is sworn to help the user.

You have access to bash commands and text editing capabilities. Use
them effectively to help achieve what the user requests.
```

## Initial Project Setup

### Commit: 8b11ef7 - Add generated simple agent with local bash tool use
**Generated using Goose w/ claude-opus-4-20250514**

Prompts used:
- Please write a simple agent using the Anthropic SDK for Go that can use the Bash tool
- This created a generic bash tool, instead use the built-in Bash tool Bash20250124
- Make sure that the agent can handle tool use for the bash tool in the response
- If the ContentBlock.Type is "tool_use", check if the tool name is "bash". If it is bash, then execute the command locally and return the tool result to Claude.

## System Prompt Customization

### Commit: 953064b - Allow compile-time customization of system prompt
**Generated using Goose w/ claude-sonnet-4-20250514**

Prompts used:
- Modify the simple agent in main.go to allow specifying the system prompt
- Remove the command-line option, but keep the ability to specify a system prompt in the source code

### Commit: d2ac329 - Embed prompt.txt as system prompt and give Gollum a personality
**Generated using Goose w/ claude-sonnet-4-20250514**

Prompts used:
- Can you use the go embed package to embed the contents of prompt.txt in the const systemPrompt in main.go?

## Model Configuration

### Commit: d09f9ea - Add command-line option to specify a specific model
**Generated using Goose w/ claude-sonnet-4-20250514**

Prompts used:
- Let's upgrade the agent in this project to using Claude 4 with a default model of claude-sonnet-4.
- Add a command-line option to specify an alternate model to use

### Commit: b206dfb - Change default model to latest Claude 3.5 Sonnet
**Generated using Goose w/ claude-sonnet-4-20250514**

Prompts used:
- Can you change the default model in main.go to be Claude Sonnet 3.5 instead of 4.0?

### Commit: 507b2b5 - Remove support for models that don't support bash and text editor tools
**Generated using Goose w/ claude-sonnet-4-20250514**

Prompts used:
- Models earlier than Claude 3.5 don't support the text editor or bash tools. Can you remove support for models before Claude 3.5 as well as removing support for Claude 3.5 Haiku?

### Commit: 9887338 - Change default model to Claude 4 Sonnet and fix pointer printing bug
**Generated using Gollum w/ claude-4-sonnet**

Prompts used:
- Please help me change the default model used by this LLM agent to claude-4-sonnet in ./main.go
- The current directory is an agent written in Go. It has a bug when using the text editor tool that prints out the value of pointers instead of the value they point to. Here is an example: Viewing: main.go (lines 0xc00028cf70-0xc00028cf78)

## Testing Infrastructure

### Commit: 8423930 - Add unit tests for BashTool and SimpleBashTool
**Generated using Goose w/ claude-sonnet-4-20250514**

Prompts used:
- Can you write unit tests for bash_tool.go?
- Those tests are specific to the specific SimpleBashTool implementation, can you refactor the tests to focus on testing the interfaces in BashTool and taking an instance of BashTool as a parameter? Then you can run those as subtests of a top-level test of SimpleBashTool. This will make it easier to reuse that test code as there are more implementations of the BashTool interface.

## Text Editor Tool Implementation

### Commit: b18f572 - Add text editor tool support
**Generated using Goose w/ claude-sonnet-4-20250514**

Prompts used:
- I've written an interface for an LLM agent's text editor tool in text_editor_tool.go. Can you add a simple implementation of it called SimpleTextEditorTool with unit tests? Make sure that the unit tests include subtests that focus on testing the interface methods so that they are reusable across implementations of the TextEditorTool interface.
- Now let's add this text editor tool to the main agent loop in main.go. Make sure to add the TextEditorTool interface to a new provider in the toolProviders struct to make it easy to use other providers of that interface in the future.

### Commit: e50b784 - Add line numbers to the output of the view text editor tool
**Generated using Goose w/ claude-sonnet-4-20250514**

Prompts used:
- Claude prefers that the text editor view tool result includes file contents with line numbers prepended to each line (e.g., "1: def is_prime(n):"). Can you add these line numbers to the output received from textEditor.View in the onTextEditorToolUse function in main.go?

## Bug Fixes and Runtime Issues

### Commit: ff78f1e - Fix runtime errors
**Generated using Goose w/ claude-sonnet-4-20250514**

Prompts used:
- My project is a simple LLM agent using the Claude SDK. When I run it with Claude 3.5, I get the following error:
  ```
  Stream error: POST "https://api.anthropic.com/v1/messages?beta=true":
    400 Bad Request {
      "type":"error",
      "error": {
        "type":"invalid_request_error",
        "message":"tools.1.custom.input_schema: Field required"
      }
    }
  ```
  Can you help me fix this?
- The BetaToolParam is for a custom tool, but the text editor tool is a built-in tool. Can you use the BetaToolTextEditor20250429Param structure instead of BetaToolParam for the "text_editor" tool?
- The default Claude 3.5 model does not support text_editor_20250429, only Claude 4 does. Can you make sure to use the text_editor_20241022 tool when using Claude 3.5, text_editor_20250124 for Claude 3.7, and text_editor_20250429 for Claude 4.0? This will require using the correct param structures like BetaToolTextEditor20241022Param, BetaToolTextEditor20250124Param, and BetaToolTextEditor20250429Param depending on the model version.
- It looks like the documentation may be inaccurate and that Claude 3.5 requires the "text_editor_20250124" tool instead of "text_editor_20241022". Can you update this and the associated param structure when using Claude 3.5?

## User Experience Improvements

### Commit: a3178c8 - Add readline for better UX and special command handling
**Generated using Gollum (!!!) w/ claude-sonnet-4-20250514**

Prompts used:
- Please add the go library github.com/chzyer/readline and support line editing and history to reading input from the user to the agent implemented in main.go
- Now make a small change to make all special commands start with a '/'
- We don't need backwards compatibility with commands without a leading slash, go ahead and remove that.
- Now let's add the history file to .gitignore

## Documentation

### Commit: 7af3760 - Add Goose-generated README.md using claude-sonnet-4-20250514
**Generated using Goose w/ claude-sonnet-4-20250514**

Note: Made some small human edits to wrap lines and correct some facts.

## Manual Development Notes

Several commits were noted as manual work without LLM assistance:
- **3432721** - Add hand-written interface for TextEditorTool (Based on Anthropic's documentation)
- **21af078** - Manually refactor bash tool into interface and simple implementation
- **c760817** - Respond to bash session restart requests, but do nothing
- **b3359fb** - Do some manual refactoring to reduce indention level of agent loop
- **d9f2ea0** - Combine TextEditorTool interface and SimpleTextEditorTool in one file
- **1d78169** - Small update to prompt
- **640c46e** - Move .goosehints to CODING-STYLE.md
- **729aaa8** - Formatting
- **1055fe3** - Add initial local goosehints
- **13883de** - Add manually created go.mod with latest Anthropic Go SDK
- **566a529** - Initial commit: add .gitignore

## Summary

This project was built using a combination of:
- **Primary LLM Tool**: Goose (Block's open-source AI agent)
- **Primary Models**: Claude Sonnet 4 (claude-sonnet-4-20250514), Claude 4 Sonnet (claude-4-sonnet), and Claude Opus 4 (claude-opus-4-20250514)
- **Development Approach**: Incremental development with specific, focused prompts for individual features
- **Key Areas**: Tool integration (bash and text editor), model configuration, testing, UX improvements, and bug fixes

The development shows a pattern of iterative improvement, with many prompts focused on specific technical challenges like tool parameter structures, model compatibility, and user interface enhancements.