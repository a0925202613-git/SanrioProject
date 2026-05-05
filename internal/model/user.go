package model

import "time"

// User 對應 users 資料表，代表系統中的一個帳號。
// PasswordHash 使用 json:"-" 確保序列化時永不對外暴露密碼雜湊。
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`       // 永不回傳給前端
	Balance      float64   `json:"balance"` // 帳戶餘額，用於購買商品與參與拍賣
	Role         string    `json:"role"`    // 例如 "user" 或 "admin"，用於權限控制
	TotalSpent   int       `json:"total_spent"`
	VipLevel     string    `json:"vip_level"`
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest 是 POST /auth/register 的請求體。
// Balance 允許初始存入，方便測試時直接設定餘額。
type RegisterRequest struct {
	Username string  `json:"username" validate:"required,min=3,max=50"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=6"`
	Balance  float64 `json:"balance" validate:"min=0"` // 選填，預設 0
	Role     string  `json:"role"`                     // 例如 "user" 或 "admin"，用於權限控制
}

// LoginRequest 是 POST /auth/login 的請求體。
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse 是登入成功後回傳的資料，包含 JWT token 與使用者資訊。
// User.PasswordHash 因 json:"-" 不會出現在回應中。
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
