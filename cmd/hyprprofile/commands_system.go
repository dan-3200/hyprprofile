package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dan-3200/hyprprofile/internal/session"
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

func installSession(arguments []string) error {
	flags := flag.NewFlagSet("install-session", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	directory := flags.String("directory", session.DefaultSessionDirectory, "Wayland session directory")
	launcher := flags.String("launcher", "", "absolute launcher path")
	if err := flags.Parse(arguments); err != nil {
		return fmt.Errorf("parse install-session flags: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("usage: install-session [--launcher <absolute-path>] [--directory <absolute-path>]")
	}
	resolved, err := resolveSessionLauncher(*launcher)
	if err != nil {
		return err
	}
	destination, err := session.InstallDesktopEntry(*directory, resolved)
	if err != nil {
		return fmt.Errorf("install Wayland session: %w", err)
	}
	fmt.Printf("Installed Wayland session at %s\n", destination)
	return nil
}

func uninstallSession(arguments []string) error {
	flags := flag.NewFlagSet("uninstall-session", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	directory := flags.String("directory", session.DefaultSessionDirectory, "Wayland session directory")
	if err := flags.Parse(arguments); err != nil {
		return fmt.Errorf("parse uninstall-session flags: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("usage: uninstall-session [--directory <absolute-path>]")
	}
	destination, err := session.RemoveDesktopEntry(*directory)
	if err != nil {
		return fmt.Errorf("uninstall Wayland session: %w", err)
	}
	fmt.Printf("Removed Wayland session at %s\n", destination)
	return nil
}

func resolveSessionLauncher(explicit string) (string, error) {
	if explicit != "" {
		if !filepath.IsAbs(explicit) {
			return "", errors.New("launcher path must be absolute")
		}
		return filepath.Clean(explicit), nil
	}
	current, err := os.Executable()
	if err == nil {
		sibling := filepath.Join(filepath.Dir(current), "hyprprofile-session-launcher")
		if info, statErr := os.Stat(sibling); statErr == nil && info.Mode().IsRegular() && info.Mode()&0o111 != 0 {
			return sibling, nil
		}
	}
	launcher, err := exec.LookPath("hyprprofile-session-launcher")
	if err != nil {
		return "", errors.New("hyprprofile-session-launcher was not found; install it or use --launcher")
	}
	launcher, err = filepath.Abs(launcher)
	if err != nil {
		return "", fmt.Errorf("resolve launcher path: %w", err)
	}
	return launcher, nil
}

func printUsage() {
	fmt.Print(`HyprProfile manages self-contained Hyprland profiles.

Usage:
  hyprprofile [command]

Commands:
  list
  status
  create <id> --name <name> [--template <template>]
  templates
  clone <source-id> <new-id> [--name <name>]
  import <absolute-directory> <id> [--name <name>]
  validate <id>
  activate <id>
  remove <id>
  restore <id>
  doctor
  install-session [--launcher <absolute-path>] [--directory <absolute-path>]
  uninstall-session [--directory <absolute-path>]
`)
}
