package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"wb-project/internal/logger/sl"
	"wb-project/internal/metric"
	"wb-project/internal/models"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// 1. Объявляем интерфейс.
//
//go:generate mockery --name=OrderProvider --output=./mocks --case=underscore
type OrderProvider interface {
	GetOrder(ctx context.Context, uid string) (models.Order, error)
}

type OrderHandler struct {
	service OrderProvider // Используем интерфейс
}

func NewOrderHandler(s OrderProvider) *OrderHandler {
	return &OrderHandler{service: s}
}

//Запустить HTTP-сервер для выдачи данных по ID: реализовать HTTP-эндпоинт, который по order_id будет
//возвращать данные заказа из кеша (JSON API). Если в кеше данных нет, можно подтягивать из БД.

func (s *OrderHandler) GetOrderHandler(c *gin.Context) {
	ctx := c.Request.Context()
	uid := c.Param("order_uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неправильный ID"})
		return
	}

	slog.Info("выполняем запрос",
		slog.String("uid", uid),
		sl.Traced(ctx),
	)
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("http.request.order_uid", uid))

	order, err := s.service.GetOrder(ctx, uid)
	if err != nil {
		slog.Error("order не найден",
			slog.String("uid", uid),
			slog.Any("error", err),
			sl.Traced(ctx))
		span.RecordError(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Введен неверный ID: заказ не найден"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()
		// После того как хендлер отработал, фиксируем время и статус
		duration := time.Since(start)
		status := c.Writer.Status()

		metric.ObserveRequest(duration, status)
	}
}
