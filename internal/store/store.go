package store

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"net/url"
	"time"
)

func Open(ctx context.Context, rawURL string) (*sql.DB, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	q := parsed.Query()
	q.Set("timezone", "UTC")
	parsed.RawQuery = q.Encode()
	db, err := sql.Open("postgres", parsed.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
