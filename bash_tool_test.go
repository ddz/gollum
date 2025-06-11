// Package main contains tests for the BashTool interface and its implementations.
//
// The test design follows a pattern where interface-level tests are separated
// from implementation-specific tests:
//
//  1. testBashToolInterface() contains comprehensive tests that can be applied
//     to any BashTool implementation, testing the interface contract.
//
//  2. TestStatelessBashTool() tests the specific StatelessBashTool implementation,
//     including constructor behavior and running the interface tests.
//
// This design makes it easy to test new BashTool implementations by reusing
// the interface test suite.
package main

import (
	"fmt"
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

// TestStatelessBashTool tests the StatelessBashTool implementation
func TestStatelessBashTool(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		tool := NewStatelessBashTool()
		if tool == nil {
			t.Fatal("NewStatelessBashTool() returned nil")
		}

		// Verify it implements the BashTool interface
		var _ BashTool = tool
	})

	t.Run("Interface", func(t *testing.T) {
		tool := NewStatelessBashTool()
		testBashToolInterface(t, tool)
	})
}

// TestStatefulBashTool tests the StatefulBashTool implementation
func TestStatefulBashTool(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		tool := NewStatefulBashTool()
		if tool == nil {
			t.Fatal("NewStatefulBashTool() returned nil")
		}

		// Verify it implements the BashTool interface
		var _ BashTool = tool

		// Clean up
		defer func() {
			tool.stopSession()
		}()
	})

	t.Run("Interface", func(t *testing.T) {
		tool := NewStatefulBashTool()
		defer func() {
			tool.stopSession()
		}()
		testBashToolInterface(t, tool)
	})

	t.Run("StatefulBehavior", func(t *testing.T) {
		testStatefulBashTool_StatefulBehavior(t)
	})

	t.Run("EnvironmentPersistence", func(t *testing.T) {
		testStatefulBashTool_EnvironmentPersistence(t)
	})

	t.Run("DirectoryPersistence", func(t *testing.T) {
		testStatefulBashTool_DirectoryPersistence(t)
	})

	t.Run("RestartClearsState", func(t *testing.T) {
		testStatefulBashTool_RestartClearsState(t)
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		testStatefulBashTool_ConcurrentAccess(t)
	})

	t.Run("SessionRecovery", func(t *testing.T) {
		testStatefulBashTool_SessionRecovery(t)
	})

	t.Run("RestartMessage", func(t *testing.T) {
		testStatefulBashTool_RestartMessage(t)
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
		/*
			{
				name:        "invalid syntax",
				command:     "if [ missing bracket",
				expectError: true,
			},
		*/
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

	// Accept both StatelessBashTool and StatefulBashTool restart messages
	validMessages := []string{
		"Bash session restarted",
		"Stateful bash session restarted",
	}

	validMessage := false
	for _, validMsg := range validMessages {
		if message == validMsg {
			validMessage = true
			break
		}
	}

	if !validMessage {
		t.Errorf("Expected restart message to be one of %v, got %q", validMessages, message)
	}
}

// testBashTool_Restart_Multiple tests multiple restart calls
func testBashTool_Restart_Multiple(t *testing.T, tool BashTool) {
	t.Helper()

	// Valid restart messages for both implementations
	validMessages := []string{
		"Bash session restarted",
		"Stateful bash session restarted",
	}

	// Test that restart can be called multiple times
	for i := 0; i < 3; i++ {
		message, err := tool.Restart()

		if err != nil {
			t.Errorf("Restart() call %d: Expected no error, but got: %v", i+1, err)
		}

		validMessage := false
		for _, validMsg := range validMessages {
			if message == validMsg {
				validMessage = true
				break
			}
		}

		if !validMessage {
			t.Errorf("Restart() call %d: Expected message to be one of %v, got %q",
				i+1, validMessages, message)
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

// testStatefulBashTool_StatefulBehavior tests that state persists between commands
func testStatefulBashTool_StatefulBehavior(t *testing.T) {
	t.Helper()

	tool := NewStatefulBashTool()
	defer tool.stopSession()

	// Test that variables persist between commands
	_, _, err := tool.ExecuteCommand("TEST_VAR=stateful_value")
	if err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}

	stdout, stderr, err := tool.ExecuteCommand("echo $TEST_VAR")
	if err != nil {
		t.Fatalf("Failed to echo variable: %v", err)
	}

	if stderr != "" {
		t.Errorf("Unexpected stderr: %q", stderr)
	}

	expectedOutput := "stateful_value\n"
	if stdout != expectedOutput {
		t.Errorf("Expected stdout %q, got %q", expectedOutput, stdout)
	}
}

// testStatefulBashTool_EnvironmentPersistence tests environment variable persistence
func testStatefulBashTool_EnvironmentPersistence(t *testing.T) {
	t.Helper()

	tool := NewStatefulBashTool()
	defer tool.stopSession()

	// Set multiple environment variables
	commands := []string{
		"export VAR1=value1",
		"export VAR2=value2",
		"VAR3=value3",
	}

	for _, cmd := range commands {
		_, _, err := tool.ExecuteCommand(cmd)
		if err != nil {
			t.Fatalf("Failed to execute command %q: %v", cmd, err)
		}
	}

	// Test that all variables are accessible
	tests := []struct {
		command  string
		expected string
	}{
		{"echo $VAR1", "value1\n"},
		{"echo $VAR2", "value2\n"},
		{"echo $VAR3", "value3\n"},
		{"echo \"$VAR1-$VAR2-$VAR3\"", "value1-value2-value3\n"},
	}

	for _, test := range tests {
		stdout, stderr, err := tool.ExecuteCommand(test.command)
		if err != nil {
			t.Errorf("Failed to execute command %q: %v", test.command, err)
			continue
		}

		if stderr != "" {
			t.Errorf("Command %q produced stderr: %q", test.command, stderr)
		}

		if stdout != test.expected {
			t.Errorf("Command %q: expected %q, got %q", test.command, test.expected, stdout)
		}
	}
}

// testStatefulBashTool_DirectoryPersistence tests that directory changes persist
func testStatefulBashTool_DirectoryPersistence(t *testing.T) {
	t.Helper()

	tool := NewStatefulBashTool()
	defer tool.stopSession()

	// Get initial directory
	initialDir, _, err := tool.ExecuteCommand("pwd")
	if err != nil {
		t.Fatalf("Failed to get initial directory: %v", err)
	}
	initialDir = strings.TrimSpace(initialDir)

	// Change to /tmp
	_, _, err = tool.ExecuteCommand("cd /tmp")
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Verify we're in /tmp
	currentDir, _, err := tool.ExecuteCommand("pwd")
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	currentDir = strings.TrimSpace(currentDir)

	if currentDir != "/tmp" {
		t.Errorf("Expected current directory to be /tmp, got %q", currentDir)
	}

	// Change back to original directory
	_, _, err = tool.ExecuteCommand("cd " + initialDir)
	if err != nil {
		t.Fatalf("Failed to change back to initial directory: %v", err)
	}

	// Verify we're back
	finalDir, _, err := tool.ExecuteCommand("pwd")
	if err != nil {
		t.Fatalf("Failed to get final directory: %v", err)
	}
	finalDir = strings.TrimSpace(finalDir)

	if finalDir != initialDir {
		t.Errorf("Expected to be back in initial directory %q, got %q", initialDir, finalDir)
	}
}

// testStatefulBashTool_RestartClearsState tests that restart clears session state
func testStatefulBashTool_RestartClearsState(t *testing.T) {
	t.Helper()

	tool := NewStatefulBashTool()
	defer tool.stopSession()

	// Set environment variable and change directory
	_, _, err := tool.ExecuteCommand("export TEST_RESTART=before_restart")
	if err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}

	_, _, err = tool.ExecuteCommand("cd /tmp")
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Verify state is set
	stdout, _, err := tool.ExecuteCommand("echo $TEST_RESTART")
	if err != nil {
		t.Fatalf("Failed to echo variable: %v", err)
	}
	if strings.TrimSpace(stdout) != "before_restart" {
		t.Errorf("Variable not set correctly before restart: %q", stdout)
	}

	// Restart the session
	message, err := tool.Restart()
	if err != nil {
		t.Fatalf("Failed to restart: %v", err)
	}

	expectedMessage := "Stateful bash session restarted"
	if message != expectedMessage {
		t.Errorf("Expected restart message %q, got %q", expectedMessage, message)
	}

	// Verify state is cleared
	stdout, _, err = tool.ExecuteCommand("echo $TEST_RESTART")
	if err != nil {
		t.Fatalf("Failed to echo variable after restart: %v", err)
	}

	// Variable should be empty after restart
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("Expected empty variable after restart, got %q", stdout)
	}

	// Directory should be reset (not /tmp anymore)
	stdout, _, err = tool.ExecuteCommand("pwd")
	if err != nil {
		t.Fatalf("Failed to get directory after restart: %v", err)
	}

	currentDir := strings.TrimSpace(stdout)
	if currentDir == "/tmp" {
		t.Errorf("Directory should have been reset after restart, still in /tmp")
	}
}

// testStatefulBashTool_ConcurrentAccess tests concurrent access to the same tool
func testStatefulBashTool_ConcurrentAccess(t *testing.T) {
	t.Helper()

	tool := NewStatefulBashTool()
	defer tool.stopSession()

	// Use a channel to synchronize goroutines
	done := make(chan bool, 2)
	errors := make(chan error, 2)

	// Goroutine 1: Set and read variable
	go func() {
		defer func() { done <- true }()

		_, _, err := tool.ExecuteCommand("export CONCURRENT_VAR=goroutine1")
		if err != nil {
			errors <- err
			return
		}

		stdout, _, err := tool.ExecuteCommand("echo $CONCURRENT_VAR")
		if err != nil {
			errors <- err
			return
		}

		if strings.TrimSpace(stdout) != "goroutine1" {
			errors <- fmt.Errorf("goroutine1: expected 'goroutine1', got %q", strings.TrimSpace(stdout))
		}
	}()

	// Goroutine 2: Different operations
	go func() {
		defer func() { done <- true }()

		_, _, err := tool.ExecuteCommand("echo 'hello from goroutine2'")
		if err != nil {
			errors <- err
			return
		}

		// Small delay to let goroutine1 potentially interfere
		_, _, err = tool.ExecuteCommand("sleep 0.1")
		if err != nil {
			errors <- err
			return
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

// testStatefulBashTool_SessionRecovery tests session recovery after process death
func testStatefulBashTool_SessionRecovery(t *testing.T) {
	t.Helper()

	tool := NewStatefulBashTool()
	defer tool.stopSession()

	// Execute a normal command first
	stdout, _, err := tool.ExecuteCommand("echo 'before kill'")
	if err != nil {
		t.Fatalf("Failed to execute command before kill: %v", err)
	}
	if strings.TrimSpace(stdout) != "before kill" {
		t.Errorf("Unexpected output before kill: %q", stdout)
	}

	// Manually kill the bash process to simulate session death
	if tool.cmd != nil && tool.cmd.Process != nil {
		tool.cmd.Process.Kill()
		tool.cmd.Wait() // Wait for the process to actually terminate
	}

	// Try to execute another command - should recover automatically
	stdout, _, err = tool.ExecuteCommand("echo 'after recovery'")
	if err != nil {
		t.Fatalf("Failed to execute command after recovery: %v", err)
	}
	if strings.TrimSpace(stdout) != "after recovery" {
		t.Errorf("Unexpected output after recovery: %q", stdout)
	}
}

// testStatefulBashTool_RestartMessage tests the restart message
func testStatefulBashTool_RestartMessage(t *testing.T) {
	t.Helper()

	tool := NewStatefulBashTool()
	defer tool.stopSession()

	message, err := tool.Restart()
	if err != nil {
		t.Fatalf("Restart failed: %v", err)
	}

	expectedMessage := "Stateful bash session restarted"
	if message != expectedMessage {
		t.Errorf("Expected restart message %q, got %q", expectedMessage, message)
	}

	// Test multiple restarts
	for i := 0; i < 3; i++ {
		message, err := tool.Restart()
		if err != nil {
			t.Errorf("Restart %d failed: %v", i+1, err)
		}
		if message != expectedMessage {
			t.Errorf("Restart %d: expected message %q, got %q", i+1, expectedMessage, message)
		}
	}
}
