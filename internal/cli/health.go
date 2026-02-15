package cli

import (
	"fmt"

	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run configuration and environment health checks",
}

var healthValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate GDF config, profiles, and app definitions",
	RunE:  runHealthValidate,
}

var healthDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check environment health and readiness",
	RunE:  runHealthDoctor,
}

var healthFixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Apply safe auto-fixes for common health issues",
	RunE:  runHealthFixCmdE,
}

var healthCICmd = &cobra.Command{
	Use:   "ci",
	Short: "Run fail-fast checks intended for CI",
	RunE:  runHealthCI,
}

var healthValidateJSON bool
var healthDoctorJSON bool
var healthCIJSON bool
var healthFixGuarded bool
var healthFixDryRun bool

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.AddCommand(healthValidateCmd)
	healthCmd.AddCommand(healthDoctorCmd)
	healthCmd.AddCommand(healthFixCmd)
	healthCmd.AddCommand(healthCICmd)

	healthValidateCmd.Flags().BoolVar(&healthValidateJSON, "json", false, "Output findings as JSON")
	healthDoctorCmd.Flags().BoolVar(&healthDoctorJSON, "json", false, "Output findings as JSON")
	healthCICmd.Flags().BoolVar(&healthCIJSON, "json", false, "Output findings as JSON")
	healthFixCmd.Flags().BoolVar(&healthFixGuarded, "guarded", false, "Include guarded, higher-impact fixes that require backups")
	healthFixCmd.Flags().BoolVar(&healthFixDryRun, "dry-run", false, "Preview fix actions without making changes")
}

func runHealthValidate(cmd *cobra.Command, args []string) error {
	report, err := runHealthValidateReport(platform.ConfigDir())
	if err != nil {
		return err
	}

	if healthValidateJSON {
		if err := writeHealthJSON(cmd.OutOrStdout(), report); err != nil {
			return err
		}
	} else {
		writeHealthText(cmd.OutOrStdout(), report)
	}

	if report.Errors > 0 {
		return withExitCode(fmt.Errorf("validation issues found"), exitCodeHealthIssues)
	}
	return nil
}

func runHealthDoctor(cmd *cobra.Command, args []string) error {
	report, err := runHealthDoctorReport(platform.ConfigDir())
	if err != nil {
		return err
	}

	if healthDoctorJSON {
		if err := writeHealthJSON(cmd.OutOrStdout(), report); err != nil {
			return err
		}
	} else {
		writeHealthText(cmd.OutOrStdout(), report)
	}

	if report.Errors > 0 {
		return withExitCode(fmt.Errorf("doctor found blocking issues"), exitCodeHealthIssues)
	}
	return nil
}

func runHealthFixCmdE(cmd *cobra.Command, args []string) error {
	return runHealthFix(platform.ConfigDir(), cmd.OutOrStdout())
}

func runHealthCI(cmd *cobra.Command, args []string) error {
	validate, err := runHealthValidateReport(platform.ConfigDir())
	if err != nil {
		return err
	}
	doctor, err := runHealthDoctorReport(platform.ConfigDir())
	if err != nil {
		return err
	}

	combined := &healthReport{
		Command: "health ci",
		OK:      true,
	}
	seen := map[string]bool{}
	addUnique := func(f healthFinding) {
		key := fmt.Sprintf("%s|%s|%s", f.Code, f.Path, f.Title)
		if seen[key] {
			return
		}
		seen[key] = true
		combined.add(f)
	}
	for _, f := range validate.Findings {
		addUnique(f)
	}
	for _, f := range doctor.Findings {
		addUnique(f)
	}
	combined.sort()

	if healthCIJSON {
		if err := writeHealthJSON(cmd.OutOrStdout(), combined); err != nil {
			return err
		}
	} else {
		writeHealthText(cmd.OutOrStdout(), combined)
	}

	if combined.Errors > 0 {
		return withExitCode(fmt.Errorf("CI health checks failed"), exitCodeHealthIssues)
	}
	return nil
}
