package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/rztaylor/GoDotFiles/internal/library"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var libraryCmd = &cobra.Command{
	Use:     "library",
	Aliases: []string{"lib"},
	Short:   "Manage and explore the app library",
	Long:    `Explore the built-in app library containing recipes for common tools.`,
}

var libraryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available recipes",
	RunE:  runLibraryList,
}

var libraryDescribeCmd = &cobra.Command{
	Use:   "describe <recipe>",
	Short: "Show details of a recipe",
	Args:  cobra.ExactArgs(1),
	RunE:  runLibraryDescribe,
}

func init() {
	libraryCmd.AddCommand(libraryListCmd)
	libraryCmd.AddCommand(libraryDescribeCmd)
}

func runLibraryList(cmd *cobra.Command, args []string) error {
	mgr := library.New()
	recipes, err := mgr.List()
	if err != nil {
		return fmt.Errorf("listing recipes: %w", err)
	}

	if len(recipes) == 0 {
		fmt.Println("No recipes found.")
		return nil
	}

	sort.Strings(recipes)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "RECIPE")
	for _, name := range recipes {
		fmt.Fprintf(w, "%s\n", name)
	}
	w.Flush()
	return nil
}

func runLibraryDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	mgr := library.New()

	recipe, err := mgr.Get(name)
	if err != nil {
		// Try fuzzy matching or helpful error? For now, strict.
		return fmt.Errorf("getting recipe: %w", err)
	}

	// Encode recipe to YAML for display
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	if err := encoder.Encode(recipe); err != nil {
		return fmt.Errorf("encoding recipe: %w", err)
	}

	// Add a helpful note if dependencies or companions exist
	if len(recipe.Dependencies) > 0 {
		fmt.Printf("\n# Dependencies: %s\n", strings.Join(recipe.Dependencies, ", "))
	}
	if len(recipe.Companions) > 0 {
		fmt.Printf("# Companions: %s\n", strings.Join(recipe.Companions, ", "))
	}

	return nil
}
