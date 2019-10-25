package main

import (
	"os"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.uber.org/zap"
)

func newAWSSession(region string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
}

// taken from https://golangcode.com/uploading-a-file-to-s3/
func uploadFile(s *session.Session, bucket string, fn string, folder string, platform string, game string) error {
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer file.Close()

	// TODO: Create value used for `Key` below which incorporates platform/game info
	// e.g. key = folder/platform-game-fn

	uploader := s3manager.NewUploader(s)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fn),
		Body:   file,
	})

	return err
}

func processMediaURLs(urls []string) {}

func processMediaURL(url string, platform string, game string) {
	logger.Info("processing URL", zap.String("url", url))
}
