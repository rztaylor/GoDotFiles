package updater

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rztaylor/GoDotFiles/internal/state"
)

// PromptUpdate asks the user if they want to update.
func PromptUpdate(release *ReleaseInfo, st *state.State) error {
	fmt.Printf("\n📦 New version available: %s (current: %s)\n", release.Version, "current_version_placeholder")
	fmt.Printf("Run 'gdf update' to update manually or answer below.\n")
	fmt.Print("Update now? [Y/n] ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "y" || response == "yes" || response == "" {
		fmt.Println("Updating...")
		return Update()
	}

	// If No, snooze for 7 days
	fmt.Println("You will not be prompted to update for 7 days.")
	fmt.Println("To disable auto-update checks run 'gdf update --never'")

	st.UpdateCheck.SnoozeUntil = time.Now().Add(7 * 24 * time.Hour)

	if st.Path != "" {
		if err := st.Save(st.Path); err != nil {
			fmt.Printf("Warning: failed to save snooze state: %v\n", err)
		}
	}
	return nil
}
