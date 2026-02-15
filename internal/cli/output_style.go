package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var globalVerbose bool
var globalColorMode = "auto"

var uiColorSectionHeadings = true
var uiHighlightKeyValues = true

type outputStatus string

const (
	outputStatusOK    outputStatus = "ok"
	outputStatusWarn  outputStatus = "warn"
	outputStatusError outputStatus = "error"
	outputStatusInfo  outputStatus = "info"
)

const (
	ansiReset = "\033[0m"

	ansiBlueBold   = "\033[1;34m"
	ansiCyan       = "\033[36m"
	ansiGreen      = "\033[32m"
	ansiYellow     = "\033[33m"
	ansiRed        = "\033[31m"
	ansiBrightBlue = "\033[94m"
)

type keyValue struct {
	Key   string
	Value string
}

func configureOutputStyle(cmd *cobra.Command) error {
	mode, err := normalizeColorMode(globalColorMode)
	if err != nil {
		return err
	}
	globalColorMode = mode

	uiColorSectionHeadings = true
	uiHighlightKeyValues = true

	cfg, err := config.LoadConfig(platform.ConfigFile())
	if err != nil {
		return nil
	}
	if cfg.UI == nil {
		return nil
	}

	if cmd == nil || !cmd.Flags().Changed("color") {
		cfgMode, cfgModeErr := normalizeColorMode(cfg.UI.ColorDefault())
		if cfgModeErr == nil {
			globalColorMode = cfgMode
		}
	}
	uiColorSectionHeadings = cfg.UI.ColorSectionHeadingsDefault()
	uiHighlightKeyValues = cfg.UI.HighlightKeyValuesDefault()

	return nil
}

func normalizeColorMode(mode string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "auto":
		return "auto", nil
	case "always":
		return "always", nil
	case "never":
		return "never", nil
	default:
		return "", fmt.Errorf("invalid --color value %q (expected: auto, always, never)", mode)
	}
}

func useColorOutput() bool {
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}

	switch globalColorMode {
	case "always":
		return true
	case "never":
		return false
	default:
		fi, err := os.Stdout.Stat()
		if err != nil {
			return false
		}
		return fi.Mode()&os.ModeCharDevice != 0
	}
}

func colorize(value, ansi string) string {
	if !useColorOutput() {
		return value
	}
	return ansi + value + ansiReset
}

func statusIcon(status outputStatus) string {
	switch status {
	case outputStatusOK:
		return colorize("✓", ansiGreen)
	case outputStatusWarn:
		return colorize("!", ansiYellow)
	case outputStatusError:
		return colorize("✗", ansiRed)
	default:
		return colorize("i", ansiBrightBlue)
	}
}

func printStatusLine(status outputStatus, message string) {
	fmt.Printf("%s %s\n", statusIcon(status), message)
}

func printSectionHeading(title string) {
	if uiColorSectionHeadings {
		fmt.Println(colorize(title, ansiBlueBold))
		return
	}
	fmt.Println(title)
}

func printKeyValueLines(items []keyValue) {
	if len(items) == 0 {
		return
	}

	maxKeyLen := 0
	for _, item := range items {
		if len(item.Key) > maxKeyLen {
			maxKeyLen = len(item.Key)
		}
	}

	for _, item := range items {
		keyLabel := fmt.Sprintf("%-*s:", maxKeyLen, item.Key)
		if uiHighlightKeyValues {
			keyLabel = colorize(keyLabel, ansiCyan)
		}
		fmt.Printf("  %s %s\n", keyLabel, item.Value)
	}
}

func printNextStep(command string) {
	fmt.Println()
	printSectionHeading("Next Step")
	fmt.Printf("  Run `%s`\n", command)
}
