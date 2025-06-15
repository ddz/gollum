package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestCommandHandlerWithWriter(t *testing.T) {
	// Create a new reader
	reader, err := NewReader()
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Register a test command that writes to the provided writer
	reader.RegisterCommand("test", "Test command for testing", func(w io.Writer) error {
		w.Write([]byte("Hello from test command!"))
		return nil
	})

	// Test the command handler directly
	err = reader.handleSpecialCommand("/test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test with a custom writer
	reader.RegisterCommand("testwriter", "Test command with custom writer", func(w io.Writer) error {
		w.Write([]byte("Custom writer output"))
		return nil
	})

	// Capture output using our buffer
	testHandler := reader.cmds["testwriter"].Handler
	err = testHandler(&buf)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Custom writer output") {
		t.Errorf("Expected output to contain 'Custom writer output', got: %s", output)
	}
}

func TestBuiltinCommandsWithWriter(t *testing.T) {
	reader, err := NewReader()
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// Test help command with a buffer
	var buf bytes.Buffer
	helpHandler := reader.cmds["help"].Handler
	err = helpHandler(&buf)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Special commands:") {
		t.Errorf("Expected help output to contain 'Special commands:', got: %s", output)
	}

	// Test exit command with a buffer
	buf.Reset()
	exitHandler := reader.cmds["exit"].Handler
	err = exitHandler(&buf)
	if err != io.EOF {
		t.Errorf("Expected io.EOF from exit command, got %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("Expected exit output to contain 'Goodbye!', got: %s", output)
	}
}

func TestCommandHandlerErrorPropagation(t *testing.T) {
	// Create a new reader
	reader, err := NewReader()
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// Register a command that returns an error
	testError := fmt.Errorf("test command error")
	reader.RegisterCommand("testerror", "Test command that returns an error", func(w io.Writer) error {
		w.Write([]byte("Command executed before error"))
		return testError
	})

	// Test handleSpecialCommand directly
	err = reader.handleSpecialCommand("/testerror")
	if err == nil {
		t.Error("Expected error from handleSpecialCommand, got nil")
	}
	if err != testError {
		t.Errorf("Expected specific test error, got: %v", err)
	}

	// Test handleSpecialCommand error propagation again
	err = reader.handleSpecialCommand("/testerror")
	if err == nil {
		t.Error("Expected error from handleSpecialCommand, got nil")
	}
	if err != testError {
		t.Errorf("Expected specific test error, got: %v", err)
	}
}

func TestExitCommandReturnsEOF(t *testing.T) {
	// Create a new reader
	reader, err := NewReader()
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// Test handleSpecialCommand with exit command
	err = reader.handleSpecialCommand("/exit")
	if err != io.EOF {
		t.Errorf("Expected io.EOF from exit command, got: %v", err)
	}

	// Test handleSpecialCommand with exit command again
	err = reader.handleSpecialCommand("/exit")
	if err != io.EOF {
		t.Errorf("Expected io.EOF from handleSpecialCommand with exit command, got: %v", err)
	}
}

func TestAutoCompleteIncludesRegisteredCommands(t *testing.T) {
	// Create a new reader
	reader, err := NewReader()
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// Get the initial set of registered commands (built-in commands)
	initialCommands := reader.commands()

	// Verify that all commands in the auto-completer correspond to registered handlers
	for _, cmd := range initialCommands {
		if _, exists := reader.cmds[cmd]; !exists {
			t.Errorf("Command '%s' found in registered commands but no handler exists", cmd)
		}
	}

	// Register a new command and verify it gets added to auto-completion
	reader.RegisterCommand("testcmd", "Test command for auto-completion", func(w io.Writer) error {
		w.Write([]byte("test command"))
		return nil
	})

	// Verify the new command is now in the registered commands
	updatedCommands := reader.commands()
	found := false
	for _, cmd := range updatedCommands {
		if cmd == "testcmd" {
			found = true
			break
		}
	}
	if !found {
		t.Error("New command 'testcmd' not found in registered commands after registration")
	}

	// Verify all commands still have handlers
	for _, cmd := range updatedCommands {
		if _, exists := reader.cmds[cmd]; !exists {
			t.Errorf("Command '%s' found in registered commands but no handler exists", cmd)
		}
	}
}

func TestBuiltinCommandsArrayUsage(t *testing.T) {
	// Create a new reader
	reader, err := NewReader()
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// Verify that all built-in commands from the global array are registered
	registeredCommands := reader.commands()

	for _, builtinCmd := range builtinCommands {
		found := false
		for _, registeredCmd := range registeredCommands {
			if registeredCmd == builtinCmd.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Built-in command '%s' from global array not found in registered commands", builtinCmd.Name)
		}
	}

	// Test that help command uses the global array for descriptions
	var buf bytes.Buffer
	helpHandler := reader.cmds["help"].Handler
	err = helpHandler(&buf)
	if err != nil {
		t.Errorf("Expected no error from help command, got %v", err)
	}

	output := buf.String()

	// Verify that help output contains descriptions from the global array
	for _, builtinCmd := range builtinCommands {
		if !strings.Contains(output, builtinCmd.Description) {
			t.Errorf("Help output doesn't contain description for command '%s': %s", builtinCmd.Name, builtinCmd.Description)
		}
	}
}

