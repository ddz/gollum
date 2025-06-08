// Package main contains tests for the BashTool interface and its implementations.
//
// The test design follows a pattern where interface-level tests are separated
// from implementation-specific tests:
//
//  1. testBashToolInterface() contains comprehensive tests that can be applied
//     to any BashTool implementation, testing the interface contract.
//
//  2. TestSimpleBashTool() tests the specific SimpleBashTool implementation,
//     including constructor behavior and running the interface tests.
//
// This design makes it easy to test new BashTool implementations by reusing
// the interface test suite.
package main

import (
	"runtime"
	"strings"
	"testing"
)

// testBashToolInterface runs comprehensive tests against any BashTool implementation.
// This allows us to reuse the same test suite for different implementations.
//
// Example usage for testing a new implementation:
//
//	func TestMyNewBashTool(t *testing.T) {
//		t.Run("Constructor", func(t *testing.T) {
//			tool := NewMyNewBashTool()
//			if tool == nil {
//				t.Fatal("NewMyNewBashTool() returned nil")
//			}
//			// Verify it implements the BashTool interface
//			var _ BashTool = tool
//		})
//
//		t.Run("Interface", func(t *testing.T) {
//			tool := NewMyNewBashTool()
//			testBashToolInterface(t, tool)
//		})
//	}
func testBashToolInterface(t *testing.T, tool BashTool) {
	t.Helper()

	t.Run("ExecuteCommand_Success", func(t *testing.T) {
		testBashTool_ExecuteCommand_Success(t, tool)
	})

	t.Run("ExecuteCommand_Errors", func(t *testing.T) {
		testBashTool_ExecuteCommand_Errors(t, tool)
	})

	t.Run("ExecuteCommand_WithBothOutputs", func(t *testing.T) {
		testBashTool_ExecuteCommand_WithBothOutputs(t, tool)
	})

	t.Run("ExecuteCommand_EmptyCommand", func(t *testing.T) {
		testBashTool_ExecuteCommand_EmptyCommand(t, tool)
	})

	t.Run("ExecuteCommand_LongRunningCommand", func(t *testing.T) {
		testBashTool_ExecuteCommand_LongRunningCommand(t, tool)
	})

	t.Run("ExecuteCommand_EnvironmentVariables", func(t *testing.T) {
		testBashTool_ExecuteCommand_EnvironmentVariables(t, tool)
	})

	t.Run("ExecuteCommand_PipeCommands", func(t *testing.T) {
		testBashTool_ExecuteCommand_PipeCommands(t, tool)
	})

	t.Run("Restart", func(t *testing.T) {
		testBashTool_Restart(t, tool)
	})

	t.Run("Restart_Multiple", func(t *testing.T) {
		testBashTool_Restart_Multiple(t, tool)
	})

	t.Run("ExecuteCommand_AfterRestart", func(t *testing.T) {
		testBashTool_ExecuteCommand_AfterRestart(t, tool)
	})

	t.Run("CommandWithSpecialCharacters", func(t *testing.T) {
		testBashTool_CommandWithSpecialCharacters(t, tool)
	})
}

// TestSimpleBashTool tests the SimpleBashTool implementation
func TestSimpleBashTool(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		tool := NewSimpleBashTool()
		if tool == nil {
			t.Fatal("NewSimpleBashTool() returned nil")
		}

		// Verify it implements the BashTool interface
		var _ BashTool = tool
	})

	t.Run("Interface", func(t *testing.T) {
		tool := NewSimpleBashTool()
		testBashToolInterface(t, tool)
	})
}

