package source

import (
	"github.com/Vilsol/yeet/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rs/zerolog/log"
	"mime"
	"path/filepath"
	"strings"
)

func GetS3(client *s3.S3, bucket string, path string) *utils.StreamHijacker {
	cleanedKey := strings.TrimPrefix(path, "/")

	object, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(cleanedKey),
	})

	if err != nil {
		log.Err(err).Msg("failed to get object")
		return nil
	}

	fileType := mime.TypeByExtension(filepath.Ext(filepath.Base(path)))

	return utils.NewStreamHijacker(int(*object.ContentLength), fileType, object.Body)
}
