package handler

import (
	"strconv"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/internal/service"
	"sanrio-auction-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrderHandler struct {
	orderService *service.OrderService
	validate     *validator.Validate
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService, validate: validator.New()}
}

// Create 處理 POST /api/v1/orders，需要 JWT 認證。
// userID 從 JWT context 取得，Request body 只需提供 product_id 與 quantity。
// Service 層會在同一個 Transaction 內完成：庫存檢查 → 餘額檢查 → 扣餘額 → 扣庫存 → 建訂單。
func (h *OrderHandler) Create(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, order)
}

// GetByID 處理 GET /api/v1/orders/:id，需要 JWT 認證。
// Service 層會驗證訂單是否屬於當前使用者，防止越權查看他人訂單（IDOR）。
func (h *OrderHandler) GetByID(c *gin.Context) {
	userID := c.GetInt64("userID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	order, err := h.orderService.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}
	response.Success(c, order)
}

// UpdateStatus 處理 PATCH /api/v1/orders/:id/status，需要 JWT 認證。
// 只能更新自己的訂單；Repository 層以 userID 條件防止 IDOR。
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	userID := c.GetInt64("userID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	var req model.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	order, err := h.orderService.UpdateStatus(c.Request.Context(), id, userID, req.Status)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}
	response.Success(c, order)
}

// ListMyOrders 處理 GET /api/v1/orders，需要 JWT 認證。
// 只回傳當前使用者自己的訂單（以 userID 過濾），依建立時間降冪排列。
func (h *OrderHandler) ListMyOrders(c *gin.Context) {
	userID := c.GetInt64("userID")

	orders, err := h.orderService.ListByUser(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, orders)
}

// BatchDelete 處理 DELETE /api/v1/orders/batch-delete，需要 JWT 認證。
// 只能刪除自己的訂單；Repository 層以 userID 條件防止 IDOR。
func (h *OrderHandler) BatchDelete(c *gin.Context) {
	var req model.BatchOrderRequest // 加上 model. 前綴
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "無效的請求資料"})
		return
	}

	userID := c.MustGet("userID").(int64)
	// 呼叫 Service 層
	if err := h.orderService.BatchDeleteOrders(c.Request.Context(), req.IDs, userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "批次刪除成功"})
}

func (h *OrderHandler) BatchUpdateStatus(c *gin.Context) {
	var req model.BatchOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "無效的請求資料"})
		return
	}

	userID := c.MustGet("userID").(int64)
	if err := h.orderService.BatchUpdateOrders(c.Request.Context(), req.IDs, userID, req.Status); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "批次更新成功"})
}
