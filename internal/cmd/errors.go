package cmd

import (
	"fmt"
	"os/exec"
)

type cmdOutputError struct {
	err    error
	path   string
	args   []string
	output []byte
}

func newCmdOutputError(cmd *exec.Cmd, output []byte, err error) *cmdOutputError {
	return &cmdOutputError{
		path:   cmd.Path,
		args:   cmd.Args,
		output: output,
		err:    err,
	}
}

func (e *cmdOutputError) Error() string {
	if len(e.output) == 0 {
		return fmt.Sprintf("%s: %v", shellQuoteCommand(e.path, e.args[1:]), e.err)
	}
	return fmt.Sprintf("%s: %v\n%s", shellQuoteCommand(e.path, e.args[1:]), e.err, e.output)
}

func (e *cmdOutputError) Unwrap() error {
	return e.err
}

type parseCmdOutputError struct {
	err     error
	command string
	args    []string
	output  []byte
}

func newParseCmdOutputError(command string, args []string, output []byte, err error) *parseCmdOutputError {
	return &parseCmdOutputError{
		command: command,
		args:    args,
		output:  output,
		err:     err,
	}
}

func (e *parseCmdOutputError) Error() string {
	return fmt.Sprintf("%s: %v\n%s", shellQuoteCommand(e.command, e.args), e.err, e.output)
}

func (e *parseCmdOutputError) Unwrap() error {
	return e.err
}

type parseVersionError struct {
	err    error
	output []byte
}

func (e *parseVersionError) Error() string {
	return fmt.Sprintf("%s: cannot parse version: %v", e.output, e.err)
}

func (e *parseVersionError) Unwrap() error {
	return e.err
}
