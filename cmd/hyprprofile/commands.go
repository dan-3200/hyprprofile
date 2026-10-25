package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dan-3200/hyprprofile/internal/profile"
	"github.com/dan-3200/hyprprofile/internal/session"
	"github.com/dan-3200/hyprprofile/internal/storage"
)

func run(arguments []string, store *storage.Store, manager session.Manager) error {
	if len(arguments) == 0 {
		return errors.New("a command is required")
	}
	switch arguments[0] {
	case "list":
		return listProfiles(store, manager)
	case "status":
		return showStatus(store, manager)
	case "create":
		return createProfile(arguments[1:], store)
	case "clone":
		return cloneProfile(arguments[1:], store)
	case "import":
		return importProfile(arguments[1:], store)
	case "validate":
		return validateProfile(arguments[1:], store)
	case "activate":
		return activateProfile(arguments[1:], store, manager)
	case "remove":
		return removeProfile(arguments[1:], store, manager)
	case "restore":
		return restoreProfile(arguments[1:], store)
	case "doctor":
		return doctor(store)
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", arguments[0])
	}
}

func listProfiles(store *storage.Store, manager session.Manager) error {
	items, err := store.List()
	if err != nil {
		return fmt.Errorf("list profiles: %w", err)
	}
	active, err := manager.Get()
	if err != nil {
		return fmt.Errorf("read active profile: %w", err)
	}
	for _, item := range items {
		marker := " "
		if item.ID == active {
			marker = "*"
		}
		fmt.Printf("%s %-20s %s\n", marker, item.ID, item.Name)
	}
	return nil
}

func showStatus(store *storage.Store, manager session.Manager) error {
	active, err := manager.Get()
	if err != nil {
		return fmt.Errorf("read active profile: %w", err)
	}
	if active == "" {
		fmt.Println("No active profile")
		return nil
	}
	manifest, err := store.Load(active)
	if err != nil {
		return fmt.Errorf("active profile %q is unavailable: %w", active, err)
	}
	fmt.Printf("Active profile: %s (%s)\n", manifest.Name, manifest.ID)
	return nil
}

func createProfile(arguments []string, store *storage.Store) error {
	if len(arguments) == 0 {
		return errors.New("usage: create <id> --name <name>")
	}
	flags := flag.NewFlagSet("create", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	name := flags.String("name", "", "profile name")
	description := flags.String("description", "", "profile description")
	entrypoint := flags.String("entrypoint", profile.DefaultEntrypoint, "relative entrypoint")
	if err := flags.Parse(arguments[1:]); err != nil {
		return fmt.Errorf("parse create flags: %w", err)
	}
	manifest := profile.NewManifest(arguments[0], *name, *description, *entrypoint, time.Now())
	if err := store.Create(manifest); err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	fmt.Printf("Created profile %q\n", manifest.ID)
	return nil
}

func cloneProfile(arguments []string, store *storage.Store) error {
	if len(arguments) < 2 {
		return errors.New("usage: clone <source-id> <new-id> --name <name>")
	}
	flags := flag.NewFlagSet("clone", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	name := flags.String("name", "", "profile name")
	if err := flags.Parse(arguments[2:]); err != nil {
		return fmt.Errorf("parse clone flags: %w", err)
	}
	source, err := store.Load(arguments[0])
	if err != nil {
		return fmt.Errorf("load source profile: %w", err)
	}
	if *name == "" {
		*name = source.Name + " (clone)"
	}
	target := source
	target.ID = arguments[1]
	target.Name = *name
	target.CreatedAt = time.Now().UTC()
	target.UpdatedAt = target.CreatedAt
	target.Origin = source.ID
	if err := store.Clone(source.ID, target); err != nil {
		return fmt.Errorf("clone profile: %w", err)
	}
	fmt.Printf("Cloned %q as %q\n", source.ID, target.ID)
	return nil
}
