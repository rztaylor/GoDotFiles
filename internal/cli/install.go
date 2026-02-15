package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/config"
	"github.com/rztaylor/GoDotFiles/internal/packages"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <app>",
	Short: "Install an app and learn how to do it",
	Long: `Install an app on the current system.

If the app is already defined and has instructions for this OS, it will be installed.
If instructions are missing, GDF will ask you how to install it and save the answer.`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

var installScript string
var installPackage string
var installProfile string

func init() {
	appCmd.AddCommand(installCmd)
	installCmd.Flags().StringVar(&installPackage, "package", "", "Specify package name manually")
	installCmd.Flags().StringVar(&installScript, "script", "", "Specify custom install script manually")
	installCmd.Flags().StringVarP(&installProfile, "profile", "p", "default", "Profile to add app to")
}

func runInstall(cmd *cobra.Command, args []string) error {
	appName := args[0]
	gdfDir := platform.ConfigDir()
	plat := platform.Detect()
	pkgMgr := packages.ForPlatform(plat)

	// 1. Load or Create App Bundle
	appPath := filepath.Join(gdfDir, "apps", appName+".yaml")
	var bundle *apps.Bundle
	var isNewApp bool

	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		fmt.Printf("App '%s' not found, creating new bundle...\n", appName)
		bundle = &apps.Bundle{
			Name:        appName,
			Description: fmt.Sprintf("App bundle for %s", appName),
		}
		isNewApp = true
	} else {
		var err error
		bundle, err = apps.Load(appPath)
		if err != nil {
			return fmt.Errorf("loading app bundle: %w", err)
		}
	}

	// 2. Check Profile Association
	// We want to ensure the app is in at least one profile, unless user explicitly skips via interactive prompt.
	if err := ensureAppInProfile(gdfDir, appName, installProfile, isNewApp); err != nil {
		return err
	}

	// 3. Check if we know how to install
	if bundle.Package == nil {
		bundle.Package = &apps.Package{}
	}

	pkgName, defined := bundle.Package.ResolveName(pkgMgr.Name())

	// 4. Learning Phase
	if !defined {
		// If manual flags provided, use them
		if installPackage != "" {
			updateBundlePackage(bundle, pkgMgr.Name(), installPackage)
			pkgName = installPackage
			defined = true
			fmt.Printf("✓ Saved package instruction: %s: %s\n", pkgMgr.Name(), pkgName)
		} else {
			if globalYes {
				if pkgMgr.Name() == "none" {
					return withExitCode(
						fmt.Errorf("no package manager detected for '%s'; rerun interactively or provide --package", appName),
						exitCodeNonInteractiveStop,
					)
				}
				updateBundlePackage(bundle, pkgMgr.Name(), appName)
				pkgName = appName
				defined = true
			} else {
				// Interactive Prompt
				fmt.Printf("❓ How do you install '%s' on %s?\n", appName, plat.OS)
				if pkgMgr.Name() != "none" {
					fmt.Printf("   1. Package Manager (%s)\n", pkgMgr.Name())
				}
				fmt.Println("   2. Custom Script (Not implemented in MVP)")
				fmt.Println("   3. Skip (Manual/External)")

				choice, err := readInteractiveLine("Select [1-3]: ")
				if err != nil {
					return err
				}
				choice = strings.TrimSpace(choice)

				switch choice {
				case "1":
					if pkgMgr.Name() == "none" {
						fmt.Println("No package manager detected.")
						return nil
					}
					inputName, err := readInteractiveLine(fmt.Sprintf("Enter package name for %s (default: %s): ", pkgMgr.Name(), appName))
					if err != nil {
						return err
					}
					inputName = strings.TrimSpace(inputName)
					if inputName == "" {
						inputName = appName
					}
					updateBundlePackage(bundle, pkgMgr.Name(), inputName)
					pkgName = inputName
					defined = true

				case "2":
					fmt.Println("Custom script recording not yet implemented.")
					return nil
				case "3":
					fmt.Println("Skipping installation learning.")
					return nil
				default:
					fmt.Println("Invalid choice.")
					return nil
				}
			}
		}

		// Save if we updated
		if defined {
			if err := bundle.Save(appPath); err != nil {
				return fmt.Errorf("saving app bundle: %w", err)
			}
			fmt.Printf("✓ Updated app definition for '%s'\n", appName)
		}
	}

	// 5. Install Phase
	if defined && pkgName != "" {
		fmt.Printf("📦 Installing %s via %s...\n", pkgName, pkgMgr.Name())
		if err := pkgMgr.Install(pkgName); err != nil {
			return fmt.Errorf("installing package: %w", err)
		}
		fmt.Println("✅ Installed successfully")
	} else {
		fmt.Println("⏭️  Skipping installation (unknown method)")
	}

	return nil
}

