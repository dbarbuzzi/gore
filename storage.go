package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// TODO: define and implement an interface to allow easier swapping of backends

func newAWSSession(region string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
}

// taken from https://golangcode.com/uploading-a-file-to-s3/
func uploadFile(logger *zap.Logger, s *session.Session, bucket string, filepath string, folder string, filename string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	key := fmt.Sprintf("%s/%s", folder, filename)

	logger.Info(
		"uploading file to S3 bucket",
		zap.String("bucket name", bucket),
		zap.String("filepath", filepath),
		zap.String("Dest filename", key),
	)

	uploader := s3manager.NewUploader(s)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})

	return err
}

func processMediaURLs(config Config, logger *zap.Logger, urls []string, timestamp string, platform string, game string) error {
	for _, url := range urls {
		processMediaURL(config, logger, url, timestamp, platform, game)
	}
	return nil
}

func processMediaURL(config Config, logger *zap.Logger, url string, timestamp string, platform string, game string) (string, error) {
	// logger.Info("processing URL", zap.String("url", url))
	// save from Twitter to temp file

	logger.Info("saving image to disk", zap.String("url", url))
	rsp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	imageData, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	urlChunks := strings.Split(url, ".")
	extension := urlChunks[len(urlChunks)-1]

	fn := fmt.Sprintf("%s-%s-%s.%s", platform, game, timestamp, extension)
	logger.Debug("creating temp file", zap.String("filename", fn))
	tempfile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempfile.Name())
	if _, err := tempfile.Write(imageData); err != nil {
		return "", err
	}
	logger.Info("temp image file", zap.String("filename", tempfile.Name()))

	// upload temp file to S3

	logger.Debug("creating new AWS session")
	session, err := newAWSSession(os.Getenv("AWS_REGION "))
	if err != nil {
		return "", err
	}
	logger.Debug("uploading temp file to S3")
	err = uploadFile(logger, session, config.S3.BucketName, tempfile.Name(), platform, fn)

	return "", err
}
