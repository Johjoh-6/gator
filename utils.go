package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"gator/internal/database"
	"gator/internal/models"
	"html"
	"io"
	"net/http"
	"time"
)

func fetchFeed(ctx context.Context, feedURL string) (*models.RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	feed := &models.RSSFeed{}
	err = xml.Unmarshal(data, feed)
	if err != nil {
		return nil, err
	}

	// clean up
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, item := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(item.Title)
		feed.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}

	return feed, nil
}

func scrapeFeeds(s *state) {
	// get next feed to fetch
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Errorf("failed to get next feed to fetch: %w", err)
		return
	}

	// mark as fetched
	updated := time.Now()
	if err := s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: updated,
		ID:        feed.ID,
	}); err != nil {
		fmt.Errorf("failed to mark feed as fetched: %w", err)
		return
	}
	// fetch feed
	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Errorf("failed to fetch feed: %w", err)
		return
	}

	// print in the console
	fmt.Printf("Found feed: %s\n", feed.Name)
	for _, item := range rssFeed.Channel.Item {
		fmt.Printf("* %s - published at %s \n", item.Title, item.PubDate)
		fmt.Printf("%s\n", item.Link)
		fmt.Printf(" %s\n", item.Description)
	}
	fmt.Println("========================================")
}
