package model

import (
	"time"
)

// OrderStatus 是資料庫的 ENUM 型別，追蹤訂單的生命週期。
// 對應 migrations/init.sql 中的 CREATE TYPE order_status AS ENUM (...)
type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"   // 已建立，待付款
	OrderPaid      OrderStatus = "paid"      // 已付款
	OrderShipped   OrderStatus = "shipped"   // 已出貨
	OrderCompleted OrderStatus = "completed" // 已完成
	OrderCanceled  OrderStatus = "canceled"  // 已取消
)

// Order 對應 orders 資料表。
// UnitPrice 記錄下單當下的單價（快照），避免商品改價後影響歷史訂單。
// TotalPrice = UnitPrice * Quantity，在 Repository 層計算後寫入。
// Product 為選填的關聯資料，可由 Service 層額外查詢後填入。
type Order struct {
	ID         int64       `json:"id"`
	UserID     int64       `json:"user_id"`
	ProductID  int64       `json:"product_id"`
	Quantity   int         `json:"quantity"`
	UnitPrice  float64     `json:"unit_price"`  // 下單時的單價快照
	TotalPrice float64     `json:"total_price"` // 總金額 = UnitPrice * Quantity
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`

	Product *Product `json:"product,omitempty"` // 選填，由 Service 層填入
}

// CreateOrderRequest 是 POST /orders 的請求體。
// UserID 從 JWT context 取得，不需要在請求體中提供。
type CreateOrderRequest struct {
	ProductID int64 `json:"product_id" validate:"required,gt=0"`
	Quantity  int   `json:"quantity" validate:"required,min=1"`
}

// UpdateStatusRequest 是 PATCH /orders/:id/status 的請求體。
type UpdateStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required,oneof=pending paid shipped completed canceled"`
}

// BatchOrderRequest 是批次更新訂單狀態的請求體，適用於 PATCH /orders/batch-status。
type BatchOrderRequest struct {
	IDs    []int64     `json:"ids" binding:"required,min=1"`
	Status OrderStatus `json:"status"` // 僅批次更新狀態時需要
}
