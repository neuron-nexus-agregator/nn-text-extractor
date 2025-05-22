package app

import (
	"os"
	"sync"

	"agregator/text-extractor/internal/interfaces"
	model "agregator/text-extractor/internal/model/kafka"
	"agregator/text-extractor/internal/service/extractor"
	"agregator/text-extractor/internal/service/kafka"
)

type App struct {
	ext    *extractor.Extractor
	kafka  *kafka.Kafka
	logger interfaces.Logger
}

func New(logger interfaces.Logger) (*App, error) {
	ext, err := extractor.New("../../config/cfg.json", logger)
	if err != nil {
		logger.Error("Ошибка при создании сервиса извлечения текста", "error", err)
		return nil, err
	}
	kafka := kafka.New([]string{os.Getenv("KAFKA_ADDR")}, "extract-full-text", "preprocessor", "extractor", logger)
	return &App{
		ext:    ext,
		kafka:  kafka,
		logger: logger,
	}, nil
}

func (a *App) Run() {
	wg := sync.WaitGroup{}
	output := make(chan model.Item)
	input := make(chan model.Item)
	wg.Add(3)
	go func() {
		defer wg.Done()
		a.kafka.StartReading(output)
	}()
	go func() {
		defer wg.Done()
		a.kafka.StartWriting(input)
	}()
	go func() {
		defer wg.Done()
		for item := range output {
			data, err := a.ext.Extract(item.Link)
			if err == nil {
				item.FullText = data
			} else {
				a.logger.Error("Ошибка при извлечении текста", "error", err)
			}
			input <- item
		}
	}()
	wg.Wait()
}
