.PHONY: migrate run build test tidy frontend-install frontend-dev

# 對本機 PostgreSQL 執行 migrations（需先確認 postgres 已啟動，且 psql 在 PATH 中）
migrate:
	psql -U tosiatung -d sanrio_auction -f migrations/init.sql

run:
	go run .

build:
	go build -o bin/sanrio-api.exe .

test:
	go test ./... -race -count=1 -v

tidy:
	go mod tidy

# React frontend（需先執行 frontend-install 安裝 node_modules）
frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev
