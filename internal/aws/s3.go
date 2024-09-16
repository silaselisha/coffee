package aws

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	types "github.com/silaselisha/coffee-api/types"
)

type CoffeeShopBucket interface {
	UploadImage(ctx context.Context, objectKey, bucketName, extension string, image []byte) error
	UploadMultipleImages(ctx context.Context, payload []*types.PayloadUploadImage, bucket string) error
	DeleteImage(ctx context.Context, objectKey string, bucket string) error
}

type CoffeeShopS3Client struct {
	client *s3.Client
}

func NewS3Client(config aws.Config, opts ...func(*s3.Options)) CoffeeShopBucket {
	client := s3.NewFromConfig(config, opts...)
	return &CoffeeShopS3Client{
		client: client,
	}
}

func (csb *CoffeeShopS3Client) UploadImage(ctx context.Context, 
	objectKey string, 
	bucketName string, 
	extension string, 
	image []byte) error {
	body := bytes.NewBuffer(image)
	_, err := csb.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        body,
		ACL:         s3Types.ObjectCannedACL(*aws.String("public-read")),
		ContentType: aws.String(fmt.Sprintf("image/%s", extension)),
	})

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error occured while uploading the image to AWS s3 bucket %w", err)
	}

	return nil
}

func (csb *CoffeeShopS3Client) UploadMultipleImages(ctx context.Context, 
	payload []*types.PayloadUploadImage, 
	bucketName string) error {
	for _, image := range payload {
		body := bytes.NewBuffer(image.Image)
		_, err := csb.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(image.ObjectKey),
			Body:        body,
			ACL:         s3Types.ObjectCannedACL(*aws.String("public-read")),
			ContentType: aws.String(fmt.Sprintf("image/%s", image.Extension)),
		})

		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("error occured while uploading the image to AWS s3 bucket %w", err)
		}
	}
	return nil
}

func (csb *CoffeeShopS3Client) DeleteImage(ctx context.Context, objectKey string, bucketName string) error {
	_, err := csb.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return fmt.Errorf("error occured while deleting object %s from s3 aws bucket", objectKey)
	}

	return nil
}
