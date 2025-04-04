package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const configFile = "config.toml"

var logger *zap.Logger

func main() {
	fmt.Println("gore-bot v0.1")

	logger, err := newLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	logger.Debug("loading config file", zap.String("configFile", configFile))
	config, err := loadConfig(configFile)
	if err != nil {
		logger.Error("failed to load config file", zap.String("configFile", configFile), zap.Error(err))
		log.Fatal(err)
	}
	logger.Debug("config loaded", zap.Any("config", config))

	logger.Debug("attempting to load .env file")
	err = godotenv.Load()
	if err != nil {
		logger.Warn("no .env found, depending on actual environment variables")
	}

	creds := TwitterCredentials{
		AccessToken:       os.Getenv("TWITTER_ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("TWITTER_CONSUMER_SECRET"),
	}
	logger.Debug("creating new Twitter client")
	client, err := newClient(creds, logger)
	if err != nil {
		logger.Error("error creating creating client", zap.Error(err))
		log.Fatal(err)
	}

	demux := newTwitterStreamDemuxer(createTweetHandler(config, logger))
	stream, err := newFilteredStream(client, []string{config.Twitter.FollowUserID}, logger)
	if err != nil {
		logger.Error("error creating filtered stream", zap.Error(err))
		log.Fatal(err)
	}

	// logging to stdout for debug purposes to see a timestamp at a glance
	log.Println("attaching to stream")
	logger.Info("attaching to stream")
	go demux.HandleChan(stream.Messages)

	// listen for SIGINT/SIGTERM (e.g. Ctrl-C)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("received signal", zap.Any("signal", <-ch))

	logger.Info("stopping stream")
	stream.Stop()
}

func newLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"debug.log"}
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	return config.Build()
}

// Config holds the basic config parameters defined in a config.toml file
type Config struct {
	Twitter struct {
		FollowUserID string `toml:"follow_userid"`
	} `toml:"twitter"`
	S3 struct {
		BucketName   string `toml:"bucket_name"`
		PS4Folder    string `toml:"ps4_folder"`
		SwitchFolder string `toml:"switch_folder"`
	} `toml:"s3"`
}

func loadConfig(fn string) (Config, error) {
	config := Config{}
	_, err := toml.DecodeFile(fn, &config)
	return config, err
}
