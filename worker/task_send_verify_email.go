package worker

import (
	"context"

	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
)

const TaskSendVerifyEmail = "task:send_verify:email" // 任务名

// 存储在redis中的任务
type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

// 任务创建、分配
func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal failed: %v", err)
	}

	// 构建任务
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	// 发送任务到redis
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("could not enqueue task: %v", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")

	return nil
}

// 处理任务
func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("json unmarshal failed: %v", asynq.SkipRetry)
	}

	// 从数据库中获取用户
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		// if err == sql.ErrNoRows {
		// 	return fmt.Errorf("user does not exist: %v", asynq.SkipRetry) // 跳过重试
		// }
		return fmt.Errorf("could not get user: %v", err)
	}

	// 在数据库中创建verify_email
	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("could not create verify email: %v", err)
	}

	// 发送验证邮件
	subject := "Simple Bank - Verify Email"
	verifyUrl := fmt.Sprintf("http://localhost:8080/v1/verify_email?email_id=%d&secret_code=%s",
		verifyEmail.ID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>`,user.FullName, verifyUrl)
	to := []string{user.Email}

	err = processor.mailer.SenderEmail(subject,content,to,nil,nil,nil)
	if err!= nil {
		return fmt.Errorf("could not send email: %v", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processing task: send verify email")
	return nil
}
