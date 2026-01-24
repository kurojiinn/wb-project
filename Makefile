# Переменные для удобства
DB_DSN=postgres://user:pass@localhost:5432/order_db?sslmode=disable
MIGRATIONS_DIR=./migrations

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" status