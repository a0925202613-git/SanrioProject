package service

import (
	"context"
	"errors"
	"time"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/internal/repository"
	"sanrio-auction-api/pkg/middleware"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials 是登入失敗的統一錯誤，不區分「帳號不存在」或「密碼錯誤」，
// 防止攻擊者透過不同的錯誤訊息來枚舉系統中存在的帳號（User Enumeration）。
var ErrInvalidCredentials = errors.New("invalid email or password")

type AuthService struct {
	userRepo    *repository.UserRepository
	jwtSecret   string
	jwtExpireHr int
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, jwtExpireHr int) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret, jwtExpireHr: jwtExpireHr}
}

// Register 使用 bcrypt 雜湊密碼後存入資料庫。
// bcrypt.DefaultCost = 10，在安全性與雜湊速度之間取得平衡。
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Balance:  req.Balance,
		Role:     req.Role, // 從註冊請求中取得角色資訊,
	}
	return s.userRepo.Create(ctx, user, string(hash))
}

// Login 驗證 email 與密碼，成功後回傳 JWT token 與使用者資訊。
// 不論找不到帳號或密碼錯誤，一律回傳 ErrInvalidCredentials（避免 User Enumeration）。
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// bcrypt 的設計：同一個明文每次雜湊結果都不同（因為內嵌隨機 salt），
	// 所以無法用「重新雜湊再比較」的方式驗證。
	// CompareHashAndPassword 的做法是：從儲存的 hash 中取出當初用的 salt，
	// 把輸入的密碼用相同 salt 再雜湊一次，然後比對兩個 hash 是否一致。
	// 整個過程刻意設計得很慢（cost=10 代表執行 2^10=1024 次），讓暴力破解的成本極高。
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return &model.LoginResponse{Token: token, User: user}, nil
}

// GetMe 透過 JWT 中的 userID 查詢使用者資料，用於 GET /me。
// 回傳的 User 不含 password_hash（由 json:"-" 保護）。
func (s *AuthService) GetMe(ctx context.Context, userID int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// generateToken 產生一個 JWT（JSON Web Token）字串，結構是三段 Base64 用 "." 串接：
//
//	Header.Payload.Signature
//	  ↑              ↑           ↑
//	演算法資訊    放 userID    用 secret 簽名
//
// Payload（Claims）存放：
//   - UserID：讓後續 API 從 token 直接知道是誰，不需要再查資料庫
//   - ExpiresAt：token 的有效期限，過期後必須重新登入
//   - IssuedAt：token 的發行時間
//
// 簽名方式 HS256（HMAC-SHA256）：
//
//	把 Header+Payload 用 jwtSecret 做 HMAC 運算產生簽名。
//	驗證時只要用相同 secret 重新算一次，比對簽名是否一致即可。
//	沒有 secret 的人就算拿到 token、改掉 Payload，簽名也會對不上 → 偽造失敗。
func (s *AuthService) generateToken(user *model.User) (string, error) { // 改為傳入 user 物件
	claims := &middleware.Claims{
		UserID: user.ID,
		Role:   user.Role, // 將角色資訊放進 Token 載荷中
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpireHr) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
