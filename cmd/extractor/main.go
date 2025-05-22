package main

import (
	"agregator/text-extractor/internal/pkg/app"
	"log/slog"
)

func main() {
	app, err := app.New(slog.Default())
	if err != nil {
		panic(err)
	}
	app.Run()
}
