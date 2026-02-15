package cli

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/git"
	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/rztaylor/GoDotFiles/internal/state"
)

func runHealthDoctorReport(gdfDir string) (*healthReport, error) {
	report := &healthReport{
		Command: "health doctor",
		OK:      true,
	}

	if !git.IsRepository(gdfDir) {
		report.add(healthFinding{
			Code:     "repo_not_initialized",
			Severity: healthSeverityError,
			Title:    "GDF repository is not initialized",
			Path:     gdfDir,
			Hint:     "Run 'gdf init' first",
		})
		report.sort()
		return report, nil
	}

	checkRequiredPaths(gdfDir, report)
	checkConfigAndState(gdfDir, report)
	checkShellIntegration(gdfDir, report)
	checkPackageManager(report)
	checkWritePermissions(gdfDir, report)
	report.sort()
	return report, nil
}

func checkRequiredPaths(gdfDir string, report *healthReport) {
	required := []struct {
		path     string
		code     string
		severity healthSeverity
		title    string
	}{
		{filepath.Join(gdfDir, "apps"), "apps_dir_missing", healthSeverityError, "Missing apps directory"},
		{filepath.Join(gdfDir, "profiles"), "profiles_dir_missing", healthSeverityError, "Missing profiles directory"},
		{filepath.Join(gdfDir, "dotfiles"), "dotfiles_dir_missing", healthSeverityError, "Missing dotfiles directory"},
		{filepath.Join(gdfDir, "generated"), "generated_dir_missing", healthSeverityWarning, "Generated directory is missing"},
	}

	for _, req := range required {
		if _, err := os.Stat(req.path); os.IsNotExist(err) {
			report.add(healthFinding{
				Code:     req.code,
				Severity: req.severity,
				Title:    req.title,
				Path:     req.path,
				Hint:     "Run 'gdf health fix' to create missing safe directories",
			})
		}
	}
}

func checkConfigAndState(gdfDir string, report *healthReport) {
	cfgPath := filepath.Join(gdfDir, "config.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		report.add(healthFinding{
			Code:     "config_missing",
			Severity: healthSeverityError,
			Title:    "Missing config.yaml",
			Path:     cfgPath,
			Hint:     "Run 'gdf health fix' to create a default config",
		})
	} else if _, err := config.LoadConfig(cfgPath); err != nil {
		report.add(healthFinding{
			Code:     "config_invalid",
			Severity: healthSeverityError,
			Title:    "Invalid config.yaml",
			Path:     cfgPath,
			Detail:   err.Error(),
		})
	}

	statePath := filepath.Join(gdfDir, "state.yaml")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		report.add(healthFinding{
			Code:     "state_missing",
			Severity: healthSeverityInfo,
			Title:    "state.yaml does not exist yet",
			Path:     statePath,
			Hint:     "This is created automatically after a successful apply",
		})
	} else if _, err := state.Load(statePath); err != nil {
		report.add(healthFinding{
			Code:     "state_invalid",
			Severity: healthSeverityError,
			Title:    "Invalid state.yaml",
			Path:     statePath,
			Detail:   err.Error(),
		})
	}
}

func checkShellIntegration(gdfDir string, report *healthReport) {
	initPath := filepath.Join(gdfDir, "generated", "init.sh")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		report.add(healthFinding{
			Code:     "generated_init_missing",
			Severity: healthSeverityWarning,
			Title:    "Generated shell init script is missing",
			Path:     initPath,
			Hint:     "Run 'gdf apply <profile>' to regenerate shell integration",
		})
	}

	shellName := platform.DetectShell()
	if shellName != "bash" && shellName != "zsh" {
		return
	}

	rcPath := detectRCPath(shellName)
	if rcPath == "" {
		return
	}

	hasSource, err := rcHasGDFSourceLine(rcPath)
	if err != nil {
		report.add(healthFinding{
			Code:     "rc_unreadable",
			Severity: healthSeverityWarning,
			Title:    "Shell RC file is unreadable",
			Path:     rcPath,
			Detail:   err.Error(),
		})
		return
	}

	if !hasSource {
		report.add(healthFinding{
			Code:     "rc_source_missing",
			Severity: healthSeverityWarning,
			Title:    "Shell RC file does not source GDF init script",
			Path:     rcPath,
			Hint:     "Run 'gdf health fix' or add: [ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh",
		})
	}
}

func checkPackageManager(report *healthReport) {
	p := platform.Detect()
	mgr := packages.ForPlatform(p)
	if mgr.Name() == "none" {
		report.add(healthFinding{
			Code:     "pkg_manager_unavailable",
			Severity: healthSeverityWarning,
			Title:    "No supported package manager detected",
			Detail:   "Package installation steps will be skipped on apply",
		})
	}
}

func checkWritePermissions(gdfDir string, report *healthReport) {
	f, err := os.CreateTemp(gdfDir, ".gdf-write-check-*")
	if err != nil {
		report.add(healthFinding{
			Code:     "repo_not_writable",
			Severity: healthSeverityError,
			Title:    "GDF repository is not writable",
			Path:     gdfDir,
			Detail:   err.Error(),
		})
		return
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
}

func detectRCPath(shellName string) string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}

	if shellName == "bash" {
		bashrc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		return filepath.Join(home, ".bash_profile")
	}

	return filepath.Join(home, ".zshrc")
}

func rcHasGDFSourceLine(rcPath string) (bool, error) {
	f, err := os.Open(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
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
