package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/dan-3200/hyprprofile/internal/app"
	"github.com/dan-3200/hyprprofile/internal/profile"
)

func run(arguments []string, profiles *app.ProfileService, system *app.SystemService) error {
	if len(arguments) == 0 {
		return errors.New("a command is required")
	}
	switch arguments[0] {
	case "list":
		return listProfiles(profiles)
	case "status":
		return showStatus(profiles)
	case "create":
		return createProfile(arguments[1:], profiles)
	case "templates":
		return listTemplates(arguments[1:], profiles)
	case "clone":
		return cloneProfile(arguments[1:], profiles)
	case "import":
		return importProfile(arguments[1:], profiles)
	case "validate":
		return validateProfile(arguments[1:], profiles)
	case "activate":
		return activateProfile(arguments[1:], profiles)
	case "remove":
		return removeProfile(arguments[1:], profiles)
	case "restore":
		return restoreProfile(arguments[1:], profiles)
	case "doctor":
		return doctor(system)
	case "install-session":
		return installSession(arguments[1:], system)
	case "uninstall-session":
		return uninstallSession(arguments[1:], system)
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", arguments[0])
	}
}

func listProfiles(profiles *app.ProfileService) error {
	items, err := profiles.ListProfiles()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	for _, item := range items {
		marker := " "
		if item.Active {
			marker = "*"
		}
		fmt.Printf("%s %-20s %s\n", marker, item.Manifest.DirectoryID(), item.Manifest.Name)
	}
	return nil
}

func showStatus(profiles *app.ProfileService) error {
	item, selected, err := profiles.Status()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if !selected {
		fmt.Println("No active profile")
		return nil
	}
	fmt.Printf("Active profile: %s (%s)\n", item.Name, item.DirectoryID())
	return nil
}

func createProfile(arguments []string, profiles *app.ProfileService) error {
	if len(arguments) == 0 {
		return errors.New("usage: create <id> --name <name> [--template <template>]")
	}
	flags := flag.NewFlagSet("create", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	name := flags.String("name", "", "profile name")
	description := flags.String("description", "", "profile description")
	templateName := flags.String("template", profile.DefaultTemplate, "profile template")
	if err := flags.Parse(arguments[1:]); err != nil {
		return fmt.Errorf("parse create flags: %w", err)
	}
	if flags.NArg() != 0 {
		return errors.New("usage: create <id> --name <name> [--template <template>]")
	}
	manifest, err := profiles.CreateProfile(app.CreateProfileInput{ID: arguments[0], Name: *name, Description: *description, Template: *templateName})
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("Created profile %q from template %q\n", manifest.DirectoryID(), *templateName)
	return nil
}

func listTemplates(arguments []string, profiles *app.ProfileService) error {
	if len(arguments) != 0 {
		return errors.New("usage: templates")
	}
	for _, template := range profiles.Templates() {
		marker := " "
		if template.Name == profile.DefaultTemplate {
			marker = "*"
		}
		fmt.Printf("%s %-12s %s\n", marker, template.Name, template.Description)
	}
	return nil
}

func cloneProfile(arguments []string, profiles *app.ProfileService) error {
	if len(arguments) < 2 {
		return errors.New("usage: clone <source-id> <new-id> --name <name>")
	}
	flags := flag.NewFlagSet("clone", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	name := flags.String("name", "", "profile name")
	if err := flags.Parse(arguments[2:]); err != nil {
		return fmt.Errorf("parse clone flags: %w", err)
	}
	target, err := profiles.CloneProfile(app.CloneProfileInput{SourceID: arguments[0], ID: arguments[1], Name: *name})
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("Cloned %q as %q\n", arguments[0], target.DirectoryID())
	return nil
}
