package handler

import (
	"net/http"
	"orderfc/cmd/order/usecase"
	"orderfc/infrastructure/log"
	"orderfc/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	OrderUsecase usecase.OrderUsecase
}

func NewOrderHandler(orderUsecase usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{OrderUsecase: orderUsecase}
}

func (h *OrderHandler) Ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	}
}

// CheckOutOrder godoc
// @Summary 주문 생성
// @Description 인증된 사용자의 주문을 생성하고 order_id를 반환합니다.
// @Tags ORDER
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CheckoutRequest true "주문 요청"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders [post]
func (h *OrderHandler) CheckOutOrder(c *gin.Context) {
	var checkoutRequest models.CheckoutRequest
	if err := c.ShouldBindJSON(&checkoutRequest); err != nil {
		log.Logger.Info().Err(err).Msg("Invalid JSON format in checkout request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIdStr, ok := c.Get("user_id")
	if !ok {
		log.Logger.Info().Msg("User ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}
	userId, ok := userIdStr.(float64)
	if !ok {
		log.Logger.Info().Msg("Invalid user ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	if len(checkoutRequest.Items) == 0 {
		log.Logger.Info().Msg("Items are required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Items are required"})
		return
	}

	checkoutRequest.UserID = int64(userId)

	orderId, err := h.OrderUsecase.CheckOutOrder(c.Request.Context(), &checkoutRequest)
	if err != nil {
		log.Logger.Info().Err(err).Msg("Error checking out order")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully", "order_id": orderId})
}

// GetOrderHistoryByUserId godoc
// @Summary 주문 내역 조회
// @Description 인증된 사용자의 주문 내역을 조회합니다.
// @Tags ORDER
// @Security BearerAuth
// @Produce json
// @Param status query int false "주문 상태"
// @Success 200 {array} models.Order
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders/history [get]
func (h *OrderHandler) GetOrderHistoryByUserId(c *gin.Context) {
	var params models.OrderHistoryParam

	userIdStr, ok := c.Get("user_id")
	if !ok {
		log.Logger.Info().Msg("User ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}
	userId, ok := userIdStr.(float64)
	if !ok {
		log.Logger.Info().Msg("Invalid user ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	params.UserID = int64(userId)

	statusStr := c.Query("status")
	if statusStr != "" {
		status, err := strconv.Atoi(statusStr)
		if err != nil {
			log.Logger.Info().Err(err).Msg("Invalid status")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}
		params.Status = status
	}

	results, err := h.OrderUsecase.GetOrderHistoryByUserId(c.Request.Context(), params)
	if err != nil {
		log.Logger.Info().Err(err).Msg("Error getting order history by user id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetSalesReport godoc
// @Summary 매출 리포트 조회
// @Description 일 단위 매출 리포트를 조회합니다.
// @Tags ORDER
// @Security BearerAuth
// @Produce json
// @Param days query int false "조회 기간(일)" default(30)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/orders/sales-report [get]
func (h *OrderHandler) GetSalesReport(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter"})
		return
	}

	results, err := h.OrderUsecase.GetDailySalesReport(c.Request.Context(), days)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Error getting sales report")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": results,
		"days":   days,
		"note":   "Uses CTE + Window Functions (cumulative_revenue, revenue_rank)",
	})
}
