package main

import (
	"fmt"

	"github.com/manifoldco/promptui"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/database"
	"github.com/salmanmorshed/simplelinkshortener/internal/utils"
)

func addUser(_ *config.Config, store database.Store) error {
	prompt1 := promptui.Prompt{
		Label:    "Username",
		Validate: utils.CheckUsernameValidity,
	}
	username, err1 := prompt1.Run()
	if err1 != nil {
		fmt.Println("aborted")
		return nil
	}

	prompt2 := promptui.Prompt{
		Label:    "Password",
		Validate: utils.CheckPasswordStrengthValidity,
		Mask:     '*',
	}
	password, err2 := prompt2.Run()
	if err2 != nil {
		fmt.Println("aborted")
		return nil
	}

	newUser, err := store.CreateUser(username, password)
	if err != nil {
		return err
	}

	fmt.Println("Created new user:", newUser.Username)
	return nil
}

func displayUserSelection(store database.Store, promptMessage string) (*database.User, error) {
	users, err := store.RetrieveAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users")
	}

	usernames := make([]string, len(users))
	for idx, user := range users {
		usernames[idx] = user.Username
	}

	prompt1 := promptui.Select{
		Label: promptMessage,
		Items: usernames,
	}
	_, username, err := prompt1.Run()
	if err != nil {
		fmt.Println("aborted")
		return nil, fmt.Errorf("user selection failed")
	}

	user, err := store.RetrieveUser(username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user %s", username)
	}

	return user, nil
}

func modifyUser(_ *config.Config, store database.Store) error {
	user, err := displayUserSelection(store, "Select user")
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
		fmt.Println("aborted")
		return nil
	}

	if action == "Change password" {
		prompt2 := promptui.Prompt{
			Label:    "New password",
			Validate: utils.CheckPasswordStrengthValidity,
			Mask:     '*',
		}
		newPassword, err2 := prompt2.Run()
		if err2 != nil {
			fmt.Println("aborted")
			return nil
		}

		if err := store.UpdatePassword(user.Username, newPassword); err != nil {
			return err
		}

		fmt.Println("Updated password for", user.Username)
	} else if action == "Change username" {
		prompt2 := promptui.Prompt{
			Label:     "New username",
			Validate:  utils.CheckUsernameValidity,
			Default:   user.Username,
			AllowEdit: true,
		}
		newUsername, err2 := prompt2.Run()
		if err2 != nil {
			fmt.Println("aborted")
			return nil
		}

		oldUsername := user.Username
		if err := store.UpdateUsername(user.Username, newUsername); err != nil {
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

		if err := store.ToggleAdmin(user.Username); err != nil {
			return err
		}

		fmt.Println("Admin status updated for", user.Username)
	}

	return nil
}

func deleteUser(_ *config.Config, store database.Store) error {
	user, err := displayUserSelection(store, "Select user to delete")
	if err != nil {
		return nil
	}

	prompt1 := promptui.Prompt{
		Label:     fmt.Sprintf("Delete user: %s", user.Username),
		IsConfirm: true,
	}
	confirm, err := prompt1.Run()
	if err != nil || (confirm != "y" && confirm != "Y") {
		fmt.Println("aborted")
		return nil
	}

	if err := store.DeleteUser(user.Username); err != nil {
		return fmt.Errorf("failed to delete user %s", user.Username)
	}

	fmt.Println("Deleted user", user.Username)

	return nil
}
