package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	kafka_model "agregator/text-extractor/internal/model/kafka"

	"agregator/text-extractor/internal/interfaces"
)

type Kafka struct {
	writer *kafka.Writer
	reader *kafka.Reader
	logger interfaces.Logger
}

func New(brokers []string, readTopic, writeTopic, groupID string, logger interfaces.Logger) *Kafka {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    writeTopic,
		Balancer: &kafka.LeastBytes{},
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   readTopic,
	})
	return &Kafka{
		writer: writer,
		reader: reader,
		logger: logger,
	}
}

// Функция для чтения сообщений из Kafka
func (k *Kafka) StartReading(output chan<- kafka_model.Item) {
	defer k.reader.Close()
	for {
		msg, err := k.reader.ReadMessage(context.Background())
		if err != nil {
			k.logger.Error("Error reading message from Kafka", "error", err)
			continue
		}
		item := kafka_model.Item{}
		err = json.Unmarshal(msg.Value, &item)
		if err != nil {
			k.logger.Error("Error decoding Kafka message", "error", err)
			continue
		}
		output <- item
	}
}

// Функция для записи сообщений в Kafka
func (k *Kafka) StartWriting(input <-chan kafka_model.Item) {
	for item := range input {
		data, err := json.Marshal(item)
		if err != nil {
			k.logger.Error("Error encoding item to JSON", "error", err)
			continue
		}
		message := kafka.Message{
			Key:   []byte(item.MD5), // Используем MD5 как ключ
			Value: data,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = k.writer.WriteMessages(ctx, message)
		cancel()
		if err != nil {
			k.logger.Error("Error writing message to Kafka", "error", err)
		}
	}
}
