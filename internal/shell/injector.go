package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const sourceLineComment = "# Added by gdf for shell integration"
const sourceLine = "[ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh"

// Injector handles injecting the source line into RC files.
type Injector struct{}

// NewInjector creates a new RC injector.
func NewInjector() *Injector {
	return &Injector{}
}

// InjectSourceLine adds the GDF source line to the appropriate RC file.
func (i *Injector) InjectSourceLine(shellType ShellType) error {
	rcPath := i.getRCPath(shellType)
	if rcPath == "" {
		return fmt.Errorf("cannot determine RC file for shell type: %s", shellType)
	}

	// Check if already injected
	hasSource, err := i.hasSourceLine(rcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to check RC file: %w", err)
	}

	if hasSource {
		// Already injected, nothing to do
		return nil
	}

	// Backup existing RC file if it exists
	if _, err := os.Stat(rcPath); err == nil {
		backupPath := rcPath + ".gdf.backup"
		if err := copyFile(rcPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup RC file: %w", err)
		}
	}

	// Append source line
	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open RC file: %w", err)
	}
	defer f.Close()

	// Add newlines if file doesn't end with one
	stat, _ := f.Stat()
	if stat.Size() > 0 {
		// Read last byte to check for newline
		content, _ := os.ReadFile(rcPath)
		if len(content) > 0 && content[len(content)-1] != '\n' {
			f.WriteString("\n")
		}
		f.WriteString("\n")
	}

	// Write source line with comment
	_, err = f.WriteString(fmt.Sprintf("%s\n%s\n", sourceLineComment, sourceLine))
	if err != nil {
		return fmt.Errorf("failed to write source line: %w", err)
	}

	return nil
}

// getRCPath returns the RC file path for the given shell type.
func (i *Injector) getRCPath(shellType ShellType) string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}

	switch shellType {
	case Bash:
		// Prefer .bashrc, but use .bash_profile on macOS if .bashrc doesn't exist
		bashrc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		// Fall back to .bash_profile
		return filepath.Join(home, ".bash_profile")
	case Zsh:
		return filepath.Join(home, ".zshrc")
	default:
		return ""
	}
}

// hasSourceLine checks if the RC file already contains the GDF source line.
func (i *Injector) hasSourceLine(rcPath string) (bool, error) {
	f, err := os.Open(rcPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "~/.gdf/generated/init.sh") {
			return true, nil
		}
	}

	return false, scanner.Err()
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
