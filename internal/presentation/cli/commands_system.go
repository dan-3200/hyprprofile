package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"

	systemapp "github.com/dan-3200/hyprprofile/internal/application/system"
)

func doctor(system *systemapp.Service) error {
	report, err := system.Doctor()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	for _, command := range report {
		if !command.Available {
			fmt.Printf("missing  %s\n", command.Name)
			continue
		}
		fmt.Printf("available %s\n", command.Name)
	}
	return nil
}

func installSession(arguments []string, system *systemapp.Service, sessionDirectory string) error {
	flags := flag.NewFlagSet("install-session", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	directory := flags.String("directory", sessionDirectory, "Wayland session directory")
	launcher := flags.String("launcher", "", "absolute launcher path")
	if err := flags.Parse(arguments); err != nil {
		return fmt.Errorf("parse install-session flags: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("usage: install-session [--launcher <absolute-path>] [--directory <absolute-path>]")
	}
	destination, err := system.InstallSession(*directory, *launcher)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("Installed Wayland session at %s\n", destination)
	return nil
}

func uninstallSession(arguments []string, system *systemapp.Service, sessionDirectory string) error {
	flags := flag.NewFlagSet("uninstall-session", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	directory := flags.String("directory", sessionDirectory, "Wayland session directory")
	if err := flags.Parse(arguments); err != nil {
		return fmt.Errorf("parse uninstall-session flags: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("usage: uninstall-session [--directory <absolute-path>]")
	}
	destination, err := system.UninstallSession(*directory)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("Removed Wayland session at %s\n", destination)
	return nil
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
