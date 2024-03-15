package main

import (
	"context"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func addUserHandler(CLICtx *cli.Context) error {
	app, err := getApp(CLICtx)
	if err != nil {
		return err
	}

	prompt1 := promptui.Prompt{
		Label:    "Username",
		Validate: db.CheckUsernameValidity,
	}
	username, err1 := prompt1.Run()
	if err1 != nil {
		return Aborted
	}

	prompt2 := promptui.Prompt{
		Label:    "Password",
		Validate: db.CheckPasswordStrengthValidity,
		Mask:     '*',
	}
	password, err2 := prompt2.Run()
	if err2 != nil {
		return Aborted
	}

	newUser, err := app.Store.CreateUser(CLICtx.Context, username, password)
	if err != nil {
		return err
	}

	fmt.Println("Created new user:", newUser.Username)
	return nil
}

func modifyUserHandler(CLICtx *cli.Context) error {
	var err error

	ctx := CLICtx.Context

	app, err := getApp(CLICtx)
	if err != nil {
		return err
	}

	user, err := displayUserSelection(ctx, app.Store, "Select user")
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
		return Aborted
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

		if err = app.Store.UpdatePassword(ctx, user.Username, newPassword); err != nil {
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
		if err := app.Store.UpdateUsername(ctx, user.Username, newUsername); err != nil {
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

		if err := app.Store.ToggleAdmin(ctx, user.Username); err != nil {
			return err
		}

		fmt.Println("Admin status updated for", user.Username)
	}

	return nil
}

func deleteUserHandler(CLICtx *cli.Context) error {
	var err error

	ctx := CLICtx.Context

	app, err := getApp(CLICtx)
	if err != nil {
		return err
	}

	user, err := displayUserSelection(ctx, app.Store, "Select user to delete")
	if err != nil {
		return nil
	}

	prompt1 := promptui.Prompt{
		Label:     fmt.Sprintf("Delete user: %s", user.Username),
		IsConfirm: true,
	}
	confirm, err := prompt1.Run()
	if err != nil || (confirm != "y" && confirm != "Y") {
		return Aborted
	}

	if err = app.Store.DeleteUser(ctx, user.Username); err != nil {
		return fmt.Errorf("failed to delete user %s", user.Username)
	}

	fmt.Println("Deleted user", user.Username)

	return nil
}

func displayUserSelection(ctx context.Context, store db.Store, prompt string) (*db.User, error) {
	users, err := store.RetrieveAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users")
	}

	usernames := make([]string, len(users))
	for idx, user := range users {
		usernames[idx] = user.Username
	}

	prompt1 := promptui.Select{
		Label: prompt,
		Items: usernames,
	}
	_, username, err := prompt1.Run()
	if err != nil {
		return nil, fmt.Errorf("user selection failed")
	}

	user, err := store.RetrieveUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user %s", username)
	}

	return user, nil
}
