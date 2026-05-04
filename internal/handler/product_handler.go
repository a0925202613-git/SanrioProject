package handler

import (
	"net/http"
	"strconv"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/internal/service"
	"sanrio-auction-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProductHandler struct {
	productService *service.ProductService
	validate       *validator.Validate
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService, validate: validator.New()}
}

// Create 處理 POST /api/v1/products，需要 JWT 認證。
// 建立成功後同時讓 Redis 的商品列表快取失效（由 Service 層負責）。
func (h *ProductHandler) Create(c *gin.Context) {
	var req model.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	p, err := h.productService.Create(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, p)
}

// GetByID 處理 GET /api/v1/products/:id，不需認證，支援 Redis 快取。
// 路徑參數 id 必須是合法的整數，否則提前回傳 400。
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	p, err := h.productService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "product not found")
		return
	}
	response.Success(c, p)
}

// List 處理 GET /api/v1/products，不需認證，支援 Redis 快取（第一頁且無篩選條件）。
// Query String 參數（character、page、page_size）由 ShouldBindQuery 自動綁定。
func (h *ProductHandler) List(c *gin.Context) {
	var query model.ListProductsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	products, err := h.productService.List(c.Request.Context(), &query)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, products)
}

// Update 處理 PUT /api/v1/products/:id，需要 JWT 認證。
// 支援 Partial Update：只傳入要修改的欄位即可，其餘保持原值。
// 更新成功後同時讓 Redis 中對應的商品快取失效（由 Service 層負責）。
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	var req model.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	p, err := h.productService.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.NotFound(c, "product not found")
		return
	}
	response.Success(c, p)
}

// Delete 處理 DELETE /api/v1/products/:id，需要 JWT 認證。
// 成功回傳 204 No Content（無 body），符合 REST 語意。
// 同時讓 Redis 中對應的商品快取失效（由 Service 層負責）。
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	if err := h.productService.Delete(c.Request.Context(), id); err != nil {
		response.NotFound(c, "product not found")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
