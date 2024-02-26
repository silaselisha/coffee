package util

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type CoffeeShopBucket struct {
	client *s3.Client
}

func NewS3Client(config aws.Config, opts ...func(*s3.Options)) *CoffeeShopBucket {
	client := s3.NewFromConfig(config, opts...)
	return &CoffeeShopBucket{
		client: client,
	}
}

func (csb *CoffeeShopBucket) UploadImage(ctx context.Context, fileName string, objectKey string, bucketName string, extension string, image []byte) error {
	body := bytes.NewBuffer(image)
	_, err := csb.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   body,
		ACL:    types.ObjectCannedACL(*aws.String("public-read")),
		ContentType: aws.String(fmt.Sprintf("image/%s", extension)),
	})

	if err != nil {
		return fmt.Errorf("error occured while uploading the image to AWS s3 bucket %w", err)
	}

	return nil
}
