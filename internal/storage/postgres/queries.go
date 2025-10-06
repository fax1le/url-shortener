package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (d *PostgresDB) StoreUrl(ctx context.Context, longUrl string, slug string, expiration time.Duration) error {
	res, err := d.db.ExecContext(ctx, `
		INSERT INTO urls (long_url, slug, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (slug) DO NOTHING`,
		longUrl,
		slug,
		time.Now(),
		time.Now().Add(expiration),
	)

	rows, _ := res.RowsAffected()

	if rows == 0 {
		return errors.New("Slug exists")
	}

	return err
}

func (d *PostgresDB) SlugExists(ctx context.Context, key string) (bool, error) {
	i := 0

	row := d.db.QueryRowContext(ctx, "SELECT 1 FROM urls WHERE slug = $1", key)

	err := row.Scan(&i)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *PostgresDB) GetUrl(ctx context.Context, key string) (string, error) {
	longUrl := ""

	row := d.db.QueryRowContext(ctx, "SELECT long_url FROM urls WHERE slug = $1", key)

	err := row.Scan(&longUrl)

	return longUrl, err
}

func (d *PostgresDB) CleanUp(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, "TRUNCATE TABLE urls RESTART IDENTITY")
	return err
}

func (d *PostgresDB) ExpireUrls(ctx context.Context) (int64, error) {
	res, err := d.db.ExecContext(ctx, "DELETE FROM urls WHERE expires_at < NOW()")

	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()

	return rowsAffected, err
}

func (d *PostgresDB) StoreClicks(ctx context.Context, query string, args ...any) error {
	tx, err := d.db.BeginTx(ctx, nil)

	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return err
	}

	return tx.Commit()
}
