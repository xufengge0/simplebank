package kafkago

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/segmentio/kafka-go"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/worker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserRegistration struct {
	Username     string `json:"username"`
	HashPassword string `json:"hash_password"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
}

var ConsumerReader *kafka.Reader

type KafkaConsumer struct {
	Store           db.Store
	TaskDistributor worker.TaskDistributor
	Reader          *kafka.Reader
}

func (consumer *KafkaConsumer) ConsumeUserRegistrationMessages(ctx context.Context) {
	ConsumerReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"host.docker.internal:29092"},
		GroupID:        "test",
		Topic:          "user_registration",
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	ConsumerWriter := kafka.Writer{
		Addr:                   kafka.TCP("host.docker.internal:29092"),
		Topic:                  "register-responses",
		Balancer:               &kafka.Hash{},
		WriteTimeout:           10 * time.Second,
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
		BatchSize:              200,                  // 批量大小
		BatchTimeout:           1 * time.Millisecond, // 批量超时时间
	}

	for {
		// 从 Kafka 中读取消息
		msg, err := ConsumerReader.ReadMessage(ctx)
		if err != nil {
			log.Printf("failed to read message from Kafka: %v", err)
			continue
		}

		// 反序列化 Kafka 消息
		var registration UserRegistration
		err = json.Unmarshal(msg.Value, &registration)
		if err != nil {
			log.Printf("failed to unmarshal user registration: %v", err)
			continue
		}

		// 执行创建用户的数据库事务
		var responseMsg kafka.Message
		user, err := consumer.createUser(ctx, registration)

		if err != nil {
			log.Printf("failed to create user: %v", err)
			responseMsg = kafka.Message{
				Key:   []byte(registration.Username),
				Value: []byte(err.Error()),
			}
		} else {

			value, _ := json.Marshal(user)
			responseMsg = kafka.Message{
				Key:   []byte(registration.Username),
				Value: []byte(value),
			}
		}

		err = ConsumerWriter.WriteMessages(ctx, responseMsg)
		if err != nil {
			log.Printf("failed to write response message: %v", err)
		}
	}
}

// 创建用户的方法
func (consumer *KafkaConsumer) createUser(ctx context.Context, registration UserRegistration) (db.User, error) {
	arg := db.CreateUserTXParams{
		CreateUserParams: db.CreateUserParams{
			Username:       registration.Username,
			HashedPassword: registration.HashPassword,
			FullName:       registration.FullName,
			Email:          registration.Email,
		},

		AfterCreate: func(user db.User) error {
			opts := []asynq.Option{
				asynq.MaxRetry(10),                // 最大重试次数
				asynq.ProcessIn(10 * time.Second), // 10秒后再处理任务
				asynq.Queue(worker.QueueCritical), // 队列名称
			}

			// 发送验证邮件任务
			payload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}
			err := consumer.TaskDistributor.DistributeTaskSendVerifyEmail(ctx, payload, opts...)
			if err != nil {
				return fmt.Errorf("failed to send verification email task: %w", err)
			}
			return nil
		},
	}

	// 执行创建用户事务
	txResult, err := consumer.Store.CreateUserTX(ctx, arg)
	if err != nil {
		return db.User{}, status.Errorf(codes.Internal, "failed to create user:%s", err)
	}

	return txResult.User, nil
}
