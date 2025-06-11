package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// StatefulBashTool maintains a persistent bash session across command executions.
// All commands are executed in the same bash process, allowing state to persist
// between commands (environment variables, current directory, etc.).
type StatefulBashTool struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mutex  sync.Mutex
}

// NewStatefulBashTool creates a new StatefulBashTool instance and starts a bash session.
func NewStatefulBashTool() *StatefulBashTool {
	tool := &StatefulBashTool{}
	tool.startSession()
	return tool
}

// startSession starts a new bash process and sets up the communication pipes.
func (s *StatefulBashTool) startSession() error {
	// Clean up existing session if any
	s.stopSession()

	// Start a new bash process
	s.cmd = exec.Command("bash")

	var err error
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return err
	}

	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		s.stdin.Close()
		return err
	}

	s.stderr, err = s.cmd.StderrPipe()
	if err != nil {
		s.stdin.Close()
		s.stdout.Close()
		return err
	}

	return s.cmd.Start()
}

// stopSession terminates the bash process and closes all pipes.
func (s *StatefulBashTool) stopSession() {
	if s.cmd != nil && s.cmd.Process != nil {
		if s.stdin != nil {
			s.stdin.Close()
		}
		if s.stdout != nil {
			s.stdout.Close()
		}
		if s.stderr != nil {
			s.stderr.Close()
		}
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}
}

// ExecuteCommand runs the given command in the persistent bash session.
// It returns the command's stdout, stderr, and any execution error.
func (s *StatefulBashTool) ExecuteCommand(command string) (stdout string, stderr string, err error) {
	return s.executeCommandInternal(command, 0)
}

// executeCommandInternal is the internal implementation with retry logic.
func (s *StatefulBashTool) executeCommandInternal(command string, retryCount int) (stdout string, stderr string, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Prevent infinite recursion
	if retryCount > 1 {
		return "", "", fmt.Errorf("exceeded maximum retry count for command: %s", command)
	}

	// Check if session needs to be started or restarted (including if process died)
	if s.cmd == nil || s.cmd.Process == nil || s.cmd.ProcessState != nil {
		if err := s.startSession(); err != nil {
			return "", "", fmt.Errorf("failed to start bash session: %w", err)
		}
	}

	// Special handling for exit commands - they terminate the session
	if strings.TrimSpace(command) == "exit 1" || strings.HasPrefix(strings.TrimSpace(command), "exit ") {
		// For exit commands, we can't capture output normally since they terminate the session
		_, writeErr := s.stdin.Write([]byte(command + "\n"))
		if writeErr != nil {
			return "", "", fmt.Errorf("failed to write exit command: %w", writeErr)
		}
		// Wait a moment for the session to terminate
		s.cmd.Wait()
		// Extract exit code from the command if possible
		parts := strings.Fields(strings.TrimSpace(command))
		if len(parts) >= 2 && parts[0] == "exit" {
			return "", "", fmt.Errorf("exit status %s", parts[1])
		}
		return "", "", fmt.Errorf("exit status 1")
	}

	// Create a unique marker to identify the end of command output
	marker := fmt.Sprintf("__GOLLUM_CMD_END_%d__", s.cmd.Process.Pid)
	exitCodeMarker := fmt.Sprintf("__GOLLUM_EXIT_CODE_%d__", s.cmd.Process.Pid)

	// Write the command followed by echo statements to mark the end and capture exit code
	fullCommand := fmt.Sprintf("%s\necho %s$? >&1\necho %s >&1\necho %s >&2\n", command, exitCodeMarker, marker, marker)

	_, err = s.stdin.Write([]byte(fullCommand))
	if err != nil {
		// Session might be dead, try to restart and retry
		if restartErr := s.startSession(); restartErr != nil {
			return "", "", fmt.Errorf("failed to write command and restart session: %w", restartErr)
		}
		// Unlock mutex temporarily for recursive call
		s.mutex.Unlock()
		result1, result2, result3 := s.executeCommandInternal(command, retryCount+1)
		s.mutex.Lock()
		return result1, result2, result3
	}

	// Read stdout until we see the markers
	var stdoutBuffer bytes.Buffer
	var exitCode string
	stdoutScanner := bufio.NewScanner(s.stdout)
	exitCodeFound := false
	
	for stdoutScanner.Scan() {
		line := stdoutScanner.Text()
		if strings.HasPrefix(line, exitCodeMarker) && !exitCodeFound {
			exitCode = strings.TrimPrefix(line, exitCodeMarker)
			exitCodeFound = true
			continue
		}
		if strings.TrimSpace(line) == marker {
			break
		}
		stdoutBuffer.WriteString(line + "\n")
	}

	// If we couldn't read anything due to scanner error, session might be dead
	if err := stdoutScanner.Err(); err != nil {
		// Try to restart and re-execute the command
		if restartErr := s.startSession(); restartErr != nil {
			return "", "", fmt.Errorf("session died and failed to restart: %w", restartErr)
		}
		// Unlock mutex temporarily for recursive call
		s.mutex.Unlock()
		result1, result2, result3 := s.executeCommandInternal(command, retryCount+1)
		s.mutex.Lock()
		return result1, result2, result3
	}

	// Read stderr until we see the marker
	var stderrBuffer bytes.Buffer
	stderrScanner := bufio.NewScanner(s.stderr)
	for stderrScanner.Scan() {
		line := stderrScanner.Text()
		if strings.TrimSpace(line) == marker {
			break
		}
		stderrBuffer.WriteString(line + "\n")
	}

	stdout = stdoutBuffer.String()
	stderr = stderrBuffer.String()

	// Check if command failed based on exit code
	if exitCode != "0" && exitCode != "" {
		err = fmt.Errorf("exit status %s", exitCode)
	}

	return stdout, stderr, err
}

// Restart terminates the current bash session and starts a new one.
// This clears all session state (environment variables, current directory, etc.).
func (s *StatefulBashTool) Restart() (message string, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.startSession(); err != nil {
		return "", fmt.Errorf("failed to restart bash session: %w", err)
	}
	return "Stateful bash session restarted", nil
}