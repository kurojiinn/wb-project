package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(orderHandler *OrderHandler) *gin.Engine {
	router := gin.Default()

	router.Static("/static", "static")
	router.StaticFile("/", "static/index.html")

	router.Use(MetricsMiddleware())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/order")
	{
		api.GET("/:order_uid", orderHandler.GetOrderHandler)
	}
	return router
}
