package main

import (
	"fmt"
	"os/exec"

	"github.com/dan-3200/hyprprofile/internal/storage"
)

func doctor(store *storage.Store) error {
	if err := store.Ensure(); err != nil {
		return fmt.Errorf("prepare profile storage: %w", err)
	}
	for _, command := range []string{"Hyprland", "hyprctl", "start-hyprland"} {
		if _, err := exec.LookPath(command); err != nil {
			fmt.Printf("missing  %s\n", command)
			continue
		}
		fmt.Printf("available %s\n", command)
	}
	return nil
}

func printUsage() {
	fmt.Print(`HyprProfile manages self-contained Hyprland profiles.

Usage:
  hyprprofile [command]

Commands:
  list
  status
  create <id> --name <name>
  clone <source-id> <new-id> [--name <name>]
  import <absolute-directory> <id> [--name <name>]
  validate <id>
  activate <id>
  remove <id>
  restore <id>
  doctor
`)
}
