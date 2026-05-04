package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"sanrio-auction-api/pkg/response"

	"github.com/gin-gonic/gin"
)

const maxUploadSize = 5 << 20 // 5 MB

type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir}
}

// UploadImage 處理 POST /api/v1/upload，需要 JWT 認證。
// 接受 multipart/form-data，欄位名稱為 image（上限 5MB）。
// 圖片儲存至 static/uploads/ 目錄，回傳可直接引用的路徑：
//
//	{ "url": "/uploads/<filename>" }
//
// 前端使用時需在前面加上 API Server 的 host，例如：
//
//	http://localhost:9090/uploads/<filename>
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// 限制請求 body 大小，防止大型檔案佔用記憶體
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		response.BadRequest(c, "請上傳圖片（欄位名稱：image）")
		return
	}
	defer file.Close()

	// 只允許常見圖片格式，防止上傳可執行檔
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
	default:
		response.BadRequest(c, "僅支援 jpg、jpeg、png、gif、webp 格式")
		return
	}

	// 用 Unix nanosecond timestamp 產生唯一檔名，避免命名衝突
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(h.uploadDir, filename)

	if err := c.SaveUploadedFile(header, dst); err != nil {
		response.InternalError(c, "儲存圖片失敗")
		return
	}

	response.Success(c, gin.H{"url": "/uploads/" + filename})
}
