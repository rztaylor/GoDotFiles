package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var globalYes bool
var globalNonInteractive bool

func confirmPrompt(prompt string) (bool, error) {
	if globalYes {
		return true, nil
	}
	if globalNonInteractive {
		return false, withExitCode(fmt.Errorf("confirmation required but running in non-interactive mode"), exitCodeNonInteractiveStop)
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("checking stdin: %w", err)
	}
	if fi.Mode()&os.ModeCharDevice == 0 {
		return false, withExitCode(fmt.Errorf("confirmation required but stdin is not interactive"), exitCodeNonInteractiveStop)
	}

	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("reading confirmation input: %w", err)
	}
	v := strings.ToLower(strings.TrimSpace(input))
	return v == "y" || v == "yes", nil
}