func ensureAppInProfile(gdfDir, appName, requestedProfile string, isNewApp bool) error {
	// If specific profile requested, add to it
	if requestedProfile != "" {
		return addAppToProfile(gdfDir, requestedProfile, appName)
	}

	// Check if already in any profile
	profilesDir := filepath.Join(gdfDir, "profiles")
	// If profiles dir doesn't exist, we definitely need to add it
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		// continue to prompt
	} else {
		allProfiles, err := config.LoadAllProfiles(profilesDir)
		if err != nil {
			return fmt.Errorf("checking existing profiles: %w", err)
		}
		for _, p := range allProfiles {
			if contains(p.Apps, appName) {
				// Already in a profile
				return nil
			}
		}
	}

	// Not in any profile. Prompt user.
	fmt.Printf("⚠️  App '%s' is not in any profile.\n", appName)

	// List existing profiles
	allProfiles, _ := config.LoadAllProfiles(profilesDir) // reload ignoring error
	var choices []string
	for _, p := range allProfiles {
		choices = append(choices, p.Name)
	}

	fmt.Println("Select a profile to add it to:")
	for i, name := range choices {
		fmt.Printf("   %d. %s\n", i+1, name)
	}
	fmt.Printf("   %d. Create new profile...\n", len(choices)+1)
	fmt.Printf("   %d. Skip (leave orphaned)\n", len(choices)+2)

	input, err := readInteractiveLine("Select: ")
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)

	var selectedProfile string

	// Parse selection
	// TODO: Handle numeric selection robustly. For now assume valid input or fall through.
	// Simplified: if input matches a number, map to choice.

	var selection int
	_, err = fmt.Sscanf(input, "%d", &selection)
	if err == nil {
		if selection > 0 && selection <= len(choices) {
			selectedProfile = choices[selection-1]
		} else if selection == len(choices)+1 {
			// Create new
			newName, err := readInteractiveLine("Enter name for new profile: ")
			if err != nil {
				return err
			}
			selectedProfile = strings.TrimSpace(newName)
			if selectedProfile == "" {
				return fmt.Errorf("profile name cannot be empty")
			}
		} else {
			// Skip
			return nil
		}
	} else {
		// Treat as profile name if not a number? Or just invalid.
		// For MVP, if they type the name exact, use it?
		// Let's stick to prompt logic above.
		fmt.Println("Invalid selection. Skipping profile association.")
		return nil
	}

	return addAppToProfile(gdfDir, selectedProfile, appName)
}

func addAppToProfile(gdfDir, profileName, appName string) error {
	profileDir := filepath.Join(gdfDir, "profiles", profileName)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}

	profilePath := filepath.Join(profileDir, "profile.yaml")
	profile, err := config.LoadProfile(profilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			profile = &config.Profile{Name: profileName}
		} else {
			return fmt.Errorf("loading profile: %w", err)
		}
	}

	if contains(profile.Apps, appName) {
		fmt.Printf("✓ App '%s' is already in profile '%s'\n", appName, profileName)
		return nil
	}

	profile.Apps = append(profile.Apps, appName)
	if err := profile.Save(profilePath); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}

	fmt.Printf("✓ Added '%s' to profile '%s'\n", appName, profileName)
	return nil
}

func updateBundlePackage(bundle *apps.Bundle, mgrName string, pkgName string) {
	switch mgrName {
	case "brew":
		bundle.Package.Brew = pkgName
	case "apt":
		if bundle.Package.Apt == nil {
			bundle.Package.Apt = &apps.AptPackage{}
		}
		bundle.Package.Apt.Name = pkgName
	case "dnf":
		bundle.Package.Dnf = pkgName
	case "pacman":
		bundle.Package.Pacman = pkgName
	}
}
