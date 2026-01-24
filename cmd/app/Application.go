package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"wb-project/internal/app"
	"wb-project/internal/cache"
	"wb-project/internal/config"
	"wb-project/internal/db/conn"
	"wb-project/internal/db/repository"
	"wb-project/internal/handler"
	"wb-project/internal/kafka"
	"wb-project/internal/service"
)

type Application struct {
	srv      *app.Server
	consumer *kafka.OrderConsumer
	producer *kafka.OrderProducer
	service  *service.OrderService
}

func NewApplication(cfg *config.Config) (*Application, error) {
	// 3. Подключение к БД
	dbConn, err := conn.Connection(&cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("подключение к БД: %w", err)
	}
	// 5. Сборка слоев
	orderCache := cache.NewOrderCache(1*time.Minute, 30*time.Second)
	orderRepo := repository.NewOrderRepository(dbConn)
	orderService := service.NewOrderService(orderRepo, orderCache)
	orderHandler := handler.NewOrderHandler(orderService)
	srv := app.NewServer(orderHandler)

	if err = kafka.EnsureTopicExists(cfg.KafkaConfig.Brokers, cfg.KafkaConfig.Topic); err != nil {
		return nil, fmt.Errorf("создание Kafka topic: %w", err)
	}

	producer, err := kafka.NewProducer(cfg.KafkaConfig.Brokers, cfg.KafkaConfig.Topic)
	if err != nil {
		return nil, fmt.Errorf("создание Kafka Producer: %w", err)
	}

	consumer, err := kafka.NewOrderConsumer(cfg.KafkaConfig.Brokers, cfg.KafkaConfig.Topic, orderService.HandleOrderMessage)
	if err != nil {
		return nil, fmt.Errorf("создание Kafka Consumer: %w", err)
	}

	return &Application{
		srv:      srv,
		consumer: consumer,
		producer: producer,
		service:  orderService,
	}, nil
}

func (app *Application) Run(ctx context.Context) error {

	if err := app.service.ReCache(ctx); err != nil {
		log.Printf("Не удалось восстановить кэш из БД: %v", err)
	}
	// Запуск консьюмера
	go func() {
		log.Println("Запуск Consumer...")
		if err := app.consumer.Start(ctx); err != nil {
			log.Printf("Consumer остановился с ошибкой: %v", err)
		}

	}()
	go func() {
		log.Println("Запуск HTTP сервера на :8080")
		if err := app.srv.Run(":8080"); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Printf("HTTP сервер в штаном режиме остановлен")
			} else {
				log.Fatalf("Критическая ошибка сервера: %v", err)
			}
		}
	}()
	if err := app.producer.Send(); err != nil {
		log.Printf("Не удалось отправить сообщение в Kafka: %v", err)
	}

	// 10. Ожидание сигнала завершения
	<-ctx.Done()
	log.Println("Получен сигнал завершения (Graceful Shutdown)...")

	// 11. Остановка HTTP сервера
	// Даем 5 секунд на завершение текущих запросов
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	app.Shutdown(ctx)

	if err := app.srv.Stop(shutdownCtx); err != nil {
		log.Printf("Ошибка при остановке HTTP сервера: %v", err)
	}

	return nil
}

func (app *Application) Shutdown(ctx context.Context) {
	if err := app.srv.Stop(ctx); err != nil {
		log.Printf("Ошибка остановки HTTP сервера: %v", err)
	}
	if err := app.consumer.Close(); err != nil {
		log.Printf("Ошибка остановки Kafka Consumer: %v", err)
	}
	if err := app.producer.Close(); err != nil {
		log.Printf("Ошибка остановки Kafka Producer: %v", err)
	}
}
