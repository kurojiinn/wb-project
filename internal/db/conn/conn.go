package conn

import (
	"database/sql"
	"fmt"
	"os"
	"wb-project/internal/config"

	_ "github.com/lib/pq"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func Connection(conf *config.DBConfig) (*sql.DB, error) {
	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Password,
		conf.DBName)
	db, err := otelsql.Open("postgres", dns,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBName(os.Getenv("db_name")),
	)
	if err != nil {
		return nil, fmt.Errorf("неудалось установить сооединение. Ошибка: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("не удалось достучаться до БД. Ошибка: %v", err)
	}

	return db, nil
}
