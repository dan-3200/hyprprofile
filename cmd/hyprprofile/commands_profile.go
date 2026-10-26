package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dan-3200/hyprprofile/internal/app"
)

func importProfile(arguments []string, profiles *app.ProfileService) error {
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
	manifest, err := profiles.ImportProfile(app.ImportProfileInput{
		Source: arguments[0], ID: arguments[1], Name: *name, Description: *description, Entrypoint: *entrypoint,
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("Imported profile %q\n", manifest.DirectoryID())
	return nil
}

func validateProfile(arguments []string, profiles *app.ProfileService) error {
	if len(arguments) != 1 {
		return errors.New("usage: validate <id>")
	}
	result, err := profiles.ValidateProfile(arguments[0])
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	if result.Valid() {
		fmt.Printf("Profile %q is valid\n", arguments[0])
	}
	for _, issue := range result.Issues {
		fmt.Printf("%s: %s\n", issue.Severity, issue.Message)
	}
	if !result.Valid() {
		return errors.New("profile validation failed")
	}
	return nil
}

func activateProfile(arguments []string, profiles *app.ProfileService) error {
	if len(arguments) != 1 {
		return errors.New("usage: activate <id>")
	}
	manifest, err := profiles.ActivateProfile(arguments[0])
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("Activated profile %q for the next login\n", manifest.DirectoryID())
	return nil
}

func removeProfile(arguments []string, profiles *app.ProfileService) error {
	if len(arguments) != 1 {
		return errors.New("usage: remove <id>")
	}
	result, err := profiles.ArchiveProfile(arguments[0])
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("Archived profile %q to %s\n", result.ID, result.Destination)
	return nil
}

func restoreProfile(arguments []string, profiles *app.ProfileService) error {
	if len(arguments) != 1 {
		return errors.New("usage: restore <id>")
	}
	if err := profiles.RestoreProfile(arguments[0]); err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Printf("Restored profile %q\n", arguments[0])
	return nil
}
