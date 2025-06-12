package main

import (
	"runtime"
	"strings"
	"testing"
)

// testBashTool tests any BashTool implementation with common functionality
func testBashTool(t *testing.T, tool BashTool) {
	t.Helper()

	// Table-driven tests for command execution
	tests := []struct {
		name           string
		command        string
		expectedStdout string
		expectedStderr string
		shouldError    bool
		skipOnWindows  bool
	}{
		{
			name:           "BasicEcho",
			command:        "echo 'hello world'",
			expectedStdout: "hello world\n",
			expectedStderr: "",
			shouldError:    false,
		},
		{
			name:        "CommandWithError",
			command:     "false",
			shouldError: true,
		},
		{
			name:        "SyntaxError",
			command:     "if [ 1 == ",
			shouldError: true,
		},
		{
			name:           "StdoutAndStderr",
			command:        "echo 'stdout'; echo 'stderr' >&2",
			expectedStdout: "stdout\n",
			expectedStderr: "stderr\n",
			shouldError:    false,
		},
		{
			name:           "PipeCommand",
			command:        "echo -e 'apple\\nbanana\\ncherry' | grep 'ban'",
			expectedStdout: "banana\n",
			expectedStderr: "",
			shouldError:    false,
		},
		{
			name:           "SingleQuotes",
			command:        "echo 'single quotes'",
			expectedStdout: "single quotes\n",
			expectedStderr: "",
			shouldError:    false,
			skipOnWindows:  true,
		},
		{
			name:           "DoubleQuotes",
			command:        `echo "double quotes"`,
			expectedStdout: "double quotes\n",
			expectedStderr: "",
			shouldError:    false,
			skipOnWindows:  true,
		},
		{
			name:           "NestedBackticks",
			command:        "echo `echo nested`",
			expectedStdout: "nested\n",
			expectedStderr: "",
			shouldError:    false,
			skipOnWindows:  true,
		},
		{
			name:           "MultipleCommands",
			command:        "echo 'first'; echo 'second'",
			expectedStdout: "first\nsecond\n",
			expectedStderr: "",
			shouldError:    false,
			skipOnWindows:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOnWindows && runtime.GOOS == "windows" {
				t.Skip("Skipping on Windows")
			}

			stdout, stderr, err := tool.ExecuteCommand(tt.command)

			if tt.shouldError {
				if err == nil {
					t.Errorf("command %q should have failed", tt.command)
				}
				return
			}

			if err != nil {
				t.Errorf("command %q failed: %v", tt.command, err)
				return
			}

			if stdout != tt.expectedStdout {
				t.Errorf("stdout mismatch for %q: expected %q, got %q", tt.command, tt.expectedStdout, stdout)
			}

			if stderr != tt.expectedStderr {
				t.Errorf("stderr mismatch for %q: expected %q, got %q", tt.command, tt.expectedStderr, stderr)
			}
		})
	}

	// Test restart functionality
	t.Run("RestartFunctionality", func(t *testing.T) {
		msg, err := tool.Restart()
		if err != nil {
			t.Errorf("restart failed: %v", err)
		}

		// Accept both possible restart messages
		validMessages := []string{
			"Bash session restarted",
			"Stateful bash session restarted",
		}
		validMessage := false
		for _, validMsg := range validMessages {
			if msg == validMsg {
				validMessage = true
				break
			}
		}
		if !validMessage {
			t.Errorf("unexpected restart message: %q", msg)
		}

		// Test command execution after restart
		stdout, stderr, err := tool.ExecuteCommand("echo 'after restart'")
		if err != nil {
			t.Errorf("command after restart failed: %v", err)
		}
		if stdout != "after restart\n" {
			t.Errorf("expected 'after restart\\n', got %q", stdout)
		}
		if stderr != "" {
			t.Errorf("expected empty stderr after restart, got %q", stderr)
		}
	})
}

func TestStatelessBashTool(t *testing.T) {
	tool := NewStatelessBashTool()
	if tool == nil {
		t.Fatal("NewStatelessBashTool() returned nil")
	}

	// Verify it implements the BashTool interface
	var _ BashTool = tool

	// Run common tests
	testBashTool(t, tool)
}

