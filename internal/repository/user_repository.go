package repository

import (
	"context"
	"errors"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/pkg/database"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 新增一筆使用者記錄，passwordHash 已由 Service 層完成 bcrypt 雜湊。
// RETURNING 子句讓 PostgreSQL 直接回傳插入後的資料，避免額外一次 SELECT。
func (r *UserRepository) Create(ctx context.Context, user *model.User, passwordHash string) (*model.User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, balance)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, balance, created_at
	`
	row := r.db.Pool.QueryRow(ctx, query, user.Username, user.Email, passwordHash, user.Balance)

	var u model.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Balance, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByEmail 查詢登入用，因此需要回傳 password_hash 以進行 bcrypt 比對。
// 其他查詢（如 GetByID）不回傳 password_hash，降低密碼外洩風險。
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, username, email, password_hash, balance, created_at FROM users WHERE email = $1`
	row := r.db.Pool.QueryRow(ctx, query, email)

	var u model.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Balance, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// GetByID 查詢一般用途的使用者資料，不包含 password_hash。
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, username, email, balance, created_at FROM users WHERE id = $1`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var u model.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Balance, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// GetByIDForUpdate 在 transaction 內執行 SELECT ... FOR UPDATE，
// 鎖定使用者列，防止同一帳號同時發起多筆訂單造成餘額超扣。
// 只選取建立訂單時需要的欄位（balance）。
func (r *UserRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id int64) (*model.User, error) {
	query := `SELECT id, username, email, balance FROM users WHERE id = $1 FOR UPDATE`
	row := tx.QueryRow(ctx, query, id)

	var u model.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// DeductBalanceTx 在 transaction 內扣除餘額，必須在 GetByIDForUpdate 之後呼叫，
// 確保餘額檢查與扣減是原子操作，不會發生 TOCTOU（time-of-check/time-of-use）問題。
func (r *UserRepository) DeductBalanceTx(ctx context.Context, tx pgx.Tx, userID int64, amount float64) error {
	_, err := tx.Exec(ctx, `UPDATE users SET balance = balance - $1 WHERE id = $2`, amount, userID)
	return err
}

// UpdateVIP 更新使用者的累積消費額與 VIP 等級
func (r *UserRepository) UpdateVIP(ctx context.Context, userID int64, amount float64) error { // 直接在 SQL 裡加總，並用 CASE 判斷等級
	query := `
		UPDATE users 
		SET total_spent = total_spent + $1,
		    vip_level = CASE 
		        WHEN total_spent + $1 >= 5000 THEN 'gold'
		        WHEN total_spent + $1 >= 1000 THEN 'silver'
		        ELSE 'normal'
		    END
		WHERE id = $2
	`
	_, err := r.db.Pool.Exec(ctx, query, amount, userID)
	return err
}
