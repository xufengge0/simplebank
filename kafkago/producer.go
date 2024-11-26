package kafkago

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	db "github.com/techschool/simplebank/db/sqlc"
)

var (
	ProduceWriter *kafka.Writer
	ProduceReader *kafka.Reader
)

// 初始化全局的 kafka.Writer，建议在应用启动时调用
func InitKafkaWriter() {
	ProduceWriter = &kafka.Writer{
		Addr:                   kafka.TCP("host.docker.internal:29092"),
		Topic:                  "user_registration",
		Balancer:               &kafka.Hash{},
		WriteTimeout:           10 * time.Second,
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
		BatchSize:              200,                  // 批量大小
		BatchTimeout:           1 * time.Millisecond, // 批量超时时间
	}
	ProduceReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"host.docker.internal:29092"},
		Topic:          "register-responses",
		GroupID:        "response-consumers",
		CommitInterval: 0, // 设置为 0 以禁用自动提交
	})
}

// 在应用退出时调用该函数关闭 writer
func CloseKafkaWriter() {
	if ProduceWriter != nil {
		ProduceWriter.Close()
		ProduceReader.Close()
	}
}

// 使用全局 writer 写入 Kafka
func WriteKafka(ctx context.Context, registration UserRegistration) (db.User, error) {
	if ProduceWriter == nil {
		return db.User{}, fmt.Errorf("kafka writer not initialized")
	}

	// 序列化注册信息为 JSON
	msg, err := json.Marshal(registration)
	if err != nil {
		return db.User{}, err
	}

	for i := 0; i < 3; i++ {
		err := ProduceWriter.WriteMessages(ctx, kafka.Message{
			Key:   []byte(registration.Username),
			Value: msg,
		})

		if err != nil {
			if err == kafka.LeaderNotAvailable {
				// 指数退避
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return db.User{}, fmt.Errorf("failed to write message: %w", err)
		}
		break
	}

	for {
		// 读取 Kafka 消息
		msg, err := ProduceReader.ReadMessage(ctx)
		if err != nil {
			return db.User{}, fmt.Errorf("failed to read message: %w", err)
		}

		// 处理注册结果
		if string(msg.Key) == registration.Username {
			value := db.User{}
			if err := json.Unmarshal(msg.Value, &value); err != nil {
				return value, errors.New(string(msg.Value))
			}

			// 手动提交偏移量
			if err := ProduceReader.CommitMessages(ctx, msg); err != nil {
				return value, fmt.Errorf("failed to commit message: %w", err)
			}

			return value, nil

		} else {
			continue
		}
	}
}
