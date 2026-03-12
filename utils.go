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
	"strings"
	"time"

	"github.com/google/uuid"
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
	// save the all the feed item as posts in the database
	for _, itemFeed := range rssFeed.Channel.Item {

		// Convert Description to sql.NullString
		description := sql.NullString{
			String: itemFeed.Description,
			Valid:  itemFeed.Description != "",
		}

		// Convert PubDate to sql.NullTime
		var publishedAt sql.NullTime
		if itemFeed.PubDate != "" {
			t, err := time.Parse(time.RFC1123, itemFeed.PubDate)
			if err == nil {
				publishedAt = sql.NullTime{
					Time:  t,
					Valid: true,
				}
			} else {
				publishedAt = sql.NullTime{Valid: false}
			}
		} else {
			publishedAt = sql.NullTime{Valid: false}
		}

		if _, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			FeedID:      feed.ID,
			Title:       itemFeed.Title,
			Url:         itemFeed.Link,
			Description: description,
			PublishedAt: publishedAt,
		}); err != nil {
			// if error is a duplicate, ignore it and continue
			if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				fmt.Errorf("failed to save feed: %w", err)
			}
			// else continue if not fatal
			continue
		}
	}

}
