package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var globalYes bool
var globalNonInteractive bool

func confirmPrompt(prompt string) (bool, error) {
	return confirmPromptWithOptions(prompt, false, true, false)
}

func confirmPromptDefaultYes(prompt string) (bool, error) {
	return confirmPromptWithOptions(prompt, true, true, true)
}

func confirmPromptUnsafe(prompt string) (bool, error) {
	return confirmPromptWithOptions(prompt, false, false, false)
}

func confirmPromptWithOptions(prompt string, defaultYes, allowGlobalYes, allowDefaultWithoutPrompt bool) (bool, error) {
	if allowGlobalYes && globalYes {
		return true, nil
	}
	if allowDefaultWithoutPrompt {
		if globalNonInteractive {
			return defaultYes, nil
		}
		fi, err := os.Stdin.Stat()
		if err == nil && fi.Mode()&os.ModeCharDevice == 0 {
			return defaultYes, nil
		}
	}

	input, err := readInteractiveLine(prompt)
	if err != nil {
		if allowDefaultWithoutPrompt && errors.Is(err, io.EOF) {
			return defaultYes, nil
		}
		return false, err
	}
	v := strings.ToLower(strings.TrimSpace(input))
	if v == "" {
		return defaultYes, nil
	}
	return v == "y" || v == "yes", nil
}

func readInteractiveLine(prompt string) (string, error) {
	if err := ensureInteractiveStdin(); err != nil {
		return "", err
	}

	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading interactive input: %w", err)
	}
	return input, nil
}

func ensureInteractiveStdin() error {
	if globalNonInteractive {
		return withExitCode(fmt.Errorf("interactive input required but running in non-interactive mode"), exitCodeNonInteractiveStop)
	}
	return nil
}
