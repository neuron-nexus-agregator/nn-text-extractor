package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	kafka_model "agregator/text-extractor/internal/model/kafka"
)

type Kafka struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

func New(brokers []string, readTopic, writeTopic, groupID string) *Kafka {
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
	}
}

// Функция для чтения сообщений из Kafka
func (k *Kafka) StartReading(output chan<- kafka_model.Item) {
	defer k.reader.Close()
	for {
		msg, err := k.reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from Kafka: %v\n", err)
			continue
		}
		item := kafka_model.Item{}
		err = json.Unmarshal(msg.Value, &item)
		if err != nil {
			log.Printf("Error decoding Kafka message: %v\n", err)
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
			log.Printf("Error encoding item to JSON: %v\n", err)
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
			log.Printf("Error writing message to Kafka: %v\n", err)
		}
	}
}
