package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = root.Execute()

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	return string(out), err
}

func TestLibraryList(t *testing.T) {
	// Reset commands for testing to avoid side effects
	libraryListCmd.SetArgs([]string{})

	// Capture os.Stdout since the command writes directly to it via tabwriter

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runLibraryList(libraryListCmd, []string{})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r) // ignoring error for brevity in example
	output := string(out)

	assert.NoError(t, err)
	assert.Contains(t, output, "RECIPE")
	assert.Contains(t, output, "git") // git should be in the embedded recipes
}

func TestLibraryDescribe(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runLibraryDescribe(libraryDescribeCmd, []string{"git"})

	w.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(r)
	output := string(out)

	assert.NoError(t, err)
	assert.Contains(t, output, "name: git")
	assert.Contains(t, output, "kind: Recipe/v1")
}

func TestLibraryDescribe_NotFound(t *testing.T) {
	err := runLibraryDescribe(libraryDescribeCmd, []string{"nonexistent-recipe-xyz"})
	assert.Error(t, err)
}
