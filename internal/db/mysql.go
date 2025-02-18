package db

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

// NewMySQLDB returns a new MySQL db connection.
// Example DSN: username:password@tcp(127.0.0.1:3306)/url_shortener?parseTime=true
func NewMySQLDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    // Test the connection
    if err = db.Ping(); err != nil {
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
    query := `
    CREATE TABLE IF NOT EXISTS short_urls (
        id INT AUTO_INCREMENT PRIMARY KEY,
        short_code VARCHAR(10) NOT NULL UNIQUE,
        original_url TEXT NOT NULL,
        access_count INT NOT NULL DEFAULT 0,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL,
        INDEX idx_short_code (short_code)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `

    _, err := db.Exec(query)
    return err
}
