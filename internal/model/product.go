package model

import "time"

// SanrioCharacter 是資料庫的 ENUM 型別，限定商品只能屬於這五個角色之一。
// 對應 migrations/init.sql 中的 CREATE TYPE sanrio_character AS ENUM (...)
type SanrioCharacter string

const (
	HelloKitty  SanrioCharacter = "hello_kitty"
	Cinnamoroll SanrioCharacter = "cinnamoroll"
	Pompompurin SanrioCharacter = "pompompurin"
	MyMelody    SanrioCharacter = "my_melody"
	Kuromi      SanrioCharacter = "kuromi"
	Hangyodon   SanrioCharacter = "hangyodon"
	BadtzMaru   SanrioCharacter = "badtz_maru"
)

// Product 對應 products 資料表。
// BasePrice 是商品定價；在拍賣中作為起標價，在限時特賣中作為原價參考。
type Product struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Character   SanrioCharacter `json:"character"` // 三麗鷗角色分類
	Description string          `json:"description"`
	BasePrice   float64         `json:"base_price"`
	Stock       int             `json:"stock"` // 一般購買（訂單）使用的庫存
	ImageURL    string          `json:"image_url"`
	CreatedAt   time.Time       `json:"created_at"`
}

// CreateProductRequest 是 POST /products 的請求體。
// Character 使用 oneof 驗證，確保只接受合法的角色值。
type CreateProductRequest struct {
	Name        string          `json:"name" validate:"required,min=1,max=255"`
	Character   SanrioCharacter `json:"character" validate:"required,oneof=hello_kitty cinnamoroll pompompurin my_melody kuromi hangyodon badtz_maru"`
	Description string          `json:"description"`
	BasePrice   float64         `json:"base_price" validate:"required,gt=0"`
	Stock       int             `json:"stock" validate:"min=0"`
	ImageURL    string          `json:"image_url"`
}

// UpdateProductRequest 是 PUT /products/:id 的請求體。
// 所有欄位都是指標型別，允許 Partial Update（PATCH 語意）：
// 只傳入的欄位才會更新，nil 欄位保持原值（搭配 SQL 的 COALESCE）。
type UpdateProductRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	BasePrice   *float64 `json:"base_price" validate:"omitempty,gt=0"`
	Stock       *int     `json:"stock" validate:"omitempty,min=0"`
	ImageURL    *string  `json:"image_url"`
}

// ListProductsQuery 是 GET /products 的 Query String 參數。
// form tag 對應 Gin 的 ShouldBindQuery。
type ListProductsQuery struct {
	Character SanrioCharacter `form:"character"` // 可依角色篩選
	Page      int             `form:"page"`
	PageSize  int             `form:"page_size"`
}
