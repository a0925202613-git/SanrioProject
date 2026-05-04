package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sanrio-auction-api/internal/model"
	"sanrio-auction-api/internal/repository"

	"github.com/redis/go-redis/v9"
)

const (
	productCacheTTL  = 5 * time.Minute
	productListKey   = "products:list"
	productKeyPrefix = "product:"
)

type ProductService struct {
	repo  *repository.ProductRepository
	cache *redis.Client
}

func NewProductService(repo *repository.ProductRepository, cache *redis.Client) *ProductService {
	return &ProductService{repo: repo, cache: cache}
}

func (s *ProductService) Create(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error) {
	p, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Del(ctx, productListKey)
	}
	return p, nil
}

// GetByID uses cache-aside: check Redis first, fallback to PostgreSQL.
func (s *ProductService) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	key := fmt.Sprintf("%s%d", productKeyPrefix, id)

	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, key).Result(); err == nil {
			var p model.Product
			if err := json.Unmarshal([]byte(cached), &p); err == nil {
				return &p, nil
			}
		}
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		if data, err := json.Marshal(p); err == nil {
			s.cache.Set(ctx, key, data, productCacheTTL)
		}
	}
	return p, nil
}

// List caches the first page of the full product list.
func (s *ProductService) List(ctx context.Context, query *model.ListProductsQuery) ([]*model.Product, error) {
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	if s.cache != nil && query.Character == "" && query.Page == 1 {
		if cached, err := s.cache.Get(ctx, productListKey).Result(); err == nil {
			var products []*model.Product
			if err := json.Unmarshal([]byte(cached), &products); err == nil {
				return products, nil
			}
		}
	}

	products, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	if s.cache != nil && query.Character == "" && query.Page == 1 {
		if data, err := json.Marshal(products); err == nil {
			s.cache.Set(ctx, productListKey, data, productCacheTTL)
		}
	}
	return products, nil
}

func (s *ProductService) Update(ctx context.Context, id int64, req *model.UpdateProductRequest) (*model.Product, error) {
	p, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Del(ctx, fmt.Sprintf("%s%d", productKeyPrefix, id))
		s.cache.Del(ctx, productListKey)
	}
	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if s.cache != nil {
		s.cache.Del(ctx, fmt.Sprintf("%s%d", productKeyPrefix, id))
		s.cache.Del(ctx, productListKey)
	}
	return nil
}