func TestStatefulBashTool(t *testing.T) {
	tool := NewStatefulBashTool()
	if tool == nil {
		t.Fatal("NewStatefulBashTool() returned nil")
	}
	defer tool.stopSession()

	// Verify it implements the BashTool interface
	var _ BashTool = tool

	// Run common tests
	testBashTool(t, tool)

	// Test stateful-specific behavior
	t.Run("StatePersistence", func(t *testing.T) {
		persistenceTests := []struct {
			name           string
			setupCommand   string
			testCommand    string
			expectedOutput string
			outputContains string
		}{
			{
				name:           "EnvironmentVariable",
				setupCommand:   "export TEST_VAR=test_value",
				testCommand:    "echo $TEST_VAR",
				expectedOutput: "test_value\n",
			},
			{
				name:           "DirectoryChange",
				setupCommand:   "cd /tmp",
				testCommand:    "pwd",
				outputContains: "/tmp",
			},
		}

		for _, tt := range persistenceTests {
			t.Run(tt.name, func(t *testing.T) {
				// Setup
				_, _, err := tool.ExecuteCommand(tt.setupCommand)
				if err != nil {
					t.Errorf("setup command %q failed: %v", tt.setupCommand, err)
					return
				}

				// Test
				stdout, _, err := tool.ExecuteCommand(tt.testCommand)
				if err != nil {
					t.Errorf("test command %q failed: %v", tt.testCommand, err)
					return
				}

				if tt.expectedOutput != "" && stdout != tt.expectedOutput {
					t.Errorf("expected exact output %q, got %q", tt.expectedOutput, stdout)
				}

				if tt.outputContains != "" && !strings.Contains(stdout, tt.outputContains) {
					t.Errorf("expected output to contain %q, got %q", tt.outputContains, stdout)
				}
			})
		}
	})

	t.Run("RestartClearsState", func(t *testing.T) {
		// Set a variable
		_, _, err := tool.ExecuteCommand("export RESTART_TEST=value")
		if err != nil {
			t.Errorf("export command failed: %v", err)
		}

		// Restart
		_, err = tool.Restart()
		if err != nil {
			t.Errorf("restart failed: %v", err)
		}

		// Check that variable is cleared
		stdout, _, err := tool.ExecuteCommand("echo $RESTART_TEST")
		if err != nil {
			t.Errorf("echo variable after restart failed: %v", err)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Errorf("variable should be cleared after restart, got %q", stdout)
		}
	})

	t.Run("ExitCommands", func(t *testing.T) {
		exitTests := []struct {
			name          string
			command       string
			shouldError   bool
			errorContains string
		}{
			{
				name:          "ExitWithCode1",
				command:       "exit 1",
				shouldError:   true,
				errorContains: "exit status 1",
			},
			{
				name:          "ExitWithCode42",
				command:       "exit 42",
				shouldError:   true,
				errorContains: "exit status 42",
			},
			{
				name:    "PlainExit",
				command: "exit",
				// Note: plain "exit" might not always fail depending on implementation
			},
		}

		for _, tt := range exitTests {
			t.Run(tt.name, func(t *testing.T) {
				_, _, err := tool.ExecuteCommand(tt.command)

				if tt.shouldError {
					if err == nil {
						t.Errorf("%s command should have failed", tt.command)
						return
					}
					if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
						t.Errorf("expected error containing %q, got: %v", tt.errorContains, err)
					}
				} else {
					// For plain exit, just log the result
					t.Logf("%s command result: %v", tt.command, err)
				}
			})
		}
	})

	t.Run("SessionRecovery", func(t *testing.T) {
		// Kill the bash process to simulate session death
		if tool.cmd != nil && tool.cmd.Process != nil {
			tool.cmd.Process.Kill()
			tool.cmd.Wait()
		}

		// Next command should recover automatically
		stdout, _, err := tool.ExecuteCommand("echo 'recovered'")
		if err != nil {
			t.Errorf("command after session death failed: %v", err)
		}
		if stdout != "recovered\n" {
			t.Errorf("expected 'recovered\\n', got %q", stdout)
		}
	})
}
