package main

import (
	"fmt"
	"log"
	"os"

	"sanrio-auction-api/config"
	"sanrio-auction-api/internal/handler"
	"sanrio-auction-api/internal/repository"
	"sanrio-auction-api/internal/service"
	"sanrio-auction-api/pkg/database"
	"sanrio-auction-api/pkg/logger"
	redispkg "sanrio-auction-api/pkg/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := logger.Init(cfg.Server.Mode); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Log.Sync() //nolint:errcheck

	db, err := database.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	redisClient := redispkg.NewClient(cfg.Redis)

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	flashSaleRepo := repository.NewFlashSaleRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// --- Services ---
	authSvc := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHours)
	productSvc := service.NewProductService(productRepo, redisClient)
	flashSaleSvc := service.NewFlashSaleService(db, flashSaleRepo, orderRepo)
	orderSvc := service.NewOrderService(db, productRepo, userRepo, orderRepo)

	// --- 建立圖片上傳目錄 ---
	const uploadDir = "static/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("failed to create upload dir: %v", err)
	}

	// --- Handlers ---
	h := &Handlers{
		Auth:      handler.NewAuthHandler(authSvc),
		Product:   handler.NewProductHandler(productSvc),
		FlashSale: handler.NewFlashSaleHandler(flashSaleSvc),
		Order:     handler.NewOrderHandler(orderSvc),
		Upload:    handler.NewUploadHandler(uploadDir),
	}

	// --- Router ---
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	RegisterRoutes(r, cfg.JWT.Secret, h)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("🌸 Sanrio Auction API starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
