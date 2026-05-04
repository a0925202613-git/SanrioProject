package repository

import (
	"context"
	"errors"
	"fmt"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/pkg/database"

	"github.com/jackc/pgx/v5"
)

type ProductRepository struct {
	db *database.DB
}

func NewProductRepository(db *database.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create 新增商品並透過 RETURNING 直接取得完整記錄（含資料庫產生的 id 與 created_at）。
func (r *ProductRepository) Create(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error) {
	query := `
		INSERT INTO products (name, character, description, base_price, stock, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, character, description, base_price, stock, image_url, created_at
	`
	row := r.db.Pool.QueryRow(ctx, query,
		req.Name, req.Character, req.Description, req.BasePrice, req.Stock, req.ImageURL,
	)

	var p model.Product
	if err := row.Scan(&p.ID, &p.Name, &p.Character, &p.Description, &p.BasePrice, &p.Stock, &p.ImageURL, &p.CreatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByID 查詢單一商品，用於一般讀取與 Redis cache miss 後的資料庫查詢。
func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	query := `SELECT id, name, character, description, base_price, stock, image_url, created_at FROM products WHERE id = $1`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var p model.Product
	if err := row.Scan(&p.ID, &p.Name, &p.Character, &p.Description, &p.BasePrice, &p.Stock, &p.ImageURL, &p.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// GetByIDForUpdate 在 transaction 內執行 SELECT ... FOR UPDATE，
// 鎖定商品列，防止建立訂單時庫存被其他 transaction 同時扣減。
// 只選取訂單邏輯需要的欄位，減少鎖競爭時的資料傳輸量。
func (r *ProductRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id int64) (*model.Product, error) {
	query := `SELECT id, name, character, base_price, stock FROM products WHERE id = $1 FOR UPDATE`
	row := tx.QueryRow(ctx, query, id)

	var p model.Product
	if err := row.Scan(&p.ID, &p.Name, &p.Character, &p.BasePrice, &p.Stock); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// List 支援依角色篩選與分頁。
// 使用動態拼接 WHERE 子句（argIdx 遞增），避免固定佔位符導致的 SQL 錯誤。
// WHERE 1=1 是讓後續條件可以統一用 AND 拼接的慣用寫法。
func (r *ProductRepository) List(ctx context.Context, q *model.ListProductsQuery) ([]*model.Product, error) {
	baseQuery := `SELECT id, name, character, description, base_price, stock, image_url, created_at FROM products WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if q.Character != "" {
		baseQuery += fmt.Sprintf(" AND character = $%d", argIdx)
		args = append(args, q.Character)
		argIdx++
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, q.PageSize, (q.Page-1)*q.PageSize)

	rows, err := r.db.Pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Character, &p.Description, &p.BasePrice, &p.Stock, &p.ImageURL, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}
	return products, nil
}

// Update 使用 COALESCE 實現 Partial Update：
// COALESCE($1, name) 的意思是「如果 $1 為 NULL，就保留原本的 name」，
// 對應 UpdateProductRequest 中的指標型別欄位（nil = 不更新）。
func (r *ProductRepository) Update(ctx context.Context, id int64, req *model.UpdateProductRequest) (*model.Product, error) {
	query := `
		UPDATE products SET
			name        = COALESCE($1, name),
			description = COALESCE($2, description),
			base_price  = COALESCE($3, base_price),
			stock       = COALESCE($4, stock),
			image_url   = COALESCE($5, image_url)
		WHERE id = $6
		RETURNING id, name, character, description, base_price, stock, image_url, created_at
	`
	row := r.db.Pool.QueryRow(ctx, query,
		req.Name, req.Description, req.BasePrice, req.Stock, req.ImageURL, id,
	)

	var p model.Product
	if err := row.Scan(&p.ID, &p.Name, &p.Character, &p.Description, &p.BasePrice, &p.Stock, &p.ImageURL, &p.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

// Delete 刪除商品，並透過 RowsAffected() 判斷商品是否存在。
func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DecrementStockTx 在 transaction 內扣減庫存，必須在 GetByIDForUpdate 之後呼叫，
// 確保庫存檢查與扣減是原子操作。
func (r *ProductRepository) DecrementStockTx(ctx context.Context, tx pgx.Tx, id int64, quantity int) error {
	_, err := tx.Exec(ctx, `UPDATE products SET stock = stock - $1 WHERE id = $2`, quantity, id)
	return err
}
