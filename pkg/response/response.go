package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "created", Data: data})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{Code: http.StatusBadRequest, Message: message})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{Code: http.StatusUnauthorized, Message: message})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{Code: http.StatusNotFound, Message: message})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{Code: http.StatusConflict, Message: message})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{Code: http.StatusInternalServerError, Message: message})
}
