package source

import (
	"github.com/Vilsol/yeet/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var _ Source = (*S3)(nil)

type S3 struct {
	S3Client *s3.S3
	Bucket   string
}

func NewS3(bucket string, key string, secret string, endpoint string, region string) (*S3, error) {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
	}

	newSession, err := session.NewSession(s3Config)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create S3 session")
	}

	s3Client := s3.New(newSession)

	return &S3{
		S3Client: s3Client,
		Bucket:   bucket,
	}, nil
}

func (s S3) Get(path string, _ []byte) (*utils.StreamHijacker, bool) {
	return GetS3(s.S3Client, s.Bucket, path)
}

func (s S3) IndexPath(dirPath string, f IndexFunc) (int64, int64, error) {
	totalSize := int64(0)
	totalCount := int64(0)

	if err := s.S3Client.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(dirPath),
	}, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, object := range page.Contents {
			cleanedPath := cleanPath(*object.Key, dirPath)
			totalSize += f(*object.Key, cleanedPath)
			totalCount++

			log.Trace().Msgf("Indexed: %s -> %s", cleanedPath, *object.Key)
		}
		return true
	}); err != nil {
		return 0, 0, err
	}

	return totalSize, totalCount, nil
}

func (s S3) Watch() (<-chan WatchEvent, error) {
	// TODO Index bucket every N minutes
	return nil, errors.New("s3 does not support watching")
}
