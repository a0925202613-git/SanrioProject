# Sanrio Auction API

三麗鷗商品電商拍賣系統後端 API，使用 Go + Gin + PostgreSQL + Redis 實作。

核心功能：

- **資料庫 Transaction**：多表原子操作
- **PostgreSQL 行級鎖**：`SELECT FOR UPDATE` vs 無鎖版本（展示 Race Condition）
- **Redis 快取**：Cache-Aside 模式（Redis 未安裝時自動略過，不影響啟動）

---

## 目錄

- [技術棧](#技術棧)
- [環境需求](#環境需求)
- [快速啟動](#快速啟動)
- [查看資料庫 (pgAdmin)](#查看資料庫-pgadmin)
- [查看 Redis](#查看-redis)
- [API 總覽](#api-總覽)
- [Postman 測試](#postman-測試)
- [專案架構](#專案架構)
- [核心功能說明](#核心功能說明)

---

## 技術棧

| 用途              | 套件                          |
| ----------------- | ----------------------------- |
| HTTP Framework    | `gin-gonic/gin`               |
| PostgreSQL Driver | `jackc/pgx/v5` + `pgxpool`    |
| Redis Client      | `redis/go-redis/v9`           |
| 設定管理          | `spf13/viper`                 |
| JWT 認證          | `golang-jwt/jwt/v5`           |
| 密碼加密          | `golang.org/x/crypto/bcrypt`  |
| 日誌              | `uber-go/zap`                 |
| 驗證              | `go-playground/validator/v10` |

---

## 環境需求

| 工具                                               | 版本  | 備註           |
| -------------------------------------------------- | ----- | -------------- |
| [Go](https://go.dev/dl/)                           | 1.21+ |                |
| [PostgreSQL](https://www.postgresql.org/download/) | 14+   | 本機安裝       |
| [Redis](https://redis.io/download/)                | 7+    | 本機安裝，選填 |
| [Postman](https://www.postman.com/downloads/)      | 任意  | API 測試       |

---

## 快速啟動

### Step 1：設定環境變數

複製 `.env.example` 為 `.env`，依實際情況修改：

```bash
cp .env.example .env
```

預設值如下，若 PostgreSQL 使用者名稱/密碼不同請修改：

```env
SERVER_PORT=9090

DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_DBNAME=sanrio_auction

REDIS_ADDR=localhost:6379

JWT_SECRET=change-this-to-a-long-random-secret
```

### Step 2：建立資料庫

在 PostgreSQL 建立資料庫（只需執行一次）：

```bash
psql -U postgres -c "CREATE DATABASE sanrio_auction;"
```

### Step 3：執行 Migration

```bash
make migrate
```

這會在 `sanrio_auction` 資料庫中建立所有資料表。所有語句都有 `IF NOT EXISTS`，重複執行安全。

### Step 4：安裝 Go 依賴

```bash
make tidy
```

### Step 5：啟動 API Server

```bash
make run
```

看到以下輸出即代表啟動成功：

```
🌸 Sanrio Auction API starting on :9090
```

> Redis 若未安裝，啟動時會顯示警告 `WARNING: Redis unavailable — caching disabled`，但 API 仍正常運作，僅快取功能停用。

### 驗證 API 是否運作

```bash
curl http://localhost:9090/api/v1/products
```

應回傳：

```json
{ "code": 0, "message": "success", "data": null }
```

---

## 查看資料庫 (pgAdmin)

### 新增伺服器連線

開啟 pgAdmin 應用程式，左側右鍵 **Servers** → **Register** → **Server...**

**General 頁籤：**
| 欄位 | 值 |
|------|------|
| Name | `Sanrio DB`（自訂即可）|

**Connection 頁籤：**
| 欄位 | 值 |
|------|------|
| Host name/address | `localhost` |
| Port | `5432` |
| Maintenance database | `sanrio_auction` |
| Username | `postgres` |
| Password | `postgres` |

### 查看資料表

連線成功後：

```
Servers > Sanrio DB > Databases > sanrio_auction > Schemas > public > Tables
```

可看到以下 4 張資料表：

- `users`
- `products`
- `flash_sales`
- `orders`

查詢資料：右鍵任一表 → **View/Edit Data** → **All Rows**

---

## 查看 Redis

### 使用 redis-cli

```bash
redis-cli

# 列出所有快取 key
KEYS *

# 查看商品快取（呼叫 GET /api/v1/products/:id 後才會出現）
GET product:1

# 查看商品列表快取
GET products:list
```

### 使用 RedisInsight（GUI）

下載 [RedisInsight](https://redis.io/insight/)，新增連線：

| 欄位 | 值          |
| ---- | ----------- |
| Host | `localhost` |
| Port | `6379`      |

> 第一次呼叫 `GET /api/v1/products/:id` 為 cache miss；第二次起即命中 Redis 快取（TTL 5 分鐘）。

---

## API 總覽

### 認證

| Method | Path                    | 說明           | Auth |
| ------ | ----------------------- | -------------- | ---- |
| POST   | `/api/v1/auth/register` | 註冊           | ❌   |
| POST   | `/api/v1/auth/login`    | 登入，取得 JWT | ❌   |

### 商品

| Method | Path                   | 說明                   | Auth |
| ------ | ---------------------- | ---------------------- | ---- |
| GET    | `/api/v1/products`     | 商品列表（Redis 快取） | ❌   |
| GET    | `/api/v1/products/:id` | 商品詳情（Redis 快取） | ❌   |
| POST   | `/api/v1/products`     | 建立商品               | ✅   |
| PUT    | `/api/v1/products/:id` | 更新商品               | ✅   |
| DELETE | `/api/v1/products/:id` | 刪除商品               | ✅   |

### 圖片上傳

| Method | Path             | 說明                                 | Auth |
| ------ | ---------------- | ------------------------------------ | ---- |
| POST   | `/api/v1/upload` | 上傳商品圖片（jpg/png/webp，限 5MB） | ✅   |

### 限時特賣

| Method | Path                                      | 說明                       | Auth |
| ------ | ----------------------------------------- | -------------------------- | ---- |
| GET    | `/api/v1/flash-sales`                     | 特賣列表                   | ❌   |
| GET    | `/api/v1/flash-sales/:id`                 | 特賣詳情                   | ❌   |
| POST   | `/api/v1/flash-sales`                     | 建立特賣                   | ✅   |
| POST   | `/api/v1/flash-sales/:id/purchase/safe`   | 搶購（FOR UPDATE，防超賣） | ✅   |
| POST   | `/api/v1/flash-sales/:id/purchase/unsafe` | 搶購（無鎖，超賣 Demo）    | ✅   |

### 訂單

| Method | Path                          | 說明                         | Auth |
| ------ | ----------------------------- | ---------------------------- | ---- |
| POST   | `/api/v1/orders`              | 建立訂單（多表 Transaction） | ✅   |
| GET    | `/api/v1/orders`              | 我的訂單列表                 | ✅   |
| GET    | `/api/v1/orders/:id`          | 訂單詳情                     | ✅   |
| PATCH  | `/api/v1/orders/:id/status`   | 更新訂單狀態                 | ✅   |

---

## Postman 測試

### 匯入 Collection

1. 開啟 Postman
2. 點選左上角 **Import**
3. 選擇專案根目錄的 `postman_collection.json`
4. 匯入完成後可看到 **Sanrio Auction API** collection

### Collection 變數

| 變數            | 預設值                  | 說明               |
| --------------- | ----------------------- | ------------------ |
| `base_url`      | `http://localhost:9090` | API 伺服器位址     |
| `token`         | _(空)_                  | 登入後自動填入 JWT |
| `product_id`    | `1`                     | 測試用商品 ID      |
| `auction_id`    | `1`                     | 測試用拍賣 ID      |
| `flash_sale_id` | `1`                     | 測試用特賣 ID      |
| `order_id`      | `1`                     | 測試用訂單 ID      |

### 建議測試順序

```
1. Auth / Register                       → 建立測試帳號
2. Auth / Login                          → 自動存入 {{token}}
3. Products / Create                     → 建立三麗鷗商品（可附圖片）
4. Products / List                       → 第一次 cache miss
5. Products / Get by ID                  → 第二次命中 Redis 快取
6. Flash Sales / Create                  → 建立限時特賣
7. Flash Sales / Purchase (Safe)         → 安全搶購（FOR UPDATE）
8. Flash Sales / Purchase (Unsafe) x N  → 觀察超賣（remaining_stock < 0）
9. Orders / Create                       → 建立訂單（觀察 balance & stock 變化）
```

---

## 專案架構

### 分層原則

```
HTTP Request
    │
    ▼
Handler          解析請求參數、驗證格式、呼叫 Service
    │
    ▼
Service          業務邏輯：Transaction、行鎖、Redis 快取
    │
    ▼
Repository       純 SQL 操作，提供 Tx（帶鎖）與一般兩種版本
    │
    ▼
PostgreSQL / Redis
```

### 目錄結構

```
sanrio-auction-api/
│
├── main.go                             # 程式進入點：初始化 DB / Redis / Service / Handler，啟動 Gin Server
├── routes.go                           # 路由註冊：定義 Handlers struct，將所有 endpoint 集中管理
├── go.mod / go.sum                     # Go module 依賴宣告
├── Makefile                            # 常用指令：migrate / run / build / test / frontend-dev
├── postman_collection.json             # Postman 測試 Collection（含自動存 token 腳本）
│
├── .env.example                        # 環境變數範本（複製為 .env 後填入實際值）
├── .vscode/
│   └── launch.json                     # VS Code Delve 除錯設定（F5 啟動）
│
├── config/
│   ├── config.go                       # Viper 設定載入，優先順序：.env > config.yaml > 預設值
│   └── config.yaml                     # 預設設定（port、DB、Redis、JWT 有效時間）
│
├── internal/                           # 應用程式核心，不對外暴露
│   │
│   ├── handler/                        # HTTP 層：只負責「收請求、回回應」，不含業務邏輯
│   │   ├── auth_handler.go             # POST /auth/register、POST /auth/login、GET /me
│   │   ├── product_handler.go          # CRUD /products、/products/:id
│   │   ├── flashsale_handler.go        # /flash-sales CRUD、/purchase/safe|unsafe、/concurrent-test
│   │   ├── order_handler.go            # POST /orders、GET /orders、GET /orders/:id
│   │   └── upload_handler.go           # POST /upload：接收圖片、儲存至 static/uploads/、回傳路徑
│   │
│   ├── service/                        # 業務邏輯層：Transaction、鎖策略、快取都在這裡決定
│   │   ├── auth_service.go             # bcrypt 密碼雜湊、JWT 簽發（HS256）、防 User Enumeration
│   │   ├── product_service.go          # Cache-Aside：Redis HIT 直接回傳，MISS 查 DB 後寫入快取
│   │   ├── flashsale_service.go        # ★ 核心 Demo：
│   │   │                               #   PurchaseWithLock()  — BEGIN + SELECT FOR UPDATE + COMMIT
│   │   │                               #   PurchaseUnsafe()    — 無鎖，可觀察 remaining_stock < 0
│   │   │                               #   RunConcurrentTest() — 啟動 N 個 goroutine 同時搶購
│   │   └── order_service.go            # ★ 核心 Demo：
│   │                                   #   5 步原子 Transaction（鎖商品→鎖用戶→扣餘額→扣庫存→建訂單）
│   │
│   ├── repository/                     # 資料存取層：純 SQL，不含業務判斷
│   │   ├── errors.go                   # 共用錯誤定義（ErrNotFound）
│   │   ├── user_repository.go          # Create、GetByEmail、GetByID、GetByIDForUpdate、DeductBalance
│   │   ├── product_repository.go       # CRUD、List（含角色篩選）、GetByIDForUpdate、DecrementStock
│   │   ├── flashsale_repository.go     # Create、GetByID、GetByIDForUpdate、DecrementStock、List
│   │   └── order_repository.go         # CreateTx、ListByUserID、GetByID
│   │
│   └── model/                          # 資料結構定義（domain model + request/response DTO）
│       ├── user.go                     # User、RegisterRequest、LoginRequest、LoginResponse
│       ├── product.go                  # Product、CreateProductRequest、UpdateProductRequest（Partial Update）、ListProductsQuery
│       ├── flashsale.go                # FlashSale、CreateFlashSaleRequest、ConcurrentTestRequest、ConcurrentTestResult
│       └── order.go                    # Order、CreateOrderRequest
│
├── pkg/                                # 可跨專案重用的通用工具
│   ├── database/
│   │   └── postgres.go                 # pgxpool 連線池、WithTx() helper（自動 Rollback on error）
│   ├── redis/
│   │   └── redis.go                    # go-redis 客戶端；Redis 不可用時回傳 nil（noopLogger 避免雜訊）
│   ├── middleware/
│   │   └── auth.go                     # JWT 驗證 Middleware：解析 Bearer token，將 userID 注入 gin.Context
│   ├── logger/
│   │   └── logger.go                   # zap 日誌初始化（debug mode 輸出 console，production 輸出 JSON）
│   └── response/
│       └── response.go                 # 統一回應格式：{ "code": 0, "message": "...", "data": ... }
│
├── migrations/
│   └── init.sql                        # 所有資料表建立語句，含 IF NOT EXISTS（可重複執行）
│                                       # 建立順序：users → products → flash_sales → orders
│
└── static/
    └── uploads/                        # 商品圖片儲存目錄（由 main.go 自動建立，已加入 .gitignore）
```

---

## 核心功能說明

### 1. 超賣問題（Flash Sale）

檔案：[`internal/service/flashsale_service.go`](internal/service/flashsale_service.go)

```
PurchaseUnsafe()      →  remaining_stock 可能變成負數（超賣）
PurchaseWithLock()    →  FOR UPDATE 確保 remaining_stock 最少為 0
```

### 2. 多表 Transaction（原子性）

檔案：[`internal/service/order_service.go`](internal/service/order_service.go)

```
BEGIN TRANSACTION
  ├── SELECT products WHERE id=? FOR UPDATE  （鎖商品，確認庫存）
  ├── SELECT users WHERE id=? FOR UPDATE     （鎖用戶，確認餘額）
  ├── UPDATE users SET balance = balance - ?  （扣餘額）
  ├── UPDATE products SET stock = stock - ?   （扣庫存）
  └── INSERT INTO orders ...                  （建訂單）
COMMIT  ← 全部成功才提交，任一失敗全部回滾
```

### 3. Redis Cache-Aside

檔案：[`internal/service/product_service.go`](internal/service/product_service.go)

```
GET /products/:id
  ├── Redis HIT  → 直接回傳（不查 DB）
  └── Redis MISS → 查 PostgreSQL → 寫入 Redis（TTL 5分鐘） → 回傳

PUT/DELETE /products/:id
  └── 更新 DB 後，刪除對應 Redis key（Cache Invalidation）
```
