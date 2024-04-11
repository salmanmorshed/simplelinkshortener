package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"

	"github.com/salmanmorshed/simplelinkshortener/internal/db"
)

var ErrAborted = errors.New("aborted")

func showUserSelection(ctx context.Context, store db.Store, prompt string) (*db.User, error) {
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
