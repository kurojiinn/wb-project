package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"wb-project/internal/handler/mocks"
	"wb-project/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderHandler_GetOrderHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Заказ найден", func(t *testing.T) {
		mockService := mocks.NewOrderProvider(t)
		orderUID := "test_uid"
		expectedOrder := models.Order{OrderUID: orderUID}

		mockService.On("GetOrder", mock.Anything, orderUID).Return(expectedOrder, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Params = []gin.Param{{Key: "order_uid", Value: orderUID}}

		h := NewOrderHandler(mockService)
		h.GetOrderHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var actualOrder models.Order
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &actualOrder))
		assert.Equal(t, expectedOrder.OrderUID, actualOrder.OrderUID)
		assert.Equal(t, http.StatusOK, w.Code)

	})

	t.Run("Заказ не найден в системе", func(t *testing.T) {
		mockService := mocks.NewOrderProvider(t)

		badUID := "unknown"
		mockService.On("GetOrder", mock.Anything, badUID).Return(models.Order{}, errors.New("not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "order_uid", Value: badUID}}
		c.Request, _ = http.NewRequest("GET", "/", nil)

		h := NewOrderHandler(mockService)
		h.GetOrderHandler(c)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("Пустой ID", func(t *testing.T) {
		mockService := mocks.NewOrderProvider(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Params = []gin.Param{{Key: "order_uid", Value: ""}}
		c.Request, _ = http.NewRequest("GET", "/", nil)

		h := NewOrderHandler(mockService)
		h.GetOrderHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertNotCalled(t, "GetOrder")
	})
}
