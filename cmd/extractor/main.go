package main

import "agregator/text-extractor/internal/pkg/app"

func main() {
	app, err := app.New()
	if err != nil {
		panic(err)
	}
	app.Run()
}
