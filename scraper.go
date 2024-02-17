package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rohit-iwnl/rss-aggregator/internal/database"
)

func startScraping(
	db *database.Queries,
	concurrency int,
	timeBetweenRequest time.Duration,
){
	log.Printf("Starting scraping with %v go routines every %s duration", concurrency, timeBetweenRequest)

	ticker := time.NewTicker(timeBetweenRequest)

	for ; ; <-ticker.C{
		feeds,err := db.GetNextFeedToFetch(
			context.Background(),
			int32(concurrency),
		)
		if err != nil {
			log.Printf("Error getting feeds to fetch: %v", err)
			continue
		}

		wg := &sync.WaitGroup{}
		for _, feed := range feeds{
			wg.Add(1)
			
			go scrapeFeed(db,wg,feed)
		}

		wg.Wait()
	}
}

func scrapeFeed(db *database.Queries,wg *sync.WaitGroup, feed database.Feed){
	defer wg.Done()

	_,err := db.MarkFeedAsFetched(context.Background(),feed.ID)
	if err != nil {
		log.Printf("Error marking feed %v as fetched: %v",feed.ID, err)
	}

	rssFeed,err := urlToFeed(feed.Url)

	if err != nil {
		log.Printf("Error fetching feed %v: %v", feed.Url, err)
		return	
	}

	for _, item := range rssFeed.Channel.Item{

		description := sql.NullString{}

		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}
		pubAt,err:= time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("Error parsing time %v: %v", item.PubDate, err)
			continue
		}

		_,err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Title: item.Title,
			Description: description,
			PublishedAt: pubAt,
			Url: item.Link,
			FeedID: feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key"){
				continue
			}	
			log.Printf("Error creating post for feed %v: %v", feed.ID, err)
		}
	}

	log.Printf("Feed %s collected, %v posts found",feed.Name, len(rssFeed.Channel.Item))

}