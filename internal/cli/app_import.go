package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rztaylor/GoDotFiles/internal/apps"
	"github.com/rztaylor/GoDotFiles/internal/platform"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [paths...]",
	Short: "Discover and adopt existing dotfiles, aliases, and common configs",
	Long: `Discover existing dotfiles and aliases on this machine and adopt them into GDF.

Modes:
  - preview-only: show what would be imported
  - guided mapping: interactively choose app and secret handling (default)
  - apply: import directly using defaults/flags`,
	RunE: runAppImport,
}

var (
	importPreview           bool
	importApply             bool
	importJSON              bool
	importProfile           string
	importSensitiveHandling string
)

type importDotfileCandidate struct {
	Path      string `json:"path"`
	App       string `json:"app"`
	Sensitive bool   `json:"sensitive"`
}

type importAliasCandidate struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	App     string `json:"app"`
	Source  string `json:"source"`
}

type importPreviewOutput struct {
	Mode     string                   `json:"mode"`
	Dotfiles []importDotfileCandidate `json:"dotfiles"`
	Aliases  []importAliasCandidate   `json:"aliases"`
}

func init() {
	appCmd.AddCommand(importCmd)
	importCmd.Flags().BoolVar(&importPreview, "preview", false, "Preview discoveries without importing")
	importCmd.Flags().BoolVar(&importApply, "apply", false, "Apply import directly without guided mapping")
	importCmd.Flags().BoolVar(&importJSON, "json", false, "Output discovery/import data as JSON")
	importCmd.Flags().StringVarP(&importProfile, "profile", "p", "", "Profile to add imported apps to")
	importCmd.Flags().StringVar(&importSensitiveHandling, "sensitive-handling", "", "Default handling for sensitive files in apply mode: ignore|secret|plain")
}

