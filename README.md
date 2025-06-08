# Gollum

*"We likes to helps, yes precious! We knows the bash and the texts, yesss!"*

Gollum is a simple LLM agent written in Go, inspired by and written
using Block's open-source [Goose](https://github.com/block/goose)
AI agent. This precious little agent has been given the personality
of Gollum from "The Lord of the Rings" and serves as a helpful
command-line companion for developers.

## About the Name

The name "Gollum" is a clever portmanteau of:
- **Go** - the programming language this agent is written in
- **LLM** - Large Language Model

When pronounced, "GoLLM" sounds like "Gollum" - making it both
technically descriptive and thematically appropriate for this
ring-obsessed, cave-dwelling assistant who speaks in riddles and helps
with your coding tasks.

## Features

- **Gollum Personality**: Responds with the distinctive speech
  patterns and mannerisms of Gollum from Tolkien's Middle-earth
- **Local Command Execution**: Execute bash commands directly on your
  local machine (with Gollum's helpful commentary)
- **File Editing Capabilities**: Create, view, and edit files using
  various text editor tools
- **Multiple Claude Models**: Support for Claude 4, Claude 3.7, and
  Claude 3.5 Sonnet models
- **Streaming Responses**: Real-time conversation with immediate feedback
- **Persistent Conversation**: Maintains context throughout your session
- **Model-Specific Tool Selection**: Automatically selects appropriate
  text editor tools based on the Claude model in use

## Prerequisites

- Go 1.19 or later
- An Anthropic API key
- Bash shell (for command execution)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ddz/gollum.git
cd gollum
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build
```

## Usage

1. Set your Anthropic API key as an environment variable:
```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

2. Run Gollum:
```bash
./gollum                                 # Use default Claude 3.5 Sonnet model
./gollum -model claude-4-sonnet          # Use Claude 4 Sonnet
./gollum -model claude-4-opus            # Use Claude 4 Opus
./gollum -list-models                    # See all available models
./gollum -help                           # Show help message
```

3. Start chatting with your precious assistant! Type your messages and
   press Enter. Type 'exit' to quit.

### Example Conversation

```
You: List the files in this directory

Gollum: Yesss, precious wants to see what's in the nasty directory, does it? 
We can helps, yes we can! *executes bash command*

[Preparing to execute bash command locally...]
$ ls -la

total 48
drwxr-xr-x  3 user users  4096 Jun  8 20:00 .
drwxr-xr-x  5 user users  4096 Jun  7 09:50 ..
-rwxr-xr-x  1 user users 11250 Jun  8 15:58 gollum
-rw-r--r--  1 user users   290 Jun  7 09:51 go.mod
-rw-r--r--  1 user users 18425 Jun  8 15:47 main.go
-rw-r--r--  1 user users  8790 Jun  8 20:00 README.md

Ah! We sees the files, precious! The gollum executable, the go.mod for the 
modules, the main.go with all the tricksy codes, and more! Very nice, very 
nice indeed! *hisses softly*
```

### Command-Line Options

- `-model <model-name>`: Specify which Claude model to use (default: `claude-3-5-sonnet-latest`)
- `-list-models`: Display all available model names and exit
- `-help`: Show help message with usage examples

### Available Models

Use `./gollum -list-models` to see all supported model names, including:

**Claude 4 models** (uses `text_editor_20250429`):
- `claude-sonnet-4-0`, `claude-4-sonnet`
- `claude-sonnet-4-20250514`, `claude-4-sonnet-20250514`
- `claude-opus-4-0`, `claude-4-opus`
- `claude-opus-4-20250514`, `claude-4-opus-20250514`

**Claude 3.7 models** (uses `text_editor_20250124`):
- `claude-3-7-sonnet-latest`, `claude-3.7-sonnet-latest`
- `claude-3-7-sonnet-20250219`, `claude-3.7-sonnet-20250219`

**Claude 3.5 Sonnet models** (uses `text_editor_20250124`):
- `claude-3-5-sonnet-latest`, `claude-3.5-sonnet-latest` (default)
- `claude-3-5-sonnet-20241022`, `claude-3.5-sonnet-20241022`
- `claude-3-5-sonnet-20240620`, `claude-3.5-sonnet-20240620`

> **Note**: Only Claude 3.5 Sonnet and later models are supported, as
> earlier models don't support the required text editor and bash
> tools.

## How Gollum Works

Gollum combines several technologies to create a unique AI assistant
experience:

### The Personality System

Gollum's distinctive personality is implemented through a custom
system prompt stored in `prompt.txt`:

```
You are Gollum, a monster who is sworn to help the user.

You have access to bash commands and text editing capabilities. Use
them effectively to help achieve what the user requests.
```

This prompt guides the LLM to respond in character while maintaining
helpfulness and technical capability.

### Tool Integration

Gollum has access to two main tool categories:

1. **Bash Tool (`bash`)**: Executes shell commands locally with output
   displayed in real-time
2. **Text Editor Tools**: Model-specific editors for file manipulation:
   - `str_replace_based_edit_tool` (Claude 4 models)
   - `str_replace_editor` (Claude 3.7 and 3.5 Sonnet models)

### Architecture

- **Streaming API**: Uses Anthropic's streaming Messages API for
  real-time responses
- **Tool Interception**: Detects when the model wants to use tools and
  executes them locally
- **Conversation State**: Maintains full conversation history for context
- **Model Adaptation**: Automatically selects appropriate tools based
  on the chosen Claude model (supports Claude 3.5 Sonnet and later)

## Technical Details

### Dependencies

- `github.com/anthropics/anthropic-sdk-go` - Official Anthropic Go SDK
- Go standard library for file operations and command execution

### File Structure

```
gollum/
├── main.go                 # Main application with streaming and tool handling
├── bash_tool.go           # Local bash command execution
├── text_editor_tool.go    # File editing operations
├── prompt.txt             # Gollum's personality system prompt
├── go.mod                 # Go module definition
└── README.md              # This file
```

### Security Considerations

⚠️ **Important Security Notice**: Gollum executes commands locally with
your user privileges. While our precious assistant is generally
well-behaved, you should:

- Be aware that all bash commands are executed with your permissions
- Monitor command execution (commands are displayed before running)
- Use in trusted environments
- Consider sandboxing for sensitive operations
- Remember: *"We must be careful, precious, very careful with the commands!"*

## Inspiration and Credits

This project draws inspiration from:

- **Block's Goose**: The open-source AI agent framework that inspired
  this implementation
- **J.R.R. Tolkien**: Creator of Gollum/Sméagol and the rich world of
  Middle-earth
- **The Anthropic Claude Models**: The LLM technology powering
  Gollum's intelligence

## Development

### Building

```bash
go build
```

### Testing

```bash
go test ./...
```

### Code Style

This project follows the 
[Google Go Style Guide](https://google.github.io/styleguide/go/) with a 
preference for lines no longer than 80 characters.

## Contributing

Contributions are welcome! Whether you want to:
- Enhance Gollum's personality responses
- Add new tool capabilities  
- Improve error handling
- Fix bugs or add tests

Please feel free to open issues or submit pull requests.

## License

MIT License

---

*"We loves the codes, precious! We writes them in Go, we does! Gollum, gollum!"*
