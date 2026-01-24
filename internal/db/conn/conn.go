package conn

import (
	"database/sql"
	"fmt"
	"wb-project/internal/config"

	_ "github.com/lib/pq"
)

func Connection(conf *config.DBConfig) (*sql.DB, error) {
	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Password,
		conf.DBName)
	db, err := sql.Open("postgres", dns)
	if err != nil {
		return nil, fmt.Errorf("неудалось установить сооединение. Ошибка: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("не удалось достучаться до БД. Ошибка: %v", err)
	}

	return db, nil
}
