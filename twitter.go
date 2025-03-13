package main

import (
	"fmt"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"go.uber.org/zap"
)

// TwitterCredentials stores all the access/consumer tokens and secret keys
// required for authentication against the twitter REST API
type TwitterCredentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

// newClient creates a new instance of twitter.Client with passed credentials
func newClient(creds TwitterCredentials, logger *zap.Logger) (*twitter.Client, error) {
	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	token := oauth1.NewToken(creds.AccessToken, creds.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	user, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		logger.Error("getClient: error verifying credentials", zap.Error(err))
		return nil, err
	}
	logger.Info("user account", zap.Any("account-info", user))
	return client, nil
}

func newTwitterStreamDemuxer(handleTweet func(*twitter.Tweet)) twitter.SwitchDemux {
	demux := twitter.NewSwitchDemux()
	demux.Tweet = handleTweet
	return demux
}

func newFilteredStream(client *twitter.Client, ids []string, logger *zap.Logger) (*twitter.Stream, error) {
	params := &twitter.StreamFilterParams{
		Follow:        ids,
		StallWarnings: twitter.Bool(true),
	}
	logger.Info("creating filtered stream", zap.Any("params", params))
	return client.Streams.Filter(params)
}

// tweetHandler is the entry point for handling incoming tweets to check for
// attached media and process any media URLs found
func createTweetHandler(config Config, logger *zap.Logger) func(*twitter.Tweet) {
	return func(tweet *twitter.Tweet) {
		fmt.Printf("[%s] tweeted “%s”\n", tweet.User.ScreenName, tweet.Text)
		logger.Info("handling tweet", zap.Any("tweet", tweet))
		platform, game := getMeta(getHashTags(tweet))
		timestamp, err := getTimestamp(tweet.CreatedAt)
		if err != nil {
			timestamp = "unknown"
			logger.Error("handling tweet: error generating timestamp from createdAt", zap.Error(err), zap.String("CreatedAt", tweet.CreatedAt))
		}
		logger.Info("handling tweet: hashtag metadata", zap.String("platform", platform), zap.String("game", game))
		urls := getMediaURLs(tweet)
		if len(urls) == 0 {
			logger.Info("handling tweet: no media entities found in tweet")
			return
		}
		logger.Info("handling tweet: extracted media URLs from tweet", zap.Strings("mediaURLs", urls))
		processMediaURLs(config, logger, urls, timestamp, platform, game)
	}
}

func getHashTags(t *twitter.Tweet) []string {
	tags := []string{}
	for _, tag := range t.Entities.Hashtags {
		tags = append(tags, tag.Text)
	}
	return tags
}

func getMeta(tags []string) (string, string) {
	platform := "unknown"
	game := "unknown"
	for _, tag := range tags {
		if tag == "ps4" {
			platform = "ps4"
		} else if tag == "switch" {
			platform = "switch"
		} else {
			game = tag
		}
	}
	return platform, game
}

func getTimestamp(at string) (string, error) {
	// parse from: "Wed Oct 30 20:43:07 +0000 2019"
	date, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", at)
	if err != nil {
		return "", err
	}
	return date.Format("20060102150405"), nil
}

func getMediaURLs(t *twitter.Tweet) []string {
	urls := []string{}
	for _, media := range t.Entities.Media {
		urls = append(urls, media.MediaURLHttps)
	}
	return urls
}
