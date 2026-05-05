package repository

import "errors"

var (
	// ErrNotFound 在 SQL 查詢回傳 pgx.ErrNoRows 時使用，
	// 讓 Service 層不需要依賴 pgx 套件即可判斷「找不到資料」。
	ErrNotFound = errors.New("record not found")

	// ErrConflict 用於 UNIQUE 約束衝突，例如 email 或 username 重複。
	ErrConflict = errors.New("record already exists")

	// ErrHasRelatedRecords 用於 FK 約束衝突，例如刪除有訂單的商品。
	ErrHasRelatedRecords = errors.New("has related records")
)
