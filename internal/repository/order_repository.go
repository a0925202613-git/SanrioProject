package repository

import (
	"context"
	"errors"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/pkg/database"

	"github.com/jackc/pgx/v5"
)

type OrderRepository struct {
	db *database.DB
}

func NewOrderRepository(db *database.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create 不帶 transaction，供閃購 PurchaseUnsafe 使用。
func (r *OrderRepository) Create(ctx context.Context, userID, productID int64, quantity int, unitPrice float64) (*model.Order, error) {
	return r.insertOrder(ctx, r.db.Pool.QueryRow, userID, productID, quantity, unitPrice)
}

// CreateTx 在 transaction 內執行，供 PurchaseWithLock 與 CreateOrder 使用。
func (r *OrderRepository) CreateTx(ctx context.Context, tx pgx.Tx, userID, productID int64, quantity int, unitPrice float64) (*model.Order, error) {
	return r.insertOrder(ctx, tx.QueryRow, userID, productID, quantity, unitPrice)
}

// insertOrder 是 Create 與 CreateTx 的共用實作，抽象了有無 transaction 的差異。
// total_price 在這裡計算（Go 層），而非讓 SQL 計算，方便未來加入折扣邏輯。
// status 固定為 pending，代表訂單剛建立，尚未付款。
func (r *OrderRepository) insertOrder(ctx context.Context, queryRow func(context.Context, string, ...interface{}) pgx.Row, userID, productID int64, quantity int, unitPrice float64) (*model.Order, error) {
	query := `
		INSERT INTO orders (user_id, product_id, quantity, unit_price, total_price, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, product_id, quantity, unit_price, total_price, status, created_at
	`
	totalPrice := unitPrice * float64(quantity)
	row := queryRow(ctx, query, userID, productID, quantity, unitPrice, totalPrice, model.OrderPending)

	var o model.Order
	if err := row.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

// GetByID 查詢單一訂單，Service 層另行驗證 UserID 確保只能查自己的訂單。
func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	query := `
		SELECT id, user_id, product_id, quantity, unit_price, total_price, status, created_at
		FROM orders WHERE id = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var o model.Order
	if err := row.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

// UpdateStatus 更新訂單狀態，只允許更新自己的訂單（userID 條件防止 IDOR）。
// 回傳更新後的完整訂單；若 id+userID 不匹配則回傳 ErrNotFound。
func (r *OrderRepository) UpdateStatus(ctx context.Context, id, userID int64, status model.OrderStatus) (*model.Order, error) {
	query := `
		UPDATE orders SET status = $1
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, product_id, quantity, unit_price, total_price, status, created_at
	`
	row := r.db.Pool.QueryRow(ctx, query, status, id, userID)

	var o model.Order
	if err := row.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

// ListByUser 查詢指定使用者的所有訂單，依建立時間降冪（最新訂單在最前）。
func (r *OrderRepository) ListByUser(ctx context.Context, userID int64) ([]*model.Order, error) {
	query := `
		SELECT id, user_id, product_id, quantity, unit_price, total_price, status, created_at
		FROM orders WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

// BatchUpdateStatus 批次更新多筆訂單狀態。
// 使用 ANY($2) 可以一次處理多個 ID，效率比跑迴圈執行多個 UPDATE 高得多。
func (r *OrderRepository) BatchUpdateStatus(ctx context.Context, ids []int64, userID int64, status model.OrderStatus) error {
	query := `
        UPDATE orders 
        SET status = $1 
        WHERE id = ANY($2) AND user_id = $3
    `
	_, err := r.db.Pool.Exec(ctx, query, status, ids, userID)
	return err
}

// BatchDelete 批次刪除多筆訂單。
// 同樣加上 user_id 條件，確保使用者只能刪除自己的訂單（安全檢查）。
func (r *OrderRepository) BatchDelete(ctx context.Context, ids []int64, userID int64) error {
	query := `
        DELETE FROM orders 
        WHERE id = ANY($1) AND user_id = $2
    `
	_, err := r.db.Pool.Exec(ctx, query, ids, userID)
	return err
}
