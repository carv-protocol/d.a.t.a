package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/resources"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/user/userlookup"
	"github.com/michimani/gotwi/user/userlookup/types"

	manageTypes "github.com/michimani/gotwi/tweet/managetweet/types"
	"github.com/michimani/gotwi/tweet/searchtweet"
	searchTypes "github.com/michimani/gotwi/tweet/searchtweet/types"
)

type TwitterMode string

const (
	TwitterModeAPI     TwitterMode = "api"
	TwitterModeScraper TwitterMode = "scraper"
)

// Interface defines the contract
type ITwitter interface {
	GetMe() string
	Tweet(ctx context.Context, text string) error
	MonitorMentioned(ctx context.Context) ([]*Tweet, error)
	ReplyToTweet(ctx context.Context, replyText, replyToTweetID string) (*Tweet, error)
	DeleteTweet(ctx context.Context, tweetID string) error
	GetTweetByID(ctx context.Context, tweetID string) (*Tweet, error)
	MonitorHashtag(ctx context.Context, hashtag string, duration time.Duration) ([]*Tweet, error)
}

// Tweet represents a simplified Twitter post structure
type Tweet struct {
	ID        string
	Text      string
	UserID    string
	CreatedAt time.Time
	Metrics   *TweetMetrics
}

// TweetMetrics contains engagement metrics for a tweet
type TweetMetrics struct {
	LikeCount    int
	RetweetCount int
	ReplyCount   int
	QuoteCount   int
}

type TwitterConfig struct {
	Mode          TwitterMode `mapstructure:"mode"`     // Mode of operation: "api" or "scraper"
	Username      string      `mapstructure:"username"` // Twitter username
	Password      string      `mapstructure:"password"` // Twitter password
	APIKey        string      `mapstructure:"api_key"`
	APIKeySecret  string      `mapstructure:"api_key_secret"`
	AccessToken   string      `mapstructure:"access_token"`
	TokenSecret   string      `mapstructure:"token_secret"`
	MonitorWindow int         `mapstructure:"monitor_window"` // Duration in minutes, e.g. 20
}

// TwitterOauth represents a Twitter API client with authentication and user context
type TwitterOauth struct {
	client *gotwi.Client
	user   *resources.User
	tweets []resources.Tweet
	config *TwitterConfig // Add config field for future reference
}

// NewTwitterClient returns the interface type
func NewTwitterClient(twitterConfig *TwitterConfig) (ITwitter, error) {
	if twitterConfig == nil {
		return nil, fmt.Errorf("twitter config is nil")
	}
	if err := validateConfig(twitterConfig); err != nil {
		return nil, fmt.Errorf("invalid twitter config: %w", err)
	}

	// Returns concrete implementations (TwitterOauth/TwitterScraper) as ITwitter
	switch twitterConfig.Mode {
	case TwitterModeAPI:
		return newTwitterAPIClient(twitterConfig) // Returns *TwitterOauth
	case TwitterModeScraper:
		return newTwitterScraper(twitterConfig) // Returns *TwitterScraper
	default:
		return nil, fmt.Errorf("invalid twitter mode: %s", twitterConfig.Mode)
	}
}

// validateConfig validates the Twitter configuration
func validateConfig(config *TwitterConfig) error {
	if config.Mode == "" {
		return fmt.Errorf("TWITTER_MODE must be specified (api or scraper)")
	}

	switch config.Mode {
	case TwitterModeAPI:
		if config.APIKey == "" || config.APIKeySecret == "" {
			return fmt.Errorf("TWITTER_API_KEY and TWITTER_API_KEY_SECRET are required for API mode")
		}
		if config.AccessToken == "" || config.TokenSecret == "" {
			return fmt.Errorf("TWITTER_ACCESS_TOKEN and TWITTER_TOKEN_SECRET are required for API mode")
		}
	case TwitterModeScraper:
		if config.Username == "" || config.Password == "" {
			return fmt.Errorf("TWITTER_USERNAME and TWITTER_PASSWORD are required for scraper mode")
		}
	default:
		return fmt.Errorf("invalid TWITTER_MODE: %s", config.Mode)
	}
	return nil
}

