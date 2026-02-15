package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rztaylor/GoDotFiles/internal/engine"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Undo the last apply operation using operation logs and snapshots",
	Long: `Rollback restores the previous state from the most recent operation log.

When snapshot history exists for a file, rollback can restore historical copies.
By default it uses the snapshot from the latest operation. Use --choose-snapshot
to select among multiple historical restore points.`,
	RunE: runRollback,
}

var rollbackYes bool
var rollbackChooseSnapshot bool
var rollbackTarget string
var rollbackConfirmPrompt = confirmRollbackPrompt

func init() {
	recoverCmd.AddCommand(rollbackCmd)
	rollbackCmd.Flags().BoolVar(&rollbackYes, "yes", false, "Skip confirmation prompt")
	rollbackCmd.Flags().BoolVar(&rollbackChooseSnapshot, "choose-snapshot", false, "Prompt to choose snapshot versions when multiple exist")
	rollbackCmd.Flags().StringVar(&rollbackTarget, "target", "", "Restore a specific target path from snapshot history")
}

func runRollback(cmd *cobra.Command, args []string) error {
	gdfDir := platform.ConfigDir()

	if rollbackTarget != "" {
		return rollbackSingleTarget(gdfDir, rollbackTarget, rollbackChooseSnapshot)
	}

	logPath, ops, err := engine.LatestOperationLog(gdfDir)
	if err != nil {
		return err
	}
	if logPath == "" || len(ops) == 0 {
		fmt.Println("No operation logs found; nothing to rollback.")
		return nil
	}

	fmt.Printf("Using operation log: %s\n", logPath)
	linkOps := 0
	withSnapshots := 0
	for _, op := range ops {
		if op.Type != "link" {
			continue
		}
		linkOps++
		if op.Details != nil && op.Details["snapshot_path"] != "" {
			withSnapshots++
		}
	}
	fmt.Printf("Rollback plan: %d link operations (%d with historical snapshots)\n", linkOps, withSnapshots)

	if !rollbackYes {
		ok, err := rollbackConfirmPrompt("Proceed with rollback? [y/N]: ")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Rollback aborted.")
			return nil
		}
	}

	var selector func(string, []engine.SnapshotCandidate) (*engine.SnapshotCandidate, error)
	if rollbackChooseSnapshot {
		selector = chooseSnapshotCandidatePrompt
	}

	result := engine.RollbackOperations(gdfDir, ops, selector)
	fmt.Printf("Rollback complete: restored=%d removed=%d failures=%d\n", result.Restored, result.Removed, len(result.Failed))
	for _, f := range result.Failed {
		fmt.Printf("  - %s\n", f)
	}

	if len(result.Failed) > 0 {
		return fmt.Errorf("rollback completed with %d failures", len(result.Failed))
	}
	return nil
}

func rollbackSingleTarget(gdfDir, target string, choose bool) error {
	expanded := platform.ExpandPath(target)
	candidates, err := engine.FindSnapshotCandidates(gdfDir, expanded)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no snapshots found for %s", expanded)
	}

	selected := &candidates[0]
	if choose && len(candidates) > 1 {
		pick, err := chooseSnapshotCandidatePrompt(expanded, candidates)
		if err != nil {
			return err
		}
		if pick != nil {
			selected = pick
		}
	}

	if !rollbackYes {
		prompt := fmt.Sprintf("Restore %s from snapshot %s (%s)? [y/N]: ", expanded, selected.SnapshotPath, selected.CapturedAt.Format("2006-01-02 15:04:05"))
		ok, err := rollbackConfirmPrompt(prompt)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Rollback aborted.")
			return nil
		}
	}

	res := engine.RollbackOperations(gdfDir, []engine.Operation{
		{
			Type:   "link",
			Target: expanded,
			Details: map[string]string{
				"snapshot_path":        selected.SnapshotPath,
				"snapshot_kind":        selected.SnapshotKind,
				"snapshot_mode":        selected.SnapshotMode,
				"snapshot_link_target": selected.LinkTarget,
				"snapshot_captured_at": selected.CapturedAt.Format(time.RFC3339Nano),
			},
		},
	}, nil)
	if len(res.Failed) > 0 {
		return fmt.Errorf("restore failed: %s", strings.Join(res.Failed, "; "))
	}
	fmt.Printf("Restored %s from snapshot captured at %s\n", expanded, selected.CapturedAt.Format("2006-01-02 15:04:05"))
	return nil
}

func confirmRollbackPrompt(prompt string) (bool, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	v := strings.ToLower(strings.TrimSpace(input))
	return v == "y" || v == "yes", nil
}

func chooseSnapshotCandidatePrompt(target string, candidates []engine.SnapshotCandidate) (*engine.SnapshotCandidate, error) {
	fmt.Printf("Multiple snapshots found for %s:\n", target)
	for i, c := range candidates {
		fmt.Printf("  %d) %s  %s\n", i+1, c.CapturedAt.Format("2006-01-02 15:04:05"), c.SnapshotPath)
	}
	fmt.Print("Select snapshot number: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || n < 1 || n > len(candidates) {
		return nil, fmt.Errorf("invalid selection")
	}
	return &candidates[n-1], nil
}
