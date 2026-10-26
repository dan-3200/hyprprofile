package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/app"
	"github.com/dan-3200/hyprprofile/internal/session"
	"github.com/dan-3200/hyprprofile/internal/storage"
)

func TestCLIProfileFlowAndOutput(t *testing.T) {
	root := t.TempDir()
	paths, err := storage.NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	store := storage.New(paths)
	profiles := app.NewProfileService(store, store, session.NewManager(paths), time.Now)
	system := app.DefaultSystemService(store)
	cases := []struct {
		arguments []string
		want      string
	}{
		{[]string{"status"}, "No active profile\n"},
		{[]string{"create", "work", "--name", "Work"}, "Created profile \"work\" from template \"minimal\"\n"},
		{[]string{"validate", "work"}, "Profile \"work\" is valid\n"},
		{[]string{"activate", "work"}, "Activated profile \"work\" for the next login\n"},
		{[]string{"status"}, "Active profile: Work (work)\n"},
		{[]string{"list"}, "* work                 Work\n"},
		{[]string{"clone", "work", "copy"}, "Cloned \"work\" as \"copy\"\n"},
		{[]string{"remove", "copy"}, "Archived profile \"copy\" to "},
		{[]string{"restore", "copy"}, "Restored profile \"copy\"\n"},
	}
	for _, tc := range cases {
		output, runErr := capturedOutput(func() error { return run(tc.arguments, profiles, system) })
		if runErr != nil || !strings.HasPrefix(output, tc.want) {
			t.Fatalf("run(%q) = %q, %v; want prefix %q", tc.arguments, output, runErr, tc.want)
		}
	}
	if _, err := runError([]string{"remove", "work"}, profiles, system); err == nil || err.Error() != "cannot remove the active profile; activate another profile first" {
		t.Fatalf("active remove error = %v", err)
	}
}

func runError(arguments []string, profiles *app.ProfileService, system *app.SystemService) (string, error) {
	return capturedOutput(func() error { return run(arguments, profiles, system) })
}

func capturedOutput(action func() error) (string, error) {
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("create output pipe: %w", err)
	}
	os.Stdout = writer
	defer func() { os.Stdout = original }()
	runErr := action()
	_ = writer.Close()
	output, readErr := io.ReadAll(reader)
	_ = reader.Close()
	if readErr != nil {
		return "", fmt.Errorf("read output: %w", readErr)
	}
	return string(output), runErr
}
