package main

import (
	"fmt"

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
func newClient(creds TwitterCredentials) (*twitter.Client, error) {
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

func newFilteredStream(client *twitter.Client, ids []string) (*twitter.Stream, error) {
	params := &twitter.StreamFilterParams{
		Follow:        ids,
		StallWarnings: twitter.Bool(true),
	}
	logger.Info("creating filtered stream", zap.Any("params", params))
	return client.Streams.Filter(params)
}

// tweetHandler is the entry point for handling incoming tweets to check for
// attached media and process any media URLs found
func tweetHandler(tweet *twitter.Tweet) {
	fmt.Printf("[%s] tweeted “%s”\n", tweet.User.ScreenName, tweet.Text)
	logger.Info("handling tweet", zap.Any("tweet", tweet))
	urls := getMediaURLs(tweet)
	if len(urls) == 0 {
		logger.Info("no media entities found in tweet")
		return
	}
	logger.Info("extracted media URLs from tweet", zap.Strings("mediaURLs", urls))
}

func getMediaURLs(t *twitter.Tweet) []string {
	urls := []string{}
	for _, media := range t.Entities.Media {
		urls = append(urls, media.MediaURLHttps)
	}
	return urls
}
