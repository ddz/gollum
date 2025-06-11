package main

import (
	"bytes"
	"os/exec"
)

// BashTool is the interface expected by Claude's bash tool use
type BashTool interface {
	// ExecuteCommand runs the given command in bash. Its output
	// to standard out is returned in stdout. Its output to
	// standard error is returned in stderr. If the command
	// returns an successful exit code, then err is nil. If the
	// command returned an error exit code, then err will be an
	// error value.
	ExecuteCommand(command string) (stdout string, stderr string, err error)

	// Restart resets the persistent bash session (if any) that is
	// used across multiple invocations of ExecuteCommand.
	Restart() (message string, err error)
}

func checkBashCommand(command string) (stdout string, stderr string, err error) {
	var stdoutBuffer, stderrBuffer bytes.Buffer

	// Check for syntax errors first with bash -n
	cmd := exec.Command("bash", "-nc", command)
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer

	err = cmd.Run()
	stdout = stdoutBuffer.String()
	stderr = stderrBuffer.String()

	return
}
