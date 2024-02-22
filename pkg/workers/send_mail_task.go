package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

type PayloadSendMail struct {
	Email string `json:"email"`
}

const SEND_MAIL_TASK_KEY string = "task:send_mail_key"

func (distributor *RedisTaskClientDistributor) SendMailTask(ctx context.Context, payload *PayloadSendMail, opts ...asynq.Option) error {
	payloadBuffer, err := json.Marshal(payload)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(SEND_MAIL_TASK_KEY, payloadBuffer, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Println(info)
	return nil
}
