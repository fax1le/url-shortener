package main

import (
	"url-shortener/internal/app"
	"url-shortener/internal/config"
	"url-shortener/internal/logger/logrus"
)

func main() {
	cfg := config.Load()

	logger := logrus.New()

	App := app.New(cfg, logger)

	App.Init()

	App.Run()
}
