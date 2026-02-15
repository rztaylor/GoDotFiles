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
		return "", fmt.Errorf("no profiles found; create one with 'gdf profile create <name>' or pass --profile")
	case 1:
		return names[0], nil
	default:
		return chooseProfileInteractive(names)
	}
}

func chooseProfileInteractive(names []string) (string, error) {
	if globalNonInteractive {
		return "", withExitCode(
			fmt.Errorf("multiple profiles found (%s); rerun with --profile", strings.Join(names, ", ")),
			exitCodeNonInteractiveStop,
		)
	}

	fmt.Println("Multiple profiles found:")
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
	return names[idx-1], nil
}

func profileNames(profiles []*config.Profile) []string {
	names := make([]string, 0, len(profiles))
	for _, p := range profiles {
		names = append(names, p.Name)
	}
	sort.Strings(names)
	return names
}
