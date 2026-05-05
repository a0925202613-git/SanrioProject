package service

import (
	"context"
	"errors"
	"log"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/internal/repository"
	"sanrio-auction-api/pkg/database"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInsufficientStock   = errors.New("insufficient product stock")
	ErrInsufficientBalance = errors.New("insufficient user balance")
)

type OrderService struct {
	db          *database.DB
	productRepo *repository.ProductRepository
	userRepo    *repository.UserRepository
	orderRepo   *repository.OrderRepository
}

func NewOrderService(
	db *database.DB,
	productRepo *repository.ProductRepository,
	userRepo *repository.UserRepository,
	orderRepo *repository.OrderRepository,
) *OrderService {
	return &OrderService{db: db, productRepo: productRepo, userRepo: userRepo, orderRepo: orderRepo}
}

// ================================================================
// CreateOrder — 多表 Transaction 示範
// ================================================================
// 展示原子性（Atomicity）：以下 5 個步驟必須全部成功，
// 任一步驟失敗則整個 transaction 回滾，資料庫恢復原狀。
//
// 步驟：
//  1. 鎖定商品列 → 檢查庫存
//  2. 鎖定用戶列 → 檢查餘額
//  3. 扣除用戶餘額
//  4. 扣減商品庫存
//  5. 建立訂單記錄
//
// ================================================================
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, req *model.CreateOrderRequest) (*model.Order, error) {
	var order *model.Order

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// Step 1: 鎖定商品列（FOR UPDATE），防止其他 transaction 同時修改庫存
		product, err := s.productRepo.GetByIDForUpdate(ctx, tx, req.ProductID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrInsufficientStock
			}
			return err
		}
		if product.Stock < req.Quantity {
			return ErrInsufficientStock
		}

		totalPrice := product.BasePrice * float64(req.Quantity)

		// Step 2: 鎖定用戶列（FOR UPDATE），防止同時發起多筆訂單
		user, err := s.userRepo.GetByIDForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}
		if user.Balance < totalPrice {
			return ErrInsufficientBalance
		}

		// Step 3: 扣除餘額
		if err := s.userRepo.DeductBalanceTx(ctx, tx, userID, totalPrice); err != nil {
			return err
		}

		// Step 4: 扣減庫存
		if err := s.productRepo.DecrementStockTx(ctx, tx, req.ProductID, req.Quantity); err != nil {
			return err
		}

		// Step 5: 建立訂單
		order, err = s.orderRepo.CreateTx(ctx, tx, userID, req.ProductID, req.Quantity, product.BasePrice)
		return err
	})

	return order, err
}

func (s *OrderService) GetByID(ctx context.Context, id, userID int64) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, repository.ErrNotFound
	}
	return order, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, id, userID int64, status model.OrderStatus) (*model.Order, error) {
	// 1. 先更新訂單狀態
	order, err := s.orderRepo.UpdateStatus(ctx, id, userID, status)
	if err != nil {
		return nil, err
	}

	// 2. 檢查：如果訂單狀態剛變成 "completed"，則觸發 VIP 升級
	if status == model.OrderCompleted {
		// 呼叫我們剛改好的增量更新方法
		// 傳入訂單的 TotalPrice 來累加
		err := s.userRepo.UpdateVIP(ctx, userID, order.TotalPrice)
		if err != nil {
			// 這裡如果失敗，通常會記錄 log，但不一定要擋住訂單完成的回傳
			log.Printf("VIP 更新失敗: %v\n", err)
		}
	}

	return order, nil
}

func (s *OrderService) ListByUser(ctx context.Context, userID int64) ([]*model.Order, error) {
	return s.orderRepo.ListByUser(ctx, userID)
}

// BatchDeleteOrders 批次刪除訂單，僅刪除屬於 userID 的訂單，防止 IDOR。
func (s *OrderService) BatchDeleteOrders(ctx context.Context, ids []int64, userID int64) error {
	return s.orderRepo.BatchDelete(ctx, ids, userID)
}

// BatchUpdateOrders 批次更新訂單狀態，僅更新屬於 userID 的訂單，防止 IDOR。
func (s *OrderService) BatchUpdateOrders(ctx context.Context, ids []int64, userID int64, status model.OrderStatus) error {
	return s.orderRepo.BatchUpdateStatus(ctx, ids, userID, status)
}
