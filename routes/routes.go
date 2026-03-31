package routes

import (
	"net/http"
	"orderfc/cmd/order/handler"
	"orderfc/cmd/order/resource"
	"orderfc/config"
	"orderfc/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(router *gin.Engine, orderHandler *handler.OrderHandler) {
	router.Use(middleware.RequestLogger())
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/ping", orderHandler.Ping())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "orderfc",
		})
	})

	router.GET("/debug/queries", func(c *gin.Context) {
		if resource.DBMonitor == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "monitor not initialized"})
			return
		}
		c.JSON(http.StatusOK, resource.DBMonitor.GetDebugInfo())
	})

	// private API (인증 필요)
	private := router.Group("/api")
	private.Use(middleware.AuthMiddleware(config.GetJwtSecret()))
	{
		private.POST("/v1/orders", orderHandler.CheckOutOrder)
		private.GET("/v1/orders/history", orderHandler.GetOrderHistoryByUserId)
		private.GET("/v1/orders/sales-report", orderHandler.GetSalesReport)
	}
}
