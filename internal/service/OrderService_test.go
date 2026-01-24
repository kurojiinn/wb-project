package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"wb-project/internal/models"
	"wb-project/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setup(t *testing.T) (*mocks.OrderRepository, *mocks.OrderCache, *OrderService) {
	mockRepo := mocks.NewOrderRepository(t)
	mockCache := mocks.NewOrderCache(t)
	svc := NewOrderService(mockRepo, mockCache)

	return mockRepo, mockCache, svc
}

// проверяем, что все прошло успешно:
// 1. json.Unmarshal упешно
// 2. метод save в бд был вызван ровно один раз
// 3. метод set был вызван
// 4. метод вернул nil
func TestOrderService_HandleOrderMessage_Success(t *testing.T) {
	//1. Arrange(подготовка)
	mockRepo, mockCache, svc := setup(t)

	jsonData, _ := os.ReadFile("testdata/test_order.json")
	var expectedOrder models.Order
	_ = json.Unmarshal(jsonData, &expectedOrder)

	mockRepo.On("Save", mock.Anything, expectedOrder).Return(nil)
	mockCache.On("Set", expectedOrder.OrderUID, &expectedOrder).Return()

	//2. Act(Действие)
	err := svc.HandleOrderMessage(context.Background(), jsonData)

	//3. Assert
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// Метод вернул ошибку, содержащую фразу "ошибка при парсинге".
func TestOrderService_HandleOrderMessage_ParsingError(t *testing.T) {
	//1. Arrange(подготовка)
	mockRepo, mockCache, svc := setup(t)
	// Передаем строку, которая не распарсится
	badData := []byte("this is not a json")
	//2. Act(Действие)
	err := svc.HandleOrderMessage(context.Background(), badData)
	//3. Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при парсинге")

	mockRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	mockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything)
}

// Метод вернул ошибку "валидация не пройдена".
func TestOrderService_HandleOrderMessage_ValidationError(t *testing.T) {
	//1. Arrange(подготовка)
	mockRepo, mockCache, svc := setup(t)

	jsonData, _ := os.ReadFile("testdata/test_order_validation.json")
	var expectedOrder models.Order
	_ = json.Unmarshal(jsonData, &expectedOrder)
	//2. Act(Действие)
	err := svc.HandleOrderMessage(context.Background(), jsonData)
	//3. Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "валидация не пройдена")

	mockRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	mockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything)
}

// Метод вернул ошибку "ошибка сохранения в БД".
func TestOrderService_HandleOrderMessage_DBError(t *testing.T) {
	//1. Arrange(подготовка)
	mockRepo, mockCache, svc := setup(t)

	jsonData, _ := os.ReadFile("testdata/test_order.json")
	var expectedOrder models.Order
	_ = json.Unmarshal(jsonData, &expectedOrder)

	dbError := fmt.Errorf("connection refused")
	mockRepo.On("Save", mock.Anything, expectedOrder).Return(dbError)

	//2. Act(Действие)
	err := svc.HandleOrderMessage(context.Background(), jsonData)
	//3. Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка сохранения в БД")

	mockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything)
}

// Test для метода GetOrder
func TestOrderService_GetOrder_DBError(t *testing.T) {
	//1. Arrange(подготовка)
	mockRepo, mockCache, svc := setup(t)
	uid := "some_uid"
	mockCache.On("Get", uid).Return(nil, false)
	dbErr := errors.New("db error")
	mockRepo.On("Get", mock.Anything, uid).Return(models.Order{}, dbErr)

	//2. Act(Действие)
	_, err := svc.GetOrder(context.Background(), uid)
	//3. Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order не найден в БД")
	mockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything)
	mockRepo.AssertNumberOfCalls(t, "Get", 1)

}

func TestOrderService_GetOrder_CacheMiss(t *testing.T) {
	//1. Arrange(подготовка)
	mockRepo, mockCache, svc := setup(t)

	order := models.Order{
		OrderUID: "1",
	}

	mockCache.On("Get", order.OrderUID).Return((*models.Order)(nil), false)
	mockRepo.On("Get", mock.Anything, order.OrderUID).Return(order, nil)
	mockCache.On("Set", order.OrderUID, &order).Return()

	//Act
	res, err := svc.GetOrder(context.Background(), order.OrderUID)

	//Assert
	assert.NoError(t, err)
	assert.Equal(t, order.OrderUID, res.OrderUID)
}

func TestOrderService_GetOrder_CacheHit(t *testing.T) {
	//1. Arrange(подготовка)
	mockRepo, mockCache, svc := setup(t)

	order := models.Order{
		OrderUID: "1",
	}

	mockCache.On("Get", order.OrderUID).Return(&order, true)

	//Act
	res, err := svc.GetOrder(context.Background(), order.OrderUID)

	//Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, order.OrderUID, res.OrderUID)
	mockCache.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "Get", 0)
}

// Test ReCache
func TestOrderService_ReCache_Success(t *testing.T) {
	//1. Arrange(подготовка)
	mockRepo, mockCache, svc := setup(t)
	orders := []models.Order{
		{OrderUID: "1"},
		{OrderUID: "2"},
	}

	mockRepo.On("GetAll", mock.Anything).Return(orders, nil)
	for _, ord := range orders {
		mockCache.On("Set", ord.OrderUID, mock.Anything).Return()

	}
	// 2. Act
	err := svc.ReCache(context.Background())

	// 3. Assert
	assert.NoError(t, err)

	mockRepo.AssertNumberOfCalls(t, "GetAll", 1)
	mockCache.AssertNumberOfCalls(t, "Set", len(orders))
}

func TestOrderService_ReCache_DBError(t *testing.T) {
	mockRepo, _, svc := setup(t)

	mockRepo.On("GetAll", mock.Anything).Return(nil, errors.New("db error"))

	err := svc.ReCache(context.Background())

	assert.Error(t, err)
}
