package main

import (
	"runtime"
	"strings"
	"testing"
)

// testBashTool tests any BashTool implementation with common functionality
func testBashTool(t *testing.T, tool BashTool) {
	t.Helper()

	// Test basic echo command
	stdout, stderr, err := tool.ExecuteCommand("echo 'hello world'")
	if err != nil {
		t.Errorf("echo command failed: %v", err)
	}
	if stdout != "hello world\n" {
		t.Errorf("expected 'hello world\\n', got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}

	// Test command with error
	_, _, err = tool.ExecuteCommand("false")
	if err == nil {
		t.Error("false command should have failed")
	}

	// Test command with syntax error
	_, _, err = tool.ExecuteCommand("if [ 1 == ")
	if err == nil {
		t.Error("'if [ 1 == ' command should fail")
	}
	
	// Test command with both stdout and stderr
	stdout, stderr, err = tool.ExecuteCommand("echo 'stdout'; echo 'stderr' >&2")
	if err != nil {
		t.Errorf("stdout/stderr command failed: %v", err)
	}
	if stdout != "stdout\n" {
		t.Errorf("expected 'stdout\\n', got %q", stdout)
	}
	if stderr != "stderr\n" {
		t.Errorf("expected 'stderr\\n', got %q", stderr)
	}

	// Test pipes
	stdout, stderr, err = tool.ExecuteCommand("echo -e 'apple\\nbanana\\ncherry' | grep 'ban'")
	if err != nil {
		t.Errorf("pipe command failed: %v", err)
	}
	if stdout != "banana\n" {
		t.Errorf("expected 'banana\\n', got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}

	// Test special characters (skip on Windows)
	if runtime.GOOS != "windows" {
		tests := []struct {
			command  string
			expected string
		}{
			{"echo 'single quotes'", "single quotes\n"},
			{`echo "double quotes"`, "double quotes\n"},
			{"echo `echo nested`", "nested\n"},
			{"echo 'first'; echo 'second'", "first\nsecond\n"},
		}

		for _, test := range tests {
			stdout, _, err := tool.ExecuteCommand(test.command)
			if err != nil {
				t.Errorf("command %q failed: %v", test.command, err)
				continue
			}
			if stdout != test.expected {
				t.Errorf("command %q: expected %q, got %q", test.command, test.expected, stdout)
			}
		}
	}

	// Test restart functionality
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
	stdout, stderr, err = tool.ExecuteCommand("echo 'after restart'")
	if err != nil {
		t.Errorf("command after restart failed: %v", err)
	}
	if stdout != "after restart\n" {
		t.Errorf("expected 'after restart\\n', got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr after restart, got %q", stderr)
	}
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
		// Test environment variable persistence
		_, _, err := tool.ExecuteCommand("export TEST_VAR=test_value")
		if err != nil {
			t.Errorf("export command failed: %v", err)
		}

		stdout, _, err := tool.ExecuteCommand("echo $TEST_VAR")
		if err != nil {
			t.Errorf("echo variable failed: %v", err)
		}
		if stdout != "test_value\n" {
			t.Errorf("variable not preserved, expected 'test_value\\n', got %q", stdout)
		}

		// Test directory persistence
		_, _, err = tool.ExecuteCommand("cd /tmp")
		if err != nil {
			t.Errorf("cd command failed: %v", err)
		}

		stdout, _, err = tool.ExecuteCommand("pwd")
		if err != nil {
			t.Errorf("pwd command failed: %v", err)
		}
		if !strings.Contains(stdout, "/tmp") {
			t.Errorf("directory change not preserved, got %q", stdout)
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
		// Test exit commands (this should handle the special exit case)
		_, _, err := tool.ExecuteCommand("exit 1")
		if err == nil {
			t.Error("exit 1 command should have failed")
		}
		if !strings.Contains(err.Error(), "exit status 1") {
			t.Errorf("expected exit status error, got: %v", err)
		}

		// Test another exit command with different code
		_, _, err = tool.ExecuteCommand("exit 42")
		if err == nil {
			t.Error("exit 42 command should have failed")
		}
		if !strings.Contains(err.Error(), "exit status 42") {
			t.Errorf("expected exit status 42 error, got: %v", err)
		}

		// Test plain exit command (defaults to exit 0, which should not be an error in this implementation)
		_, _, err = tool.ExecuteCommand("exit")
		// Note: plain "exit" might not always fail depending on implementation
		// Just check that we can handle it without crashing
		t.Logf("plain exit command result: %v", err)
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
