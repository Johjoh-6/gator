package main

import (
	"context"
	"database/sql"
	"fmt"
	"gator/internal/database"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username required")
	}
	userName := cmd.args[0]
	user, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("username not found")
		}
		return err
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("Logged as %s\n", user.Name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username required")
	}
	userName := cmd.args[0]

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		Name:      userName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("username already taken")
		}
		return fmt.Errorf("couldn't create user: %w", err)
	}
	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User created, logged in as %s\n", user.Name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't reset users: %w", err)
	}

	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't get users: %w", err)
	}
	// print all users name.
	// If user is current add (current) after their name
	for _, user := range users {
		textUser := fmt.Sprintf("* %s", user.Name)
		if user.Name == s.config.CurrentUserName {
			textUser += " (current)"
		}
		fmt.Println(textUser)
	}
	return nil
}

func handlerAggregator(s *state, cmd command) error {
	timeBetweenReqs := 1 * time.Minute

	if len(cmd.args) > 0 {
		input := cmd.args[0]
		// split input to take number and unit can be s, m or h
		parts := strings.Split(input, "")
		if len(parts) != 2 {
			return fmt.Errorf("invalid time interval: %s", cmd.args[0])
		}
		num := parts[0]
		unit := parts[1]

		// convert num to int
		numInt, err := strconv.Atoi(num)
		if err != nil {
			return fmt.Errorf("invalid time interval: %s", cmd.args[0])
		}
		// convert unit to time.Duration
		switch unit {
		case "s":
			timeBetweenReqs = time.Duration(numInt) * time.Second
		case "m":
			timeBetweenReqs = time.Duration(numInt) * time.Minute
		case "h":
			timeBetweenReqs = time.Duration(numInt) * time.Hour
		default:
			return fmt.Errorf("invalid time interval: %s", cmd.args[0])
		}
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("feed name and URL required")
	}
	name := cmd.args[0]
	url := cmd.args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't add feed: %w", err)
	}

	fmt.Printf("Added feed: %v\n", feed)

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		FeedID: feed.ID,
		UserID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't follow feed: %w", err)
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't get feeds: %w", err)
	}
	// print all feeds name.
	for _, feed := range feeds {
		fmt.Printf("* %s (%s) from %s\n", feed.Name, feed.Url, feed.Username)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("follow requires a feed URL")
	}
	feedURL := cmd.args[0]

	// get feed by URL
	feed, err := s.db.GetFeedByUrl(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("couldn't get feed: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}
	fmt.Printf("You follow %s added by %s\n", feedFollow.FeedName, feedFollow.UserName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("couldn't get feed follows: %w", err)
	}
	for _, feedFollow := range feedFollows {
		fmt.Printf("* %s\n", feedFollow.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("follow requires a feed URL")
	}
	feedURL := cmd.args[0]

	// get feed by URL
	feed, err := s.db.GetFeedByUrl(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("couldn't get feed: %w", err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't delete feed follow: %w", err)
	}

	return nil
}
