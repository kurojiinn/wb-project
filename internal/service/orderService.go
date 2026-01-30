// Package service содержит бизнес-логику приложения.
// Здесь определены службы для обработки заказов, валидации данных
// и координации работы между кэшем и репозиторием.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	"wb-project/internal/logger/sl"
	"wb-project/internal/metric"
	"wb-project/internal/models"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// OrderRepository описывает контракт для постоянного хранения и получения заказов.
// Он абстрагирует логику работы с базой данных от бизнес-логики сервиса.
//
//go:generate mockery --name=OrderRepository --output=./mocks --case=underscore
type OrderRepository interface {
	Save(ctx context.Context, order models.Order) error
	Get(ctx context.Context, uid string) (models.Order, error)
	GetAll(ctx context.Context) ([]models.Order, error)
}

// OrderCache определяет контракт для высокопроизводительного
// хранения заказов в оперативной памяти.
//
//go:generate mockery --name=OrderCache --output=./mocks --case=underscore
type OrderCache interface {
	Set(uid string, order *models.Order)
	Get(uid string) (*models.Order, bool)
}

// OrderService предоставляет методы для управления заказами,
// включая их обработку, сохранение в БД и кэширование.
type OrderService struct {
	repo     OrderRepository // Используем интерфейс, а не struct
	cache    OrderCache      // Используем интерфейс
	validate *validator.Validate
}

// NewOrderService принимает интерфейсы.
func NewOrderService(repo OrderRepository, orderCache OrderCache) *OrderService {
	return &OrderService{
		repo:     repo,
		cache:    orderCache,
		validate: validator.New(),
	}
}

// HandleOrderMessage - функция для получения заказов
func (s *OrderService) HandleOrderMessage(ctx context.Context, data []byte) error {
	tr := otel.Tracer("orderService")
	ctx, span := tr.Start(ctx, "HandleOrderMessage")

	defer span.End()
	var order models.Order

	slog.Debug("парсинг order", sl.Traced(ctx))
	//1. Парсинг
	if err := json.Unmarshal(data, &order); err != nil {
		slog.Error("failed to unmarshal order", slog.Any("error", err), sl.Traced(ctx))
		return fmt.Errorf("ошибка при парсинге, игнорируем: %w", err)
	}
	slog.Info("order успешно распарсен", slog.String("order_uid", order.OrderUID), sl.Traced(ctx))

	span.SetAttributes(attribute.String("order_uid", order.OrderUID))
	//2. Валидация данных, до сохранения в бд
	if err := s.validateOrder(&order); err != nil {
		return fmt.Errorf("валидация не пройдена %w", err)
	}

	start := time.Now()
	//3. Сохранение в бд
	if err := s.repo.Save(ctx, order); err != nil {
		slog.Error("failed to save order to db",
			slog.String("order_uid", order.OrderUID),
			slog.Any("error", err),
			sl.Traced(ctx),
		)
		span.RecordError(err)
		metric.DbOperationsTotal.WithLabelValues("save", "error").Inc()
		return fmt.Errorf("ошибка сохранения в БД: %w", err)
	}
	span.AddEvent("order сохранен в бд")

	//Метрика, которая увеличивается, чтобы показать кол-во успешных запросов в бд(сохранения заказов)
	metric.DbOperationsTotal.WithLabelValues("save", "success").Inc()
	metric.DbDuration.WithLabelValues("save").Observe(time.Since(start).Seconds())

	//4. Добавление в кеш
	s.cache.Set(order.OrderUID, &order)
	span.AddEvent("order добавлен в кеш")
	slog.Info("Успешно сохранен order", slog.String("order_uid", order.OrderUID), sl.Traced(ctx))
	return nil
}

// GetOrder - функция для получения
func (s *OrderService) GetOrder(ctx context.Context, uid string) (models.Order, error) {
	//1.1 чтобы, понимать откуда пришел отчет
	tr := otel.Tracer("orderService")
	ctx, span := tr.Start(ctx, "GetOrder")
	defer span.End()

	span.SetAttributes(attribute.String("order_uid", uid))
	//2. Поиск в кеше
	if fromCache, ok := s.cache.Get(uid); ok {
		span.AddEvent("cache hit")
		slog.Info("order найден в кеше", slog.String("uid", uid), sl.Traced(ctx))
		metric.CacheHitsTotal.WithLabelValues("hit").Inc()
		return *fromCache, nil
	}

	span.AddEvent("cache miss")
	slog.Info("Order не найдет в кеше, идем в бд", slog.String("uid", uid), sl.Traced(ctx))
	metric.CacheHitsTotal.WithLabelValues("miss").Inc()

	//3. возвращаем из БД, пробрасывая контекст
	found, err := s.repo.Get(ctx, uid)
	if err != nil {
		span.RecordError(err)
		metric.DbOperationsTotal.WithLabelValues("get", "error").Inc()
		return models.Order{}, fmt.Errorf("order не найден в БД %w", err)
	}

	//3. Нашли в бд, обновляем кеш
	s.cache.Set(uid, &found)
	metric.DbOperationsTotal.WithLabelValues("get", "success").Inc()

	return found, nil
}

// ReCache - функция для насыщения кэша
func (s *OrderService) ReCache(ctx context.Context) error {
	//1.1 trace
	tr := otel.Tracer("orderService")

	ctx, span := tr.Start(ctx, "Service.ReCache")
	defer span.End()

	slog.Info("Старт разогрева кеша", sl.Traced(ctx))
	start := time.Now()
	//2. Запрос в бд, для получения всех заказов
	orders, err := s.repo.GetAll(ctx)
	if err != nil {
		slog.Error("Ошибка при загрузке всех пользователей",
			slog.Any("error", err),
			sl.Traced(ctx))
		span.RecordError(err)
		return fmt.Errorf("не удалось прочитать данные из кэш при старте: %w", err)
	}

	//3. Добавление в кэш
	for i := range orders {
		s.cache.Set(orders[i].OrderUID, &orders[i])
	}
	//4. Обновление метрик
	metric.CacheSize.Set(float64(len(orders)))

	duration := time.Since(start)
	slog.Info("Кеш успешно разогрет",
		slog.Int("count", len(orders)),
		slog.Duration("duration", duration),
		sl.Traced(ctx),
	)

	span.SetAttributes(
		attribute.Int("orders.count", len(orders)),
		attribute.String("duration", duration.String()),
	)
	return nil
}

// validateOrder - функция для валидации заказов
func (s *OrderService) validateOrder(order *models.Order) error {
	if err := s.validate.Struct(order); err != nil {
		return err
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("заказ не содержит товаров")
	}
	return nil
}