// testBashTool_ExecuteCommand_Success tests successful command execution scenarios
func testBashTool_ExecuteCommand_Success(t *testing.T, tool BashTool) {
	t.Helper()

	tests := []struct {
		name           string
		command        string
		expectedStdout string
		expectedStderr string
		expectError    bool
	}{
		{
			name:           "simple echo command",
			command:        "echo 'Hello, World!'",
			expectedStdout: "Hello, World!\n",
			expectedStderr: "",
			expectError:    false,
		},
		{
			name:           "echo to stderr",
			command:        "echo 'Error message' >&2",
			expectedStdout: "",
			expectedStderr: "Error message\n",
			expectError:    false,
		},
		{
			name:           "simple arithmetic",
			command:        "echo $((2 + 3))",
			expectedStdout: "5\n",
			expectedStderr: "",
			expectError:    false,
		},
		{
			name:           "pwd command",
			command:        "pwd",
			expectedStdout: "", // We'll check that it's not empty
			expectedStderr: "",
			expectError:    false,
		},
		{
			name:           "multiline output",
			command:        "echo -e 'line1\\nline2\\nline3'",
			expectedStdout: "line1\nline2\nline3\n",
			expectedStderr: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := tool.ExecuteCommand(tt.command)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if tt.name == "pwd command" {
				// For pwd, just check that we got some output
				if strings.TrimSpace(stdout) == "" {
					t.Errorf("Expected non-empty stdout for pwd command, got empty string")
				}
			} else if stdout != tt.expectedStdout {
				t.Errorf("Expected stdout %q, got %q", tt.expectedStdout, stdout)
			}

			if stderr != tt.expectedStderr {
				t.Errorf("Expected stderr %q, got %q", tt.expectedStderr, stderr)
			}
		})
	}
}

// testBashTool_ExecuteCommand_Errors tests error handling scenarios
func testBashTool_ExecuteCommand_Errors(t *testing.T, tool BashTool) {
	t.Helper()

	tests := []struct {
		name        string
		command     string
		expectError bool
	}{
		{
			name:        "command not found",
			command:     "nonexistentcommand12345",
			expectError: true,
		},
		{
			name:        "false command",
			command:     "false",
			expectError: true,
		},
		{
			name:        "exit with non-zero code",
			command:     "exit 1",
			expectError: true,
		},
		{
			name:        "invalid syntax",
			command:     "if [ missing bracket",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := tool.ExecuteCommand(tt.command)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for command %q, but got nil", tt.command)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for command %q, but got: %v", tt.command, err)
			}

			// For error cases, we should still get the output
			// (stdout and stderr should be captured regardless of error)
			t.Logf("Command: %q, stdout: %q, stderr: %q, err: %v",
				tt.command, stdout, stderr, err)
		})
	}
}

// testBashTool_ExecuteCommand_WithBothOutputs tests commands that output to both stdout and stderr
func testBashTool_ExecuteCommand_WithBothOutputs(t *testing.T, tool BashTool) {
	t.Helper()

	// Command that outputs to both stdout and stderr
	command := "echo 'stdout message'; echo 'stderr message' >&2"
	stdout, stderr, err := tool.ExecuteCommand(command)

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	expectedStdout := "stdout message\n"
	expectedStderr := "stderr message\n"

	if stdout != expectedStdout {
		t.Errorf("Expected stdout %q, got %q", expectedStdout, stdout)
	}

	if stderr != expectedStderr {
		t.Errorf("Expected stderr %q, got %q", expectedStderr, stderr)
	}
}

