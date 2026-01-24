package handler

import (
	"context"
	"net/http"
	"time"
	"wb-project/internal/metric"
	"wb-project/internal/models"

	"github.com/gin-gonic/gin"
)

// 1. Объявляем интерфейс.
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
	id := c.Param("order_uid")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неправильный ID"})
		return
	}
	order, err := s.service.GetOrder(c.Request.Context(), id)
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
