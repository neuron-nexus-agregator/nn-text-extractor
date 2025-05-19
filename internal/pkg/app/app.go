package app

import (
	"log"
	"os"
	"sync"

	model "agregator/text-extractor/internal/model/kafka"
	"agregator/text-extractor/internal/service/extractor"
	"agregator/text-extractor/internal/service/kafka"
)

type App struct {
	ext   *extractor.Extractor
	kafka *kafka.Kafka
}

func New() (*App, error) {
	ext, err := extractor.New("../../config/cfg.json")
	if err != nil {
		return nil, err
	}
	kafka := kafka.New([]string{os.Getenv("KAFKA_ADDR")}, "extract-full-text", "preprocessor", "extractor")
	return &App{
		ext:   ext,
		kafka: kafka,
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
				log.Default().Println(err)
			}
			input <- item
		}
	}()
	wg.Wait()
}
