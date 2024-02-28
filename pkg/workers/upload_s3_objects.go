package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/util"
)

const (
	UPLOAD_S3_OBJECT           = "task:upload_s3_object"
	UPLOAD_MULTIPLE_S3_OBJECTS = "task:upload_multiple_s3_objects"
	DELETE_S3_OBJECT           = "task:delete_s3_object"
)

func (distributor *RedisTaskClientDistributor) SendS3ObjectUploadTask(ctx context.Context, payload *util.PayloadUploadImage, opts ...asynq.Option) error {
	payloadBuffer, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(UPLOAD_S3_OBJECT, payloadBuffer, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Println("send s3 object task")
	fmt.Printf("Enqueued task: %v of max retries: %v\n", info.Type, info.MaxRetry)
	return nil
}

func (distributor *RedisTaskClientDistributor) SendMultipleS3ObjectUploadTask(ctx context.Context, payload []*util.PayloadUploadImage, opts ...asynq.Option) error {
	payloadBuffer, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(UPLOAD_MULTIPLE_S3_OBJECTS, payloadBuffer, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v\n", info.Type, info.MaxRetry)
	return nil
}

func (processor *RedisTaskServerProcessor) ProcessTaskUploadS3Object(ctx context.Context, task *asynq.Task) error {
	var Payload util.PayloadUploadImage
	err := json.Unmarshal(task.Payload(), &Payload)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("unmarshalling error %w", err)
	}

	fmt.Printf("BEGIN @%+v\n", time.Now())
	fmt.Printf("start processing task %+s\n", task.Type())

	err = processor.client.UploadImage(ctx, Payload.ObjectKey, processor.envs.S3_BUCKET_NAME, Payload.Extension, Payload.Image)
	if err != nil {
		fmt.Print(time.Now())
		return err
	}
	fmt.Printf("END @%+v\n", time.Now())

	return nil
}

func (processor *RedisTaskServerProcessor) ProcessTaskMultipleUploadS3Object(ctx context.Context, task *asynq.Task) error {
	var Payload []*util.PayloadUploadImage
	err := json.Unmarshal(task.Payload(), &Payload)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("unmarshalling error %w", err)
	}

	fmt.Printf("BEGIN @%+v\n", time.Now())
	fmt.Printf("start processing task %+s\n", task.Type())

	err = processor.client.UploadMultipleImages(ctx, Payload, processor.envs.S3_BUCKET_NAME)
	if err != nil {
		fmt.Print(time.Now())
		return err
	}
	fmt.Printf("END @%+v\n", time.Now())

	return nil
}

func (distributor *RedisTaskClientDistributor) SendS3ObjectDeleteTask(ctx context.Context, images []string, opts ...asynq.Option) error {
	payload, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("error occured while marshaling %w", err)
	}
	
	task := asynq.NewTask(DELETE_S3_OBJECT, payload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("enqueueing task error %w", err)
	}
	
	fmt.Printf("Enqueued task: %v of max retries: %v\n", info.Type, info.MaxRetry)
	return nil
}

func (processor *RedisTaskServerProcessor) ProcessTaskDeleteS3Object(ctx context.Context, task *asynq.Task) error {
	var payload []string
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("error occured while unmarshalling %w", err)
	}

	fmt.Printf("BEGIN @%+v\n", time.Now())
	fmt.Printf("start processing task %+s\n", task.Type())
	for _, image := range payload {
		err := processor.client.DeleteImage(ctx, image, processor.envs.S3_BUCKET_NAME)
		if err != nil {
			return err
		}
	}
	fmt.Printf("END @%+v\n", time.Now())
	return nil
}
