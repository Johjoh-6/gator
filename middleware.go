package main

import (
	"context"
	"fmt"
	"gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	// return new function
	return func(s *state, cmd command) error {
		// get current user
		username := s.config.CurrentUserName
		if username == "" {
			return fmt.Errorf("couldn't get current username")
		}
		user, err := s.db.GetUser(context.Background(), username)
		if err != nil {
			return fmt.Errorf("couldn't get user: %w", err)
		}
		// return error if the users is not found

		// return the handler if the user is found
		return handler(s, cmd, user)
	}
}
