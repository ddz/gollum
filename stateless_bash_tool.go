package main

import (
	"bytes"
	"os/exec"
)

// StatelessBashTool implements BashTool without maintaining state between command executions.
// Each command is executed in a separate bash process.
type StatelessBashTool struct {
}

// NewStatelessBashTool creates a new StatelessBashTool instance.
func NewStatelessBashTool() *StatelessBashTool {
	return &StatelessBashTool{}
}

// ExecuteCommand runs the given command in a new bash process.
// It returns the command's stdout, stderr, and any execution error.
func (*StatelessBashTool) ExecuteCommand(command string) (stdout string, stderr string, err error) {
	var stdoutBuffer, stderrBuffer bytes.Buffer

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer

	err = cmd.Run()
	stdout = stdoutBuffer.String()
	stderr = stderrBuffer.String()

	return
}

// Restart is a no-op for StatelessBashTool since there's no persistent session to restart.
// It returns a message indicating that the bash session was restarted.
func (*StatelessBashTool) Restart() (message string, err error) {
	// This simple, stateless bash tool implementation does not
	// need to be restarted, but tell Claude that we did so
	// anyway.
	return "Bash session restarted", nil
}