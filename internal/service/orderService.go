// Package service содержит бизнес-логику приложения.
// Здесь определены службы для обработки заказов, валидации данных
// и координации работы между кэшем и репозиторием.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
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

	//1. Парсинг
	if err := json.Unmarshal(data, &order); err != nil {
		return fmt.Errorf("ошибка при парсинге, игнорируем: %v", err)
	}

	span.SetAttributes(attribute.String("order_uid", order.OrderUID))
	//2. Валидация данных, до сохранения в бд
	if err := s.validateOrder(&order); err != nil {
		return fmt.Errorf("валидация не пройдена %v", err)
	}

	start := time.Now()
	//3. Сохранение в бд
	if err := s.repo.Save(ctx, order); err != nil {
		span.RecordError(err)
		metric.DbOperationsTotal.WithLabelValues("save", "error").Inc()
		return fmt.Errorf("ошибка сохранения в БД: %v", err)
	}
	span.AddEvent("order сохранен в бд")

	//Метрика, которая увеличивается, чтобы показать кол-во успешных запросов в бд(сохранения заказов)
	metric.DbOperationsTotal.WithLabelValues("save", "success").Inc()
	metric.DbDuration.WithLabelValues("save").Observe(time.Since(start).Seconds())

	//4. Добавление в кеш
	s.cache.Set(order.OrderUID, &order)
	span.AddEvent("order добавлен в кеш")

	fmt.Println("Успешно сохранен order: ", order.OrderUID)
	return nil
}

// GetOrder - функция для получения
func (s *OrderService) GetOrder(ctx context.Context, uid string) (models.Order, error) {
	//чтобы, понимать откуда пришел отчет
	tr := otel.Tracer("orderService")
	//запускает секундомер, и создает запись о конкретной операции
	//тут создается trace id если это самый первый вызов и span id
	//Функция смотрит в старый ctx. Если там уже лежит Trace ID (например, от Gin), то новый спан автоматически становится «ребенком» предыдущего
	ctx, span := tr.Start(ctx, "Service.GerOrder")
	defer span.End()
	//1. Поиск в кеше
	if fromCache, ok := s.cache.Get(uid); ok {
		metric.CacheHitsTotal.WithLabelValues("hit").Inc()
		return *fromCache, nil
	}
	metric.CacheHitsTotal.WithLabelValues("miss").Inc()

	span.SetAttributes(attribute.String("order_uid", uid))
	//2. возвращаем из БД, пробрасывая контекст
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
	//2. Запрос в бд, для получения всех заказов
	orders, err := s.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("не удалось прочитать данные из кэш при старте: %w", err)
	}

	//3. Добавление в кэш
	for i := range orders {
		s.cache.Set(orders[i].OrderUID, &orders[i])
	}
	metric.CacheSize.Set(float64(len(orders)))
	log.Printf("Кэш успешно восстановлен: загружено %d записей", len(orders))
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
