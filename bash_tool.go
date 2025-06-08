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

type SimpleBashTool struct {
}

func NewSimpleBashTool() *SimpleBashTool {
	return &SimpleBashTool{}
}

func (*SimpleBashTool) ExecuteCommand(command string) (stdout string, stderr string, err error) {
	var stdoutBuffer, stderrBuffer bytes.Buffer

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer

	err = cmd.Run()
	stdout = stdoutBuffer.String()
	stderr = stderrBuffer.String()

	return
}

func (*SimpleBashTool) Restart() (message string, err error) {
	// This simple, stateless bash tool implementation does not
	// need to be restarted, but tell Claude that we did so
	// anyway.
	return "Bash session restarted", nil
}