func runAppImport(cmd *cobra.Command, args []string) error {
	if importPreview && importApply {
		return fmt.Errorf("--preview and --apply cannot be used together")
	}

	mode := "guided"
	if importPreview {
		mode = "preview"
	} else if importApply {
		mode = "apply"
	}

	if mode == "guided" && globalNonInteractive {
		return withExitCode(fmt.Errorf("guided import requires interactive input; use --preview or --apply"), exitCodeNonInteractiveStop)
	}

	profileName, err := resolveProfileSelectionForCommand(platform.ConfigDir(), importProfile, "gdf app import")
	if err != nil {
		return err
	}

	home := platform.Detect().Home
	dotfiles, aliases, err := discoverImportCandidates(home, args)
	if err != nil {
		return err
	}

	if mode == "preview" {
		return printImportPreview(mode, dotfiles, aliases)
	}

	audit := newDecisionAudit("gdf app import", false)
	importedDotfiles := make([]importDotfileCandidate, 0, len(dotfiles))
	importedAliases := make([]importAliasCandidate, 0, len(aliases))

	for _, candidate := range dotfiles {
		selectedApp := candidate.App
		secret := candidate.Sensitive
		include := true

		if mode == "guided" {
			fmt.Printf("\nDotfile: %s\n", candidate.Path)
			fmt.Printf("  Suggested app: %s\n", selectedApp)
			ok, err := confirmPromptDefaultYes("Import this file? [Y/n]: ")
			if err != nil {
				return err
			}
			if !ok {
				include = false
				audit.Record(candidate.Path, "guided-import", "skip")
			} else {
				appInput, err := readInteractiveLine(fmt.Sprintf("Map to app [%s]: ", selectedApp))
				if err != nil {
					return err
				}
				appInput = strings.TrimSpace(appInput)
				if appInput != "" {
					selectedApp = AppName(appInput)
					audit.Record(candidate.Path, "app-mapping", selectedApp)
				}
			}

			if include && candidate.Sensitive {
				choice, err := chooseTrackConflictDecision(candidate.Path, "potential secret file detected", []string{"secret", "plain", "ignore"})
				if err != nil {
					return err
				}
				audit.Record(candidate.Path, "sensitive-handling", choice)
				switch choice {
				case "secret":
					secret = true
				case "plain":
					secret = false
				case "ignore":
					include = false
				}
			}
		} else if mode == "apply" && candidate.Sensitive {
			switch importSensitiveHandling {
			case "secret":
				secret = true
				audit.Record(candidate.Path, "sensitive-handling", "secret")
			case "plain":
				secret = false
				audit.Record(candidate.Path, "sensitive-handling", "plain")
			case "ignore":
				include = false
				audit.Record(candidate.Path, "sensitive-handling", "ignore")
			default:
				return fmt.Errorf("sensitive file detected (%s): set --sensitive-handling to ignore|secret|plain", candidate.Path)
			}
		}

		if !include {
			continue
		}

		result, err := trackFile(candidate.Path, trackFileOptions{
			AppName:     selectedApp,
			Secret:      secret,
			Interactive: mode == "guided",
			Audit:       audit,
		})
		if err != nil {
			return err
		}
		if result.Skipped {
			continue
		}

		if err := addAppToProfile(platform.ConfigDir(), profileName, selectedApp); err != nil {
			return err
		}

		importedDotfiles = append(importedDotfiles, importDotfileCandidate{Path: candidate.Path, App: selectedApp, Sensitive: secret})
	}

	for _, alias := range aliases {
		if mode == "guided" {
			fmt.Printf("\nAlias: %s=\"%s\" (from %s)\n", alias.Name, alias.Command, alias.Source)
			ok, err := confirmPromptDefaultYes("Import this alias? [Y/n]: ")
			if err != nil {
				return err
			}
			if !ok {
				audit.Record(alias.Name, "alias-import", "skip")
				continue
			}
		}
		if err := importAliasCandidateEntry(alias); err != nil {
			return err
		}
		importedAliases = append(importedAliases, alias)
	}

	if logPath, err := audit.Save(platform.ConfigDir()); err != nil {
		return err
	} else if logPath != "" {
		fmt.Printf("Logged import decisions: %s\n", logPath)
	}

	if importJSON {
		payload := importPreviewOutput{Mode: mode, Dotfiles: importedDotfiles, Aliases: importedAliases}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Imported %d dotfile(s) and %d alias(es).\n", len(importedDotfiles), len(importedAliases))
	return nil
}

func printImportPreview(mode string, dotfiles []importDotfileCandidate, aliases []importAliasCandidate) error {
	if importJSON {
		payload := importPreviewOutput{Mode: mode, Dotfiles: dotfiles, Aliases: aliases}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Discovered %d dotfile(s) and %d alias(es).\n", len(dotfiles), len(aliases))
	for _, c := range dotfiles {
		note := ""
		if c.Sensitive {
			note = " [sensitive]"
		}
		fmt.Printf("  dotfile: %s -> app=%s%s\n", c.Path, c.App, note)
	}
	for _, a := range aliases {
		fmt.Printf("  alias: %s=\"%s\" (app=%s, source=%s)\n", a.Name, a.Command, a.App, a.Source)
	}
	return nil
}

func discoverImportCandidates(home string, extraPaths []string) ([]importDotfileCandidate, []importAliasCandidate, error) {
	knownFiles := []string{
		".gitconfig",
		".zshrc",
		".bashrc",
		".bash_profile",
		".vimrc",
		".tmux.conf",
		filepath.Join(".config", "nvim", "init.vim"),
		filepath.Join(".config", "starship.toml"),
		filepath.Join(".aws", "credentials"),
	}

	seen := map[string]bool{}
	paths := make([]string, 0, len(knownFiles)+len(extraPaths))
	for _, rel := range knownFiles {
		paths = append(paths, filepath.Join(home, rel))
	}
	for _, p := range extraPaths {
		expanded := platform.ExpandPath(p)
		if !filepath.IsAbs(expanded) {
			expanded = filepath.Join(home, expanded)
		}
		paths = append(paths, expanded)
	}

	dotfiles := make([]importDotfileCandidate, 0)
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		info, err := os.Stat(p)
		if err != nil || info.IsDir() {
			continue
		}
		dotfiles = append(dotfiles, importDotfileCandidate{Path: p, App: apps.DetectAppFromPath(p), Sensitive: looksSensitivePath(p)})
	}
	sort.Slice(dotfiles, func(i, j int) bool { return dotfiles[i].Path < dotfiles[j].Path })

	aliases, err := discoverAliases(home)
	if err != nil {
		return nil, nil, err
	}
	return dotfiles, aliases, nil
}

func discoverAliases(home string) ([]importAliasCandidate, error) {
	files := []string{
		filepath.Join(home, ".aliases"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
	}
	aliases := make([]importAliasCandidate, 0)
	seen := map[string]bool{}

	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("opening alias source %s: %w", path, err)
		}
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := strings.TrimSpace(s.Text())
			if !strings.HasPrefix(line, "alias ") {
				continue
			}
			rest := strings.TrimSpace(strings.TrimPrefix(line, "alias "))
			eq := strings.Index(rest, "=")
			if eq <= 0 {
				continue
			}
			name := strings.TrimSpace(rest[:eq])
			cmd := strings.TrimSpace(rest[eq+1:])
			cmd = strings.Trim(cmd, "'\"")
			if name == "" || cmd == "" {
				continue
			}
			key := name + "=" + cmd
			if seen[key] {
				continue
			}
			seen[key] = true
			aliases = append(aliases, importAliasCandidate{
				Name:    name,
				Command: cmd,
				App:     apps.DetectAppFromCommand(cmd),
				Source:  path,
			})
		}
		if err := s.Err(); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("reading alias source %s: %w", path, err)
		}
		_ = f.Close()
	}

	sort.Slice(aliases, func(i, j int) bool {
		if aliases[i].Name == aliases[j].Name {
			return aliases[i].Command < aliases[j].Command
		}
		return aliases[i].Name < aliases[j].Name
	})
	return aliases, nil
}

func looksSensitivePath(path string) bool {
	lower := filepath.ToSlash(strings.ToLower(path))
	base := filepath.Base(lower)

	switch base {
	case "credentials", "id_rsa", "id_ed25519", ".env", ".env.local", ".env.production":
		return true
	}
	if strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") {
		return true
	}
	if strings.Contains(lower, "/.aws/") && base == "credentials" {
		return true
	}
	return false
}

func importAliasCandidateEntry(candidate importAliasCandidate) error {
	oldAliasApp := aliasApp
	defer func() { aliasApp = oldAliasApp }()

	if candidate.App != "" && candidate.App != "unknown" {
		aliasApp = candidate.App
	} else {
		aliasApp = ""
	}
	return runAliasAdd(nil, []string{candidate.Name, candidate.Command})
}
