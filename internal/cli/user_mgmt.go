package cli

import (
	"context"
	"fmt"

	"github.com/manifoldco/promptui"

	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

func AddUser(ctx context.Context, cfgPath string, username string) error {
	var err error

	app, err := BootstrapApp(ctx, cfgPath)
	if err != nil {
		return err
	}

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

	newUser, err := app.Store.CreateUser(ctx, username, password)
	if err != nil {
		return err
	}

	fmt.Println("Created new user:", newUser.Username)
	return nil
}

func ModifyUser(ctx context.Context, cfgPath string, username string) error {
	var err error

	app, err := BootstrapApp(ctx, cfgPath)
	if err != nil {
		return err
	}

	var user *db.User
	if username == "" {
		user, err = displayUserSelection(ctx, app.Store, "Select user to modify")
	} else {
		user, err = app.Store.RetrieveUser(ctx, username)
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

func DeleteUser(ctx context.Context, cfgPath string, username string) error {
	var err error

	app, err := BootstrapApp(ctx, cfgPath)
	if err != nil {
		return err
	}

	var user *db.User
	if username == "" {
		user, err = displayUserSelection(ctx, app.Store, "Select user to delete")
	} else {
		user, err = app.Store.RetrieveUser(ctx, username)
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
