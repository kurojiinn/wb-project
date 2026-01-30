package handler

import (
	"context"
	"net/http"
	"time"
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
	uid := c.Param("order_uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неправильный ID"})
		return
	}
	ctx := c.Request.Context()

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("http.request.order_uid", uid))

	order, err := s.service.GetOrder(ctx, uid)
	if err != nil {
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
