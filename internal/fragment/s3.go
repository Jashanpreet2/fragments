package fragment

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/Jashanpreet2/fragments/internal/config"
	"github.com/Jashanpreet2/fragments/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3Config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client struct {
	*s3.Client
}

var s3Client *S3Client

func GetS3Client() (*S3Client, error) {
	if s3Client != nil {
		return s3Client, nil
	}

	config.Config()
	cfg, err := s3Config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Sugar.Infow("Error loading default config", "Error", err.Error())
		return nil, err
	}

	s3Client = &S3Client{s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = "us-east-1"
	})}

	return s3Client, nil
}

func (s3Client *S3Client) GetFragmentDataFromS3(username string, key string) ([]byte, error) {
	object, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket:       aws.String(os.Getenv("S3_BUCKET")),
		Key:          aws.String(username + "/" + key),
		ChecksumMode: types.ChecksumModeEnabled,
	})

	if err != nil {
		return nil, err
	}

	return io.ReadAll(object.Body)
}

func (s3Client *S3Client) UploadFragmentDataToS3(username string, key string, data []byte) error {
	_, err := s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:            aws.String(os.Getenv("S3_BUCKET")),
		Key:               aws.String(username + "/" + key),
		ChecksumAlgorithm: types.ChecksumAlgorithmCrc32,
		Body:              bytes.NewReader(data),
	})

	return err
}
