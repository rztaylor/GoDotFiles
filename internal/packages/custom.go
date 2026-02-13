package packages

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
)

// Custom executes custom installation scripts with user confirmation.
type Custom struct {
	// promptFunc allows mocking user prompts in tests
	promptFunc func(message string) (bool, error)

	// execCommand allows mocking command execution in tests
	execCommand func(string, ...string) *exec.Cmd

	// logFunc allows mocking operation logging in tests
	logFunc func(operation string) error
}

// NewCustom creates a new Custom script executor.
func NewCustom() *Custom {
	return &Custom{
		promptFunc:  promptUser,
		execCommand: exec.Command,
		logFunc:     logOperation,
	}
}

// Execute runs a custom installation script with user confirmation.
func (c *Custom) Execute(customInstall *apps.CustomInstall) error {
	if customInstall == nil {
		return fmt.Errorf("custom install configuration cannot be nil")
	}

	if customInstall.Script == "" {
		return fmt.Errorf("custom install script cannot be empty")
	}

	// Build prompt message
	message := c.buildPromptMessage(customInstall.Script, customInstall.Sudo)

	// Get confirmation (defaults to true if not specified)
	if customInstall.ConfirmDefault() {
		promptFunc := c.promptFunc
		if promptFunc == nil {
			promptFunc = promptUser
		}

		confirmed, err := promptFunc(message)
		if err != nil {
			return fmt.Errorf("failed to get user confirmation: %w", err)
		}

		if !confirmed {
			return fmt.Errorf("user declined custom script execution")
		}
	}

	// Log the operation for audit
	if c.logFunc != nil {
		_ = c.logFunc(fmt.Sprintf("custom script: %s", customInstall.Script))
	}

	// Execute the script
	execCmd := c.execCommand
	if execCmd == nil {
		execCmd = exec.Command
	}

	var cmd *exec.Cmd
	if customInstall.Sudo {
		cmd = execCmd("sudo", "sh", "-c", customInstall.Script)
	} else {
		cmd = execCmd("sh", "-c", customInstall.Script)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute custom script: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// buildPromptMessage creates the confirmation prompt message.
func (c *Custom) buildPromptMessage(script string, sudo bool) string {
	var msg strings.Builder

	msg.WriteString("\n⚠️  Custom Installation Script\n\n")
	msg.WriteString(fmt.Sprintf("Script: %s\n\n", script))

	if sudo {
		msg.WriteString("⚠️  This script requires sudo privileges.\n\n")
	}

	msg.WriteString("This will download and execute a script from an external source.\n")
	msg.WriteString("Only proceed if you trust the source.\n\n")
	msg.WriteString("Proceed with script execution? [y/N]: ")

	return msg.String()
}

// promptUser prompts the user for yes/no confirmation.
func promptUser(message string) (bool, error) {
	fmt.Print(message)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// logOperation logs a custom script operation for audit purposes.
func logOperation(operation string) error {
	// TODO: Implement proper operation logging to .gdf/.operations/
	// For now, just print to stdout
	fmt.Printf("[AUDIT] %s\n", operation)
	return nil
}
