package cli

import (
	"errors"
	"testing"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "nil error",
			err:  nil,
			want: exitCodeSuccess,
		},
		{
			name: "runtime error default",
			err:  errors.New("boom"),
			want: exitCodeRuntimeError,
		},
		{
			name: "coded error",
			err:  withExitCode(errors.New("health issues"), exitCodeHealthIssues),
			want: exitCodeHealthIssues,
		},
		{
			name: "wrapped coded error",
			err:  errors.Join(errors.New("outer"), withExitCode(errors.New("blocked"), exitCodeNonInteractiveStop)),
			want: exitCodeNonInteractiveStop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExitCode(tt.err)
			if got != tt.want {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}
