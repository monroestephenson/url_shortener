package db

import (
	"database/sql"
	"time"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"url_shortener/internal/logger"
)

// NewMySQLDB returns a new MySQL db connection.
// Example DSN: username:password@tcp(127.0.0.1:3306)/url_shortener?parseTime=true
func NewMySQLDB(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error
	
	// Try to connect with retries
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			logger.GetLogger().Error("Failed to open MySQL connection",
				zap.Error(err),
				zap.Int("attempt", i+1),
			)
			time.Sleep(time.Second * 5)
			continue
		}

		// Test the connection
		err = db.Ping()
		if err == nil {
			break
		}

		logger.GetLogger().Error("Failed to ping MySQL",
			zap.Error(err),
			zap.Int("attempt", i+1),
		)
		time.Sleep(time.Second * 5)
	}

	if err != nil {
		return nil, err
	}

	// Create schema if not exists
	if err = createSchema(db); err != nil {
		return nil, err
	}

	return db, nil
}

// createSchema creates the necessary database tables if they don't exist.
func createSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		`CREATE TABLE IF NOT EXISTS short_urls (
			id INT AUTO_INCREMENT PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL UNIQUE,
			original_url TEXT NOT NULL,
			access_count INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			user_id INT,
			INDEX idx_short_code (short_code),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
