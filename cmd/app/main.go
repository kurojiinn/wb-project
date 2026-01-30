package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wb-project/internal/config"
	"wb-project/internal/trace"
)

func main() {
	// 1. Главный контекст, который передаем
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Загрузка конфигурации
	cfg := config.LoadConfig()

	tp, err := trace.InitTracer(ctx)
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Printf("Tracer shutdown error: %v", err)
		}
	}()
	
	application, err := NewApplication(cfg)
	if err != nil {
		log.Fatalf("Ошибка при инициализации приложения: %v", err)
	}
	if err = application.Run(ctx, tp); err != nil {
		log.Fatalf("Ошибка запуске приложения: %v", err)

	}
	log.Println("Сервис успешно остановлен")
}