// MonitorMentioned monitors mentions of the authenticated user
func (t *TwitterOauth) MonitorMentioned(ctx context.Context) ([]*Tweet, error) {
	// Use configured monitor window, default to 20 minutes if not set
	// Note:Do not quickly check the tweets, because maybe the twitter api is rate limited
	monitorWindow := t.config.MonitorWindow
	if monitorWindow <= 0 {
		monitorWindow = 20
	}

	startTime := time.Now().Add(-time.Duration(monitorWindow) * time.Minute)
	l := &searchTypes.ListRecentInput{
		StartTime: &startTime,
		SortOrder: searchTypes.ListSortOrderRecency,
		Query:     fmt.Sprintf("@%s", *t.user.Username),
		TweetFields: fields.TweetFieldList{
			fields.TweetFieldAuthorID,
			fields.TweetFieldCreatedAt,
			fields.TweetFieldText,
		},
	}

	output, err := searchtweet.ListRecent(ctx, t.client, l)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent tweets: %w", err)
	}

	// Check if output or output.Data is nil
	if output == nil || output.Data == nil {
		return make([]*Tweet, 0), nil
	}

	result := make([]*Tweet, 0, len(output.Data))
	for _, tweet := range output.Data {
		// Add null checks for required fields
		if tweet.ID == nil || tweet.Text == nil || tweet.AuthorID == nil {
			continue // Skip tweets with missing required fields
		}

		tweetObj := &Tweet{
			ID:     *tweet.ID,
			Text:   *tweet.Text,
			UserID: *tweet.AuthorID,
		}

		// Add creation time if available
		if tweet.CreatedAt != nil {
			tweetObj.CreatedAt = *tweet.CreatedAt
		}

		result = append(result, tweetObj)
	}

	return result, nil
}

func (t *TwitterOauth) GetMe() string {
	return *t.user.ID
}

func (t *TwitterOauth) Tweet(ctx context.Context, tweet string) error {
	p := &manageTypes.CreateInput{
		Text: gotwi.String(tweet),
	}

	_, err := managetweet.Create(ctx, t.client, p)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// ReplyToTweet replies to a specific tweet
func (t *TwitterOauth) ReplyToTweet(ctx context.Context, replyText, replyToTweetID string) (*Tweet, error) {
	p := &manageTypes.CreateInput{
		Text: gotwi.String(replyText),
		Reply: &manageTypes.CreateInputReply{
			InReplyToTweetID: replyToTweetID,
		},
	}

	resp, err := managetweet.Create(ctx, t.client, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}

	return &Tweet{
		ID:     *resp.Data.ID,
		Text:   *resp.Data.Text,
		UserID: t.GetMe(),
	}, nil
}

// GetTweetByID retrieves a specific tweet by its ID
func (t *TwitterOauth) GetTweetByID(ctx context.Context, tweetID string) (*Tweet, error) {
	// Implementation for getting a specific tweet
	// TODO: Implement using gotwi library
	return nil, nil
}

// DeleteTweet deletes a tweet by its ID
func (t *TwitterOauth) DeleteTweet(ctx context.Context, tweetID string) error {
	p := &manageTypes.DeleteInput{
		ID: tweetID,
	}

	_, err := managetweet.Delete(ctx, t.client, p)
	if err != nil {
		return fmt.Errorf("failed to delete tweet: %w", err)
	}

	return nil
}

// LikeTweet likes a specific tweet
func (t *TwitterOauth) LikeTweet(ctx context.Context, tweetID string) error {
	// TODO: Implement using gotwi library
	return nil
}

// MonitorHashtag monitors tweets containing specific hashtags
func (t *TwitterOauth) MonitorHashtag(ctx context.Context, hashtag string, duration time.Duration) ([]*Tweet, error) {
	startTime := time.Now().Add(-duration)
	l := &searchTypes.ListRecentInput{
		StartTime: &startTime,
		SortOrder: searchTypes.ListSortOrderRecency,
		Query:     fmt.Sprintf("#%s", hashtag),
	}

	output, err := searchtweet.ListRecent(ctx, t.client, l)
	if err != nil {
		return nil, fmt.Errorf("failed to search hashtag: %w", err)
	}

	return convertTweets(output.Data), nil
}

// convertTweets converts Twitter API tweets to our Tweet struct
func convertTweets(apiTweets []resources.Tweet) []*Tweet {
	result := make([]*Tweet, 0, len(apiTweets))
	for _, tweet := range apiTweets {
		t := &Tweet{
			ID:     *tweet.ID,
			Text:   *tweet.Text,
			UserID: *tweet.AuthorID,
		}
		if tweet.CreatedAt != nil {
			t.CreatedAt = *tweet.CreatedAt
		}
		result = append(result, t)
	}
	return result
}

// Rename the existing NewTwitterClient to newTwitterAPIClient
func newTwitterAPIClient(twitterConfig *TwitterConfig) (*TwitterOauth, error) {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           twitterConfig.AccessToken,
		OAuthTokenSecret:     twitterConfig.TokenSecret,
		APIKey:               twitterConfig.APIKey,
		APIKeySecret:         twitterConfig.APIKeySecret,
	}

	c, err := gotwi.NewClient(in)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	p := &types.GetMeInput{
		Expansions: fields.ExpansionList{
			fields.ExpansionPinnedTweetID,
		},
		UserFields: fields.UserFieldList{
			fields.UserFieldCreatedAt,
		},
		TweetFields: fields.TweetFieldList{
			fields.TweetFieldCreatedAt,
		},
	}

	u, err := userlookup.GetMe(context.Background(), c, p)
	if err != nil {
		return nil, err
	}

	return &TwitterOauth{
		client: c,
		user:   &u.Data,
		tweets: u.Includes.Tweets,
		config: twitterConfig,
	}, nil
}
