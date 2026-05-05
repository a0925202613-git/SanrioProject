package handler

import (
	"strconv"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/internal/service"
	"sanrio-auction-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FlashSaleHandler struct {
	flashSaleService *service.FlashSaleService
	validate         *validator.Validate
}

func NewFlashSaleHandler(flashSaleService *service.FlashSaleService) *FlashSaleHandler {
	return &FlashSaleHandler{flashSaleService: flashSaleService, validate: validator.New()}
}

// Create 處理 POST /api/v1/flash-sales，需要 JWT 認證。
// 時間格式由 Service 層驗證（RFC3339），Handler 只負責基本欄位驗證。
func (h *FlashSaleHandler) Create(c *gin.Context) {
	var req model.CreateFlashSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	fs, err := h.flashSaleService.Create(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, fs)
}

// GetByID 處理 GET /api/v1/flash-sales/:id，不需認證。
func (h *FlashSaleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid flash sale id")
		return
	}

	fs, err := h.flashSaleService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "flash sale not found")
		return
	}
	response.Success(c, fs)
}

// List 處理 GET /api/v1/flash-sales，不需認證。
// 只回傳 end_time > NOW() 的活動。
func (h *FlashSaleHandler) List(c *gin.Context) {
	sales, err := h.flashSaleService.List(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, sales)
}

// PurchaseSafe 處理 POST /api/v1/flash-sales/:id/purchase/safe，需要 JWT 認證。
// 內部使用 SELECT FOR UPDATE + Transaction，確保 remaining_stock 不會低於 0。
func (h *FlashSaleHandler) PurchaseSafe(c *gin.Context) {
	userID := c.GetInt64("userID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid flash sale id")
		return
	}

	order, err := h.flashSaleService.PurchaseWithLock(c.Request.Context(), id, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, order)
}

// PurchaseUnsafe 處理 POST /api/v1/flash-sales/:id/purchase/unsafe，需要 JWT 認證。
// 內部不加鎖，可觀察超賣（Overselling）問題。
// 建議透過 ConcurrentTestUnsafe 觀察，效果更明顯。
func (h *FlashSaleHandler) PurchaseUnsafe(c *gin.Context) {
	userID := c.GetInt64("userID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid flash sale id")
		return
	}

	order, err := h.flashSaleService.PurchaseUnsafe(c.Request.Context(), id, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, order)
}

// ConcurrentTestSafe 處理 POST /api/v1/flash-sales/:id/concurrent-test/safe，需要 JWT 認證。
// 同時啟動 N 個 goroutine 搶購，使用 SELECT FOR UPDATE 保護。
// 預期結果：remaining_stock >= 0，success_count 精確等於原始庫存。
func (h *FlashSaleHandler) ConcurrentTestSafe(c *gin.Context) {
	userID := c.GetInt64("userID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid flash sale id")
		return
	}

	var req model.ConcurrentTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.flashSaleService.RunConcurrentTest(c.Request.Context(), id, userID, req.Goroutines, true)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// ConcurrentTestUnsafe 處理 POST /api/v1/flash-sales/:id/concurrent-test/unsafe，需要 JWT 認證。
// 同時啟動 N 個 goroutine 搶購，不加任何鎖。
// 預期結果：remaining_stock < 0（超賣），success_count 可能超過原始庫存。
func (h *FlashSaleHandler) ConcurrentTestUnsafe(c *gin.Context) {
	userID := c.GetInt64("userID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid flash sale id")
		return
	}

	var req model.ConcurrentTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.flashSaleService.RunConcurrentTest(c.Request.Context(), id, userID, req.Goroutines, false)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}
