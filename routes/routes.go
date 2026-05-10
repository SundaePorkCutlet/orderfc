package routes

import (
	"context"
	"net/http"
	"orderfc/cmd/order/handler"
	"orderfc/cmd/order/resource"
	"orderfc/config"
	"orderfc/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, orderHandler *handler.OrderHandler, db *gorm.DB, redis *redis.Client) {
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
	router.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		status := gin.H{
			"status":  "ready",
			"service": "orderfc",
			"checks":  gin.H{},
		}
		checks := status["checks"].(gin.H)

		sqlDB, err := db.DB()
		if err != nil {
			checks["database"] = "unavailable"
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			checks["database"] = "unavailable"
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		checks["database"] = "ok"

		if err := redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "unavailable"
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		checks["redis"] = "ok"

		c.JSON(http.StatusOK, status)
	})

	router.GET("/debug/queries", func(c *gin.Context) {
		if resource.DBMonitor == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "monitor not initialized"})
			return
		}
		c.JSON(http.StatusOK, resource.DBMonitor.GetDebugInfo())
	})

	router.GET("/debug/kafka", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":           "orderfc",
			"messages_produced": 0,
			"messages_consumed": 0,
			"dlq_count":         0,
			"consumer_stats":    gin.H{},
		})
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
