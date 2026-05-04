package main

import (
	"sanrio-auction-api/internal/handler"
	"sanrio-auction-api/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth      *handler.AuthHandler
	Product   *handler.ProductHandler
	FlashSale *handler.FlashSaleHandler
	Order     *handler.OrderHandler
	Upload    *handler.UploadHandler
}

func RegisterRoutes(r *gin.Engine, jwtSecret string, h *Handlers) {
	auth := middleware.AuthMiddleware(jwtSecret)

	// 靜態檔案服務：讓前端可以直接透過 /uploads/<filename> 存取上傳的圖片
	r.Static("/uploads", "./static/uploads")

	v1 := r.Group("/api/v1")

	// Auth — 公開端點
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", h.Auth.Register)
		authGroup.POST("/login", h.Auth.Login)
	}

	// Me — 查看目前登入的使用者資料
	v1.GET("/me", auth, h.Auth.Me)

	// Upload — 上傳商品圖片，需要登入；回傳 { url: "/uploads/<filename>" }
	v1.POST("/upload", auth, h.Upload.UploadImage)

	// Products — 公開查詢，需登入才能增改刪；GET 支援 Redis 快取
	products := v1.Group("/products")
	{
		products.GET("", h.Product.List)
		products.GET("/:id", h.Product.GetByID)
		productAuth := products.Group("", auth)
		{
			productAuth.POST("", h.Product.Create)
			productAuth.PUT("/:id", h.Product.Update)
			productAuth.DELETE("/:id", h.Product.Delete)
		}
	}

	// Flash Sales — 限量搶購（固定價格、先搶先得）
	flashSales := v1.Group("/flash-sales")
	{
		flashSales.GET("", h.FlashSale.List)
		flashSales.GET("/:id", h.FlashSale.GetByID)
		flashSaleAuth := flashSales.Group("", auth)
		{
			flashSaleAuth.POST("", h.FlashSale.Create)

			// 單次搶購
			flashSaleAuth.POST("/:id/purchase/safe", h.FlashSale.PurchaseSafe)     // SELECT FOR UPDATE，防超賣
			flashSaleAuth.POST("/:id/purchase/unsafe", h.FlashSale.PurchaseUnsafe) // 無鎖，可觀察超賣

			// goroutine 並發搶購測試：一次呼叫同時啟動 N 個 goroutine
			// Request body: { "goroutines": 100 }
			flashSaleAuth.POST("/:id/concurrent-test/safe", h.FlashSale.ConcurrentTestSafe)     // 有鎖：庫存精確
			flashSaleAuth.POST("/:id/concurrent-test/unsafe", h.FlashSale.ConcurrentTestUnsafe) // 無鎖：觀察超賣
		}
	}

	// Orders — 全部需要登入；CREATE 使用多表 Transaction
	orders := v1.Group("/orders", auth)
	{
		orders.GET("", h.Order.ListMyOrders)
		orders.GET("/:id", h.Order.GetByID)
		orders.POST("", h.Order.Create)
		orders.PATCH("/:id/status", h.Order.UpdateStatus)
		orders.PATCH("/batch-status", h.Order.BatchUpdateStatus) // 批次更新訂單狀態
		orders.DELETE("/batch-delete", h.Order.BatchDelete)      // 批次刪除訂單
	}
}
