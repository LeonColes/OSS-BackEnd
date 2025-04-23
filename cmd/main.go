package main

import (
	"log"

	"oss-backend/internal/app"
	"oss-backend/internal/config"
)

func main() {
	// 加载配置
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化应用
	app, err := app.Setup()
	if err != nil {
		log.Fatalf("Failed to setup application: %v", err)
	}

	// 运行应用
	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
