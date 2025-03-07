package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/conf"

	twitterscraper "github.com/tyxben/twitter-scraper"
)

// TwitterScraper represents a Twitter scraper using browser automation
type TwitterScraper struct {
	scraper *twitterscraper.Scraper
	config  *conf.TwitterConfig
	userID  string // Store logged in user's ID
}

// NewTwitterScraper creates a new Twitter scraper with improved error handling and validation
func newTwitterScraper(config *conf.TwitterConfig) (*TwitterScraper, error) {
	// Validate config
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid twitter config: %w", err)
	}

	scraper := twitterscraper.New()

	// Login with retry mechanism
	var loginErr error
	for attempts := 0; attempts < 3; attempts++ {
		loginErr = scraper.Login(config.Username, config.Password)
		if loginErr == nil && scraper.IsLoggedIn() {
			break
		}
		time.Sleep(time.Second * 2)
	}

	if loginErr != nil || !scraper.IsLoggedIn() {
		return nil, fmt.Errorf("failed to login after multiple attempts: %v", loginErr)
	}

	// Get logged in user's profile
	profile, err := scraper.GetProfile(config.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &TwitterScraper{
		scraper: scraper,
		config:  config,
		userID:  profile.UserID,
	}, nil
}

// GetMe returns the logged-in user's ID
func (ts *TwitterScraper) GetMe() string {
	return ts.userID
}

// MonitorMentioned monitors mentions of the authenticated user
func (ts *TwitterScraper) MonitorMentioned(ctx context.Context) ([]*Tweet, error) {
	monitorWindow := ts.config.MonitorWindow
	if monitorWindow <= 0 {
		monitorWindow = 20
	}

	query := fmt.Sprintf("@%s", ts.config.Username)
	return ts.SearchTweets(ctx, query, 100) // Limit to recent 100 mentions
}

// Tweet posts a new tweet
func (ts *TwitterScraper) Tweet(ctx context.Context, text string) error {

	_, err := ts.scraper.CreateTweet(twitterscraper.NewTweet{
		Text:   text,
		Medias: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to post tweet: %w", err)
	}
	return nil
}

// ReplyToTweet replies to a specific tweet
func (ts *TwitterScraper) ReplyToTweet(ctx context.Context, replyText, replyToTweetID string) (*Tweet, error) {
	_, err := ts.scraper.CreateRetweet(replyToTweetID)
	if err != nil {
		return nil, fmt.Errorf("failed to reply to tweet: %w", err)
	}

	// Note: Since the scraper doesn't return the new tweet's ID, we return a simplified response
	return &Tweet{
		Text:      replyText,
		UserID:    ts.GetMe(),
		CreatedAt: time.Now(),
	}, nil
}

// DeleteTweet deletes a tweet by its ID
func (ts *TwitterScraper) DeleteTweet(ctx context.Context, tweetID string) error {
	err := ts.scraper.DeleteTweet(tweetID)
	if err != nil {
		return fmt.Errorf("failed to delete tweet: %w", err)
	}
	return nil
}

// MonitorHashtag monitors tweets containing specific hashtags
func (ts *TwitterScraper) MonitorHashtag(ctx context.Context, hashtag string, duration time.Duration) ([]*Tweet, error) {
	query := fmt.Sprintf("#%s", hashtag)
	return ts.SearchTweets(ctx, query, 100)
}

// SearchTweets searches for tweets matching a query with rate limiting protection
func (ts *TwitterScraper) SearchTweets(ctx context.Context, query string, limit int) ([]*Tweet, error) {
	var tweets []*Tweet

	// Add rate limiting protection
	rateLimiter := time.NewTicker(500 * time.Millisecond)
	defer rateLimiter.Stop()

	for tweet := range ts.scraper.SearchTweets(ctx, query, limit) {
		select {
		case <-ctx.Done():
			return tweets, ctx.Err()
		case <-rateLimiter.C:
			if tweet.Error != nil {
				return nil, fmt.Errorf("error searching tweets: %w", tweet.Error)
			}

			tweets = append(tweets, &Tweet{
				ID:        tweet.ID,
				Text:      tweet.Text,
				UserID:    tweet.UserID,
				CreatedAt: tweet.TimeParsed,
				Metrics: &TweetMetrics{
					LikeCount:    tweet.Likes,
					RetweetCount: tweet.Retweets,
					ReplyCount:   tweet.Replies,
				},
			})
		}
	}

	return tweets, nil
}

// GetTweetByID retrieves a specific tweet by its ID
func (ts *TwitterScraper) GetTweetByID(ctx context.Context, tweetID string) (*Tweet, error) {
	tweet, err := ts.scraper.GetTweet(tweetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tweet: %w", err)
	}

	return &Tweet{
		ID:        tweet.ID,
		Text:      tweet.Text,
		UserID:    tweet.UserID,
		CreatedAt: tweet.TimeParsed,
		Metrics: &TweetMetrics{
			LikeCount:    tweet.Likes,
			RetweetCount: tweet.Retweets,
			ReplyCount:   tweet.Replies,
		},
	}, nil
}
