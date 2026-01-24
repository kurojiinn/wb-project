package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"wb-project/internal/config"
)

func main() {
	// 1. Главный контекст, который передаем
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Загрузка конфигурации
	cfg := config.LoadConfig()

	application, err := NewApplication(cfg)
	if err != nil {
		log.Fatalf("Ошибка при инициализации приложения: %v", err)
	}
	if err = application.Run(ctx); err != nil {
		log.Fatalf("Ошибка запуске приложения: %v", err)

	}
	log.Println("Сервис успешно остановлен")
}