// testBashTool_ExecuteCommand_EmptyCommand tests empty command handling
func testBashTool_ExecuteCommand_EmptyCommand(t *testing.T, tool BashTool) {
	t.Helper()

	stdout, stderr, err := tool.ExecuteCommand("")

	// Empty command should succeed with no output
	if err != nil {
		t.Errorf("Expected no error for empty command, but got: %v", err)
	}

	if stdout != "" {
		t.Errorf("Expected empty stdout for empty command, got %q", stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr for empty command, got %q", stderr)
	}
}

// testBashTool_ExecuteCommand_LongRunningCommand tests commands that take time to execute
func testBashTool_ExecuteCommand_LongRunningCommand(t *testing.T, tool BashTool) {
	t.Helper()

	// Command that takes a short time but produces output
	command := "for i in {1..3}; do echo $i; done"
	stdout, stderr, err := tool.ExecuteCommand(command)

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	expectedStdout := "1\n2\n3\n"
	if stdout != expectedStdout {
		t.Errorf("Expected stdout %q, got %q", expectedStdout, stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got %q", stderr)
	}
}

// testBashTool_ExecuteCommand_EnvironmentVariables tests environment variable usage
func testBashTool_ExecuteCommand_EnvironmentVariables(t *testing.T, tool BashTool) {
	t.Helper()

	// Test setting and using environment variables
	command := "export TEST_VAR='hello world'; echo $TEST_VAR"
	stdout, stderr, err := tool.ExecuteCommand(command)

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	expectedStdout := "hello world\n"
	if stdout != expectedStdout {
		t.Errorf("Expected stdout %q, got %q", expectedStdout, stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got %q", stderr)
	}
}

// testBashTool_ExecuteCommand_PipeCommands tests piped commands
func testBashTool_ExecuteCommand_PipeCommands(t *testing.T, tool BashTool) {
	t.Helper()

	// Test piped commands
	command := "echo -e 'apple\\nbanana\\ncherry' | grep 'ban'"
	stdout, stderr, err := tool.ExecuteCommand(command)

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	expectedStdout := "banana\n"
	if stdout != expectedStdout {
		t.Errorf("Expected stdout %q, got %q", expectedStdout, stdout)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got %q", stderr)
	}
}

// testBashTool_Restart tests the restart functionality
func testBashTool_Restart(t *testing.T, tool BashTool) {
	t.Helper()

	message, err := tool.Restart()

	if err != nil {
		t.Errorf("Expected no error from Restart(), but got: %v", err)
	}

	expectedMessage := "Bash session restarted"
	if message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, message)
	}
}

// testBashTool_Restart_Multiple tests multiple restart calls
func testBashTool_Restart_Multiple(t *testing.T, tool BashTool) {
	t.Helper()

	// Test that restart can be called multiple times
	for i := 0; i < 3; i++ {
		message, err := tool.Restart()

		if err != nil {
			t.Errorf("Restart() call %d: Expected no error, but got: %v", i+1, err)
		}

		expectedMessage := "Bash session restarted"
		if message != expectedMessage {
			t.Errorf("Restart() call %d: Expected message %q, got %q",
				i+1, expectedMessage, message)
		}
	}
}

// testBashTool_ExecuteCommand_AfterRestart tests command execution after restart
func testBashTool_ExecuteCommand_AfterRestart(t *testing.T, tool BashTool) {
	t.Helper()

	// Execute a command, restart, then execute another command
	stdout1, stderr1, err1 := tool.ExecuteCommand("echo 'before restart'")
	if err1 != nil {
		t.Errorf("First command failed: %v", err1)
	}

	_, err := tool.Restart()
	if err != nil {
		t.Errorf("Restart failed: %v", err)
	}

	stdout2, stderr2, err2 := tool.ExecuteCommand("echo 'after restart'")
	if err2 != nil {
		t.Errorf("Second command failed: %v", err2)
	}

	if stdout1 != "before restart\n" {
		t.Errorf("Expected first stdout 'before restart\\n', got %q", stdout1)
	}

	if stdout2 != "after restart\n" {
		t.Errorf("Expected second stdout 'after restart\\n', got %q", stdout2)
	}

	if stderr1 != "" || stderr2 != "" {
		t.Errorf("Expected empty stderr for both commands, got %q and %q",
			stderr1, stderr2)
	}
}

// testBashTool_CommandWithSpecialCharacters tests commands with special shell characters
func testBashTool_CommandWithSpecialCharacters(t *testing.T, tool BashTool) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("Skipping bash-specific test on Windows")
	}

	tests := []struct {
		name           string
		command        string
		expectedStdout string
		expectError    bool
	}{
		{
			name:           "single quotes",
			command:        "echo 'hello world'",
			expectedStdout: "hello world\n",
			expectError:    false,
		},
		{
			name:           "double quotes",
			command:        `echo "hello world"`,
			expectedStdout: "hello world\n",
			expectError:    false,
		},
		{
			name:           "backticks",
			command:        "echo `echo nested`",
			expectedStdout: "nested\n",
			expectError:    false,
		},
		{
			name:           "semicolon separation",
			command:        "echo 'first'; echo 'second'",
			expectedStdout: "first\nsecond\n",
			expectError:    false,
		},
		{
			name:           "ampersand redirection",
			command:        "echo 'to stderr' >&2; echo 'to stdout'",
			expectedStdout: "to stdout\n",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := tool.ExecuteCommand(tt.command)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if stdout != tt.expectedStdout {
				t.Errorf("Expected stdout %q, got %q", tt.expectedStdout, stdout)
			}

			// Log stderr for debugging (some tests expect stderr output)
			if stderr != "" {
				t.Logf("Command %q produced stderr: %q", tt.command, stderr)
			}
		})
	}
}
