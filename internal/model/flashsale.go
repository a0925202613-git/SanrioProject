package model

import "time"

// FlashSale 對應 flash_sales 資料表，代表一場限時搶購活動。
// TotalStock 是活動設定的總數量，建立後不變。
// RemainingStock 是即時剩餘數量，每次成功搶購後遞減。
// 兩者的差值即為已售出數量。
type FlashSale struct {
	ID             int64     `json:"id"`
	ProductID      int64     `json:"product_id"`
	SalePrice      float64   `json:"sale_price"`      // 搶購優惠價
	TotalStock     int       `json:"total_stock"`     // 活動開始時的總庫存
	RemainingStock int       `json:"remaining_stock"` // 即時剩餘庫存（併發控制的關鍵欄位）
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	CreatedAt      time.Time `json:"created_at"`

	Product *Product `json:"product,omitempty"` // 選填，由 Service 層填入
}

// CreateFlashSaleRequest 是 POST /flash-sales 的請求體。
// TotalStock 同時作為 remaining_stock 的初始值寫入資料庫。
// StartTime / EndTime 使用 RFC3339 格式字串，由 Service 層解析。
type CreateFlashSaleRequest struct {
	ProductID  int64   `json:"product_id" validate:"required,gt=0"`
	SalePrice  float64 `json:"sale_price" validate:"required,gt=0"`
	TotalStock int     `json:"total_stock" validate:"required,min=1"`
	StartTime  string  `json:"start_time" validate:"required"` // RFC3339 格式
	EndTime    string  `json:"end_time" validate:"required"`   // RFC3339 格式
}

// ConcurrentTestRequest 是並發搶購測試的請求體。
// Goroutines 代表同時搶購的使用者數量（goroutine 數）。
type ConcurrentTestRequest struct {
	Goroutines int `json:"goroutines" validate:"required,min=1,max=500"`
}

// ConcurrentTestResult 是並發搶購測試的結果摘要。
// 透過比對 safe/unsafe 兩種模式的結果，可以清楚看到 Race Condition 的影響：
//   - unsafe: RemainingStock 可能為負數（超賣）
//   - safe:   RemainingStock 最低為 0，SuccessCount 精確等於原始庫存
type ConcurrentTestResult struct {
	Mode           string `json:"mode"`            // "safe" 或 "unsafe"
	Goroutines     int    `json:"goroutines"`      // 同時發出的 goroutine 數量
	SuccessCount   int    `json:"success_count"`   // 搶購成功筆數
	FailureCount   int    `json:"failure_count"`   // 搶購失敗筆數（庫存不足或其他錯誤）
	RemainingStock int    `json:"remaining_stock"` // 測試後的剩餘庫存（unsafe 可能為負）
	DurationMs     int64  `json:"duration_ms"`     // 所有 goroutine 完成的總耗時（毫秒）
}
