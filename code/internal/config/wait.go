package config

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func WaitForPostgres(dsn string, retries int, interval time.Duration) error {
	var lastErr error
	for i := 0; i < retries; i++ {
		db, err := sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			_ = db.Close()
		}
		if err == nil {
			return nil
		}
		lastErr = err
		log.Printf("waiting for postgres (%d/%d): %v", i+1, retries, err)
		time.Sleep(interval)
	}
	return lastErr
}
