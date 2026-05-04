package handler

import (
	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/internal/service"
	"sanrio-auction-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService *service.AuthService
	validate    *validator.Validate
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService, validate: validator.New()}
}

// Register 處理 POST /api/v1/auth/register。
// 成功回傳 201 與使用者資訊（不含密碼）。
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, user)
}

// Login 處理 POST /api/v1/auth/login。
// 成功回傳 200 與 { token, user }，token 供後續需要認證的 API 使用。
// 帳號不存在或密碼錯誤都回傳 401，避免前端區分兩種情況（防 User Enumeration）。
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

// Me 處理 GET /api/v1/me，需要 JWT 認證。
// 回傳目前登入使用者的資料，包含 username、email、balance 等欄位。
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetInt64("userID") // 由 AuthMiddleware 從 JWT 解析並注入

	user, err := h.authService.GetMe(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.Success(c, user)
}
