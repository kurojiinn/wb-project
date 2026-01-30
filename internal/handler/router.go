package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouter(orderHandler *OrderHandler) *gin.Engine {
	router := gin.Default()
	// "wb-order-service" — это имя, по которому ты будешь искать трейсы в Jaeger
	router.Use(otelgin.Middleware("wb-order-service"))

	router.Static("/static", "static")
	router.StaticFile("/", "static/index.html")

	router.Use(MetricsMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/order")
	{
		api.GET("/:order_uid", orderHandler.GetOrderHandler)
		api.GET("/", func(context *gin.Context) {
			context.String(200, "Сервер работает")
		})
	}
	return router
}
