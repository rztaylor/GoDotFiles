package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/config"
)

const profileSelectionPromptDefaultCancel = "profile selection cancelled"

func resolveProfileSelection(gdfDir, requestedProfile string) (string, error) {
	return resolveProfileSelectionForCommand(gdfDir, requestedProfile, "")
}

func resolveProfileSelectionForCommand(gdfDir, requestedProfile, commandLabel string) (string, error) {
	requestedProfile = strings.TrimSpace(requestedProfile)
	if requestedProfile != "" {
		return requestedProfile, nil
	}

	profiles, err := config.LoadAllProfiles(filepath.Join(gdfDir, "profiles"))
	if err != nil {
		return "", fmt.Errorf("loading profiles: %w", err)
	}

	names := profileNames(profiles)
	switch len(names) {
	case 0:
		return "", fmt.Errorf("no profiles found\n\nHow to fix\n  1. Run `gdf profile create <name>`\n  2. Re-run with `--profile <name>`")
	case 1:
		return names[0], nil
	default:
		return chooseProfileInteractive(names, commandLabel)
	}
}

func chooseProfileInteractive(names []string, commandLabel string) (string, error) {
	if globalNonInteractive {
		return "", withExitCode(
			fmt.Errorf("multiple profiles found (%s)\n\nHow to fix\n  1. Re-run with `--profile <name>`\n  2. Check profiles with `gdf profile list`", strings.Join(names, ", ")),
			exitCodeNonInteractiveStop,
		)
	}

	printSectionHeading("Step 1/2: Select Profile")
	if strings.TrimSpace(commandLabel) == "" {
		fmt.Println("  Required context is missing. Choose a profile to continue.")
	} else {
		fmt.Printf("  `%s` needs a profile. Choose one to continue.\n", commandLabel)
	}
	fmt.Println()
	for i, name := range names {
		fmt.Printf("  %d. %s\n", i+1, name)
	}
	input, err := readInteractiveLine(fmt.Sprintf("Select profile [1-%d] (Enter to cancel): ", len(names)))
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf(profileSelectionPromptDefaultCancel)
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(names) {
		return "", fmt.Errorf("invalid profile selection: %q", input)
	}
	selected := names[idx-1]
	fmt.Println()
	printSectionHeading("Step 2/2: Confirm Selection")
	printKeyValueLines([]keyValue{{Key: "Profile", Value: selected}})
	return selected, nil
}

func profileNames(profiles []*config.Profile) []string {
	names := make([]string, 0, len(profiles))
	for _, p := range profiles {
		names = append(names, p.Name)
	}
	sort.Strings(names)
	return names
}
