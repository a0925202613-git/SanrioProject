package repository

import (
	"context"
	"errors"
	"time"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/pkg/database"

	"github.com/jackc/pgx/v5"
)

type FlashSaleRepository struct {
	db *database.DB
}

func NewFlashSaleRepository(db *database.DB) *FlashSaleRepository {
	return &FlashSaleRepository{db: db}
}

// Create 新增限時特賣，remaining_stock 初始值等於 total_stock。
// startTime / endTime 由 Service 層解析 RFC3339 字串後傳入。
func (r *FlashSaleRepository) Create(ctx context.Context, req *model.CreateFlashSaleRequest, startTime, endTime time.Time) (*model.FlashSale, error) {
	query := `
		INSERT INTO flash_sales (product_id, sale_price, total_stock, remaining_stock, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, product_id, sale_price, total_stock, remaining_stock, start_time, end_time, created_at
	`
	row := r.db.Pool.QueryRow(ctx, query,
		req.ProductID, req.SalePrice, req.TotalStock, req.TotalStock, startTime, endTime,
	)

	var fs model.FlashSale
	if err := row.Scan(&fs.ID, &fs.ProductID, &fs.SalePrice, &fs.TotalStock, &fs.RemainingStock, &fs.StartTime, &fs.EndTime, &fs.CreatedAt); err != nil {
		return nil, err
	}
	return &fs, nil
}

// GetByID 一般讀取，不加鎖，用於查詢特賣詳情。
func (r *FlashSaleRepository) GetByID(ctx context.Context, id int64) (*model.FlashSale, error) {
	query := `SELECT id, product_id, sale_price, total_stock, remaining_stock, start_time, end_time, created_at FROM flash_sales WHERE id = $1`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var fs model.FlashSale
	if err := row.Scan(&fs.ID, &fs.ProductID, &fs.SalePrice, &fs.TotalStock, &fs.RemainingStock, &fs.StartTime, &fs.EndTime, &fs.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &fs, nil
}

// GetByIDForUpdate 在 transaction 內執行 SELECT ... FOR UPDATE，
// 鎖定 flash_sale 列，這是防止超賣的關鍵：
// 並發的搶購請求無法同時讀取 remaining_stock，只能依序執行，
// 確保庫存檢查（> 0）與扣減（- 1）之間不會被其他請求插入。
func (r *FlashSaleRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id int64) (*model.FlashSale, error) {
	query := `
		SELECT id, product_id, sale_price, total_stock, remaining_stock, start_time, end_time
		FROM flash_sales WHERE id = $1
		FOR UPDATE
	`
	row := tx.QueryRow(ctx, query, id)

	var fs model.FlashSale
	if err := row.Scan(&fs.ID, &fs.ProductID, &fs.SalePrice, &fs.TotalStock, &fs.RemainingStock, &fs.StartTime, &fs.EndTime); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &fs, nil
}

// DecrementStock 不帶 transaction，供 PurchaseUnsafe 使用（展示超賣問題）。
func (r *FlashSaleRepository) DecrementStock(ctx context.Context, id int64) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE flash_sales SET remaining_stock = remaining_stock - 1 WHERE id = $1`, id)
	return err
}

// DecrementStockTx 在 transaction 內扣減庫存，供 PurchaseWithLock 使用（安全版本）。
func (r *FlashSaleRepository) DecrementStockTx(ctx context.Context, tx pgx.Tx, id int64) error {
	_, err := tx.Exec(ctx, `UPDATE flash_sales SET remaining_stock = remaining_stock - 1 WHERE id = $1`, id)
	return err
}

// List 只回傳 end_time > NOW() 的活動，過期的特賣不顯示。
// 依 start_time 升冪排列，即將開始的活動優先顯示。
func (r *FlashSaleRepository) List(ctx context.Context) ([]*model.FlashSale, error) {
	query := `
		SELECT id, product_id, sale_price, total_stock, remaining_stock, start_time, end_time, created_at
		FROM flash_sales
		WHERE end_time > NOW()
		ORDER BY start_time ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []*model.FlashSale
	for rows.Next() {
		var fs model.FlashSale
		if err := rows.Scan(&fs.ID, &fs.ProductID, &fs.SalePrice, &fs.TotalStock, &fs.RemainingStock, &fs.StartTime, &fs.EndTime, &fs.CreatedAt); err != nil {
			return nil, err
		}
		sales = append(sales, &fs)
	}
	return sales, nil
}
