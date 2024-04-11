package cli

import (
	"context"
	"fmt"

	"github.com/manifoldco/promptui"

	"github.com/salmanmorshed/simplelinkshortener/internal/cfg"
	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func AddUser(ctx context.Context, cfgPath string, username string) error {
	var err error

	conf, err := cfg.LoadConfigFromFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	store, err := db.NewPgStore(conf)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}
	defer store.Close()

	if username == "" {
		prompt1 := promptui.Prompt{
			Label:    "Username",
			Validate: db.CheckUsernameValidity,
		}
		username, err = prompt1.Run()
		if err != nil {
			return ErrAborted
		}
	} else {
		if err = db.CheckUsernameValidity(username); err != nil {
			return err
		}
		fmt.Println("Username:", username)
	}

	prompt2 := promptui.Prompt{
		Label:    "Password",
		Validate: db.CheckPasswordStrengthValidity,
		Mask:     '*',
	}
	password, err := prompt2.Run()
	if err != nil {
		return ErrAborted
	}

	newUser, err := store.CreateUser(ctx, username, password)
	if err != nil {
		return err
	}

	fmt.Println("Created new user:", newUser.Username)
	return nil
}

func ModifyUser(ctx context.Context, cfgPath string, username string) error {
	var err error

	conf, err := cfg.LoadConfigFromFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	store, err := db.NewPgStore(conf)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}
	defer store.Close()

	var user *db.User
	if username == "" {
		user, err = showUserSelection(ctx, store, "Select user to modify")
	} else {
		user, err = store.RetrieveUser(ctx, username)
	}
	if err != nil {
		return nil
	}

	var toggleAdminLabel string
	if user.IsAdmin {
		toggleAdminLabel = "Revoke admin status"
	} else {
		toggleAdminLabel = "Grant admin status"
	}

	prompt1 := promptui.Select{
		Label: "Operation",
		Items: []string{"Change password", "Change username", toggleAdminLabel},
	}
	_, action, err := prompt1.Run()
	if err != nil {
		return ErrAborted
	}

	if action == "Change password" {
		prompt2 := promptui.Prompt{
			Label:    "New password",
			Validate: db.CheckPasswordStrengthValidity,
			Mask:     '*',
		}
		newPassword, err := prompt2.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}

		if err = store.UpdatePassword(ctx, user.Username, newPassword); err != nil {
			return err
		}

		fmt.Println("Updated password for", user.Username)
	} else if action == "Change username" {
		prompt2 := promptui.Prompt{
			Label:     "New username",
			Validate:  db.CheckUsernameValidity,
			Default:   user.Username,
			AllowEdit: true,
		}
		newUsername, err := prompt2.Run()
		if err != nil {
			fmt.Println("aborted")
			return nil
		}

		oldUsername := user.Username
		if err := store.UpdateUsername(ctx, user.Username, newUsername); err != nil {
			return err
		}

		fmt.Println("Updated username", oldUsername, "to", newUsername)
	} else if action == toggleAdminLabel {
		prompt3 := promptui.Prompt{
			Label:     "Are you sure?",
			IsConfirm: true,
		}
		confirm, err := prompt3.Run()
		if err != nil || (confirm != "y" && confirm != "Y") {
			fmt.Println("aborted")
			return nil
		}

		if err := store.ToggleAdmin(ctx, user.Username); err != nil {
			return err
		}

		fmt.Println("Admin status updated for", user.Username)
	}

	return nil
}

func DeleteUser(ctx context.Context, cfgPath string, username string) error {
	var err error

	conf, err := cfg.LoadConfigFromFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	store, err := db.NewPgStore(conf)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}
	defer store.Close()

	var user *db.User
	if username == "" {
		user, err = showUserSelection(ctx, store, "Select user to delete")
	} else {
		user, err = store.RetrieveUser(ctx, username)
	}
	if err != nil {
		return err
	}

	prompt1 := promptui.Prompt{
		Label:     fmt.Sprintf("Delete user %s", user.Username),
		IsConfirm: true,
	}
	confirm, err := prompt1.Run()
	if err != nil || (confirm != "y" && confirm != "Y") {
		return ErrAborted
	}

	if err = store.DeleteUser(ctx, user.Username); err != nil {
		return fmt.Errorf("failed to delete user %s", user.Username)
	}

	fmt.Println("Deleted user", user.Username)

	return nil
}
