package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/internal/repository"
	"sanrio-auction-api/pkg/database"

	"github.com/jackc/pgx/v5"
)

var (
	ErrSoldOut           = errors.New("flash sale is sold out")
	ErrFlashSaleInactive = errors.New("flash sale is not currently active")
)

type FlashSaleService struct {
	db        *database.DB
	saleRepo  *repository.FlashSaleRepository
	orderRepo *repository.OrderRepository
}

func NewFlashSaleService(
	db *database.DB,
	saleRepo *repository.FlashSaleRepository,
	orderRepo *repository.OrderRepository,
) *FlashSaleService {
	return &FlashSaleService{db: db, saleRepo: saleRepo, orderRepo: orderRepo}
}

func (s *FlashSaleService) Create(ctx context.Context, req *model.CreateFlashSaleRequest) (*model.FlashSale, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, errors.New("invalid start_time, use RFC3339 format")
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, errors.New("invalid end_time, use RFC3339 format")
	}
	return s.saleRepo.Create(ctx, req, startTime, endTime)
}

func (s *FlashSaleService) GetByID(ctx context.Context, id int64) (*model.FlashSale, error) {
	return s.saleRepo.GetByID(ctx, id)
}

func (s *FlashSaleService) List(ctx context.Context) ([]*model.FlashSale, error) {
	return s.saleRepo.List(ctx)
}

// ================================================================
// PurchaseUnsafe — 危險版本，展示超賣（Overselling）Race Condition
// ================================================================
// 問題：兩個請求同時讀到 remaining_stock=1，
//       都通過 remaining_stock > 0 的檢查，
//       都執行扣減 → remaining_stock 變成 -1（超賣！）
//
// 可透過 RunConcurrentTest 的 unsafe 模式觀察此問題。
// ================================================================
func (s *FlashSaleService) PurchaseUnsafe(ctx context.Context, saleID, userID int64) (*model.Order, error) {
	// Step 1: 普通 SELECT，沒有鎖
	sale, err := s.saleRepo.GetByID(ctx, saleID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if now.Before(sale.StartTime) || now.After(sale.EndTime) {
		return nil, ErrFlashSaleInactive
	}

	// ⚠️  RACE WINDOW START ⚠️
	// 此處另一個 goroutine 可以讀到相同的 remaining_stock=1
	if sale.RemainingStock <= 0 {
		return nil, ErrSoldOut
	}
	// 兩個 goroutine 都能通過上面的 check，都執行扣減
	if err := s.saleRepo.DecrementStock(ctx, saleID); err != nil {
		return nil, err
	}
	// remaining_stock: 1 → 0 → -1（超賣！）
	// ⚠️  RACE WINDOW END ⚠️

	return s.orderRepo.Create(ctx, userID, sale.ProductID, 1, sale.SalePrice)
}

// ================================================================
// PurchaseWithLock — 安全版本，使用 SELECT FOR UPDATE 防止超賣
// ================================================================
// 解法：在 transaction 內鎖定 flash_sale 列，
//       確保同一時間只有一個 transaction 能讀取並修改庫存。
//       如果 remaining_stock=1，只有第一個 transaction 能成功，
//       後續 transaction 等待鎖後會讀到 remaining_stock=0 並返回 ErrSoldOut。
// ================================================================
func (s *FlashSaleService) PurchaseWithLock(ctx context.Context, saleID, userID int64) (*model.Order, error) {
	var order *model.Order

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// SELECT ... FOR UPDATE：鎖定 flash_sale 列
		sale, err := s.saleRepo.GetByIDForUpdate(ctx, tx, saleID)
		if err != nil {
			return err
		}

		now := time.Now()
		if now.Before(sale.StartTime) || now.After(sale.EndTime) {
			return ErrFlashSaleInactive
		}
		if sale.RemainingStock <= 0 {
			return ErrSoldOut
		}

		if err := s.saleRepo.DecrementStockTx(ctx, tx, saleID); err != nil {
			return err
		}

		order, err = s.orderRepo.CreateTx(ctx, tx, userID, sale.ProductID, 1, sale.SalePrice)
		return err
	})

	return order, err
}

// ================================================================
// RunConcurrentTest — goroutine 並發搶購測試
// ================================================================
// 同時啟動 goroutines 個 goroutine，每個都嘗試搶購同一個 flash sale。
// 透過 channel 收集每個 goroutine 的結果，最後回傳統計摘要。
//
// 使用方式：
//   safe=true  → 呼叫 PurchaseWithLock，庫存不會低於 0
//   safe=false → 呼叫 PurchaseUnsafe，可觀察 remaining_stock 變為負數
//
// 典型測試場景：建立一個 remaining_stock=10 的 flash sale，
// 同時發出 100 個 goroutine，觀察兩種模式下的結果差異。
// ================================================================
func (s *FlashSaleService) RunConcurrentTest(ctx context.Context, saleID, userID int64, goroutines int, useLock bool) (*model.ConcurrentTestResult, error) {
	// 使用獨立 context，避免 HTTP request context 在 goroutine 完成前被取消
	testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// buffered channel：讓每個 goroutine 發完後不需等待，避免 goroutine 洩漏
	type purchaseResult struct{ err error }
	results := make(chan purchaseResult, goroutines)

	start := time.Now()
	var wg sync.WaitGroup

	// 同時啟動 goroutines 個 goroutine
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			if useLock {
				_, err = s.PurchaseWithLock(testCtx, saleID, userID)
			} else {
				_, err = s.PurchaseUnsafe(testCtx, saleID, userID)
			}
			results <- purchaseResult{err}
		}()
	}

	// 等所有 goroutine 完成後關閉 channel
	wg.Wait()
	close(results)

	// 統計結果（channel 已關閉，range 不會阻塞）
	var successCount, failureCount int
	for r := range results {
		if r.err == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	// 查詢最終庫存，觀察是否發生超賣
	sale, err := s.saleRepo.GetByID(testCtx, saleID)
	if err != nil {
		return nil, err
	}

	mode := "unsafe"
	if useLock {
		mode = "safe"
	}

	return &model.ConcurrentTestResult{
		Mode:           mode,
		Goroutines:     goroutines,
		SuccessCount:   successCount,
		FailureCount:   failureCount,
		RemainingStock: sale.RemainingStock,
		DurationMs:     time.Since(start).Milliseconds(),
	}, nil
}
