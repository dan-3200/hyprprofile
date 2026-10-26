package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	app "github.com/dan-3200/hyprprofile/internal/application/usecases"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/inspection"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/session"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/storage"
	"github.com/dan-3200/hyprprofile/internal/infrastructure/templates"
	"github.com/dan-3200/hyprprofile/internal/presentation/cli"
)

func TestCLIProfileFlowAndOutput(t *testing.T) {
	root := t.TempDir()
	paths, err := storage.NewPaths(filepath.Join(root, "config"), filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	store := storage.New(paths)
	profiles := app.NewProfileService(store, store, session.NewManager(paths), inspection.DefaultInspector(), templates.Catalog{}, time.Now)
	system := app.NewSystemService(store, exec.LookPath, session.DesktopEntry{}, session.DefaultLauncherResolver())
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
		output, runErr := capturedOutput(func() error { return cli.Run(tc.arguments, profiles, system, session.DefaultSessionDirectory) })
		if runErr != nil || !strings.HasPrefix(output, tc.want) {
			t.Fatalf("run(%q) = %q, %v; want prefix %q", tc.arguments, output, runErr, tc.want)
		}
	}
	if _, err := runError([]string{"remove", "work"}, profiles, system); err == nil || err.Error() != "cannot remove the active profile; activate another profile first" {
		t.Fatalf("active remove error = %v", err)
	}
}

func TestCLIExitCodeOnMissingProfile(t *testing.T) {
	if os.Getenv("HYPRPROFILE_TEST_CHILD") == "1" {
		os.Args = []string{"hyprprofile", "activate", "missing"}
		main()
		return
	}
	root := t.TempDir()
	// #nosec G702 -- this test reruns its own executable with fixed test arguments.
	command := exec.Command(os.Args[0], "-test.run=^TestCLIExitCodeOnMissingProfile$")
	command.Env = append(os.Environ(),
		"HYPRPROFILE_TEST_CHILD=1",
		"XDG_CONFIG_HOME="+filepath.Join(root, "config"),
		"XDG_STATE_HOME="+filepath.Join(root, "state"),
	)
	output, err := command.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("exit = %v, output = %q", err, output)
	}
	if !strings.Contains(string(output), "hyprprofile: load profile:") {
		t.Fatalf("stderr = %q", output)
	}
}

func runError(arguments []string, profiles *app.ProfileService, system *app.SystemService) (string, error) {
	return capturedOutput(func() error { return cli.Run(arguments, profiles, system, session.DefaultSessionDirectory) })
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
