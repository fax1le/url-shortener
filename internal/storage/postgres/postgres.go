package postgres

import (
	"time"
	"database/sql"
	"fmt"
	"url-shortener/internal/config"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db  *sql.DB
	Cfg *config.Config
}

func StartDB(cfg *config.Config) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
	)

	DB, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	DB.SetMaxOpenConns(100)
	DB.SetMaxIdleConns(20)
	DB.SetConnMaxLifetime(time.Minute * 5)

	err = DB.Ping()

	return &PostgresDB{
		db:  DB,
		Cfg: cfg,
	}, err
}

func (d *PostgresDB) Close() error {
	err := d.db.Close()
	return err
}
