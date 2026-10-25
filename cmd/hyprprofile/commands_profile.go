package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dan-3200/hyprprofile/internal/profile"
	"github.com/dan-3200/hyprprofile/internal/session"
	"github.com/dan-3200/hyprprofile/internal/storage"
	"github.com/dan-3200/hyprprofile/internal/validation"
)

func importProfile(arguments []string, store *storage.Store) error {
	if len(arguments) < 2 {
		return errors.New("usage: import <absolute-directory> <id> --name <name>")
	}
	flags := flag.NewFlagSet("import", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	name := flags.String("name", "", "profile name")
	description := flags.String("description", "", "profile description")
	entrypoint := flags.String("entrypoint", "", "relative entrypoint, detected when omitted")
	if err := flags.Parse(arguments[2:]); err != nil {
		return fmt.Errorf("parse import flags: %w", err)
	}
	if !filepath.IsAbs(arguments[0]) {
		return errors.New("import source must be an absolute directory")
	}
	if *name == "" {
		*name = arguments[1]
	}
	manifest := profile.NewManifest(arguments[1], *name, *description, *entrypoint, time.Now())
	if *entrypoint == "" {
		manifest.Entrypoint = ""
	}
	manifest.Origin = arguments[0]
	if err := store.Import(arguments[0], manifest); err != nil {
		return fmt.Errorf("import profile: %w", err)
	}
	fmt.Printf("Imported profile %q\n", manifest.ID)
	return nil
}

func validateProfile(arguments []string, store *storage.Store) error {
	if len(arguments) != 1 {
		return errors.New("usage: validate <id>")
	}
	manifest, err := store.Load(arguments[0])
	if err != nil {
		return fmt.Errorf("load profile: %w", err)
	}
	directory, err := store.ProfilePath(manifest.ID)
	if err != nil {
		return fmt.Errorf("resolve profile path: %w", err)
	}
	result := validation.ValidateProfile(directory, manifest, true)
	if result.Valid() {
		fmt.Printf("Profile %q is valid\n", manifest.ID)
	}
	for _, issue := range result.Issues {
		fmt.Printf("%s: %s\n", issue.Severity, issue.Message)
	}
	if !result.Valid() {
		return errors.New("profile validation failed")
	}
	return nil
}

func activateProfile(arguments []string, store *storage.Store, manager session.Manager) error {
	if len(arguments) != 1 {
		return errors.New("usage: activate <id>")
	}
	manifest, err := store.Load(arguments[0])
	if err != nil {
		return fmt.Errorf("load profile: %w", err)
	}
	directory, err := store.ProfilePath(manifest.ID)
	if err != nil {
		return fmt.Errorf("resolve profile path: %w", err)
	}
	if result := validation.ValidateProfile(directory, manifest, true); !result.Valid() {
		return errors.New("cannot activate an invalid profile; run validate for details")
	}
	if err := manager.Activate(store, manifest.ID); err != nil {
		return fmt.Errorf("activate profile: %w", err)
	}
	fmt.Printf("Activated profile %q for the next login\n", manifest.ID)
	return nil
}

func removeProfile(arguments []string, store *storage.Store, manager session.Manager) error {
	if len(arguments) != 1 {
		return errors.New("usage: remove <id>")
	}
	active, err := manager.Get()
	if err != nil {
		return fmt.Errorf("read active profile: %w", err)
	}
	if arguments[0] == active {
		return errors.New("cannot remove the active profile; activate another profile first")
	}
	destination, err := store.Archive(arguments[0])
	if err != nil {
		return fmt.Errorf("archive profile: %w", err)
	}
	fmt.Printf("Archived profile %q to %s\n", arguments[0], destination)
	return nil
}

func restoreProfile(arguments []string, store *storage.Store) error {
	if len(arguments) != 1 {
		return errors.New("usage: restore <id>")
	}
	if err := store.Restore(arguments[0]); err != nil {
		return fmt.Errorf("restore profile: %w", err)
	}
	fmt.Printf("Restored profile %q\n", arguments[0])
	return nil
}
