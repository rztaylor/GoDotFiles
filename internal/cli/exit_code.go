package cli

import "errors"

const (
	exitCodeSuccess            = 0
	exitCodeRuntimeError       = 1
	exitCodeHealthIssues       = 2
	exitCodeFixFailure         = 3
	exitCodeNonInteractiveStop = 4
)

type exitCodeError struct {
	code int
	err  error
}

func (e *exitCodeError) Error() string {
	if e.err == nil {
		return "command failed"
	}
	return e.err.Error()
}

func (e *exitCodeError) Unwrap() error {
	return e.err
}

func withExitCode(err error, code int) error {
	if err == nil {
		return nil
	}
	return &exitCodeError{code: code, err: err}
}

// ExitCode returns a stable process exit code for CLI errors.
func ExitCode(err error) int {
	if err == nil {
		return exitCodeSuccess
	}

	var coded *exitCodeError
	if errors.As(err, &coded) {
		return coded.code
	}

	return exitCodeRuntimeError
}
