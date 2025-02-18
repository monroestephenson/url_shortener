package repository

import (
    "database/sql"
    "errors"
    "time"

    "url_shortener/internal/models"
)

var (
    ErrShortURLNotFound = errors.New("short URL not found")
)

type ShortURLRepository interface {
    Create(shortURL *models.ShortURL) error
    GetByShortCode(shortCode string) (*models.ShortURL, error)
    Update(shortURL *models.ShortURL) error
    DeleteByShortCode(shortCode string) error
    IncrementAccessCount(shortCode string) error
}

type shortURLRepository struct {
    db *sql.DB
}

func NewShortURLRepository(db *sql.DB) ShortURLRepository {
    return &shortURLRepository{db: db}
}

// Create inserts a new short URL record.
func (r *shortURLRepository) Create(shortURL *models.ShortURL) error {
    now := time.Now()
    shortURL.CreatedAt = now
    shortURL.UpdatedAt = now

    query := `
        INSERT INTO short_urls (short_code, original_url, access_count, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)
    `

    result, err := r.db.Exec(
        query,
        shortURL.ShortCode,
        shortURL.OriginalURL,
        shortURL.AccessCount,
        shortURL.CreatedAt,
        shortURL.UpdatedAt,
    )
    if err != nil {
        return err
    }

    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    shortURL.ID = int(id)
    return nil
}

// GetByShortCode retrieves a record by short_code.
func (r *shortURLRepository) GetByShortCode(shortCode string) (*models.ShortURL, error) {
    query := `
        SELECT id, short_code, original_url, access_count, created_at, updated_at
        FROM short_urls
        WHERE short_code = ?
    `

    row := r.db.QueryRow(query, shortCode)

    var su models.ShortURL
    err := row.Scan(
        &su.ID,
        &su.ShortCode,
        &su.OriginalURL,
        &su.AccessCount,
        &su.CreatedAt,
        &su.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, ErrShortURLNotFound
    } else if err != nil {
        return nil, err
    }

    return &su, nil
}

// Update updates the original_url (and updated_at) for a record.
func (r *shortURLRepository) Update(shortURL *models.ShortURL) error {
    shortURL.UpdatedAt = time.Now()

    query := `
        UPDATE short_urls
        SET original_url = ?, updated_at = ?
        WHERE short_code = ?
    `
    result, err := r.db.Exec(query, shortURL.OriginalURL, shortURL.UpdatedAt, shortURL.ShortCode)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return ErrShortURLNotFound
    }

    return nil
}

// DeleteByShortCode deletes a record by its short_code.
func (r *shortURLRepository) DeleteByShortCode(shortCode string) error {
    query := `
        DELETE FROM short_urls
        WHERE short_code = ?
    `
    result, err := r.db.Exec(query, shortCode)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return ErrShortURLNotFound
    }
    return nil
}

// IncrementAccessCount increments the access_count of a short URL whenever it's accessed.
func (r *shortURLRepository) IncrementAccessCount(shortCode string) error {
    query := `
        UPDATE short_urls
        SET access_count = access_count + 1
        WHERE short_code = ?
    `
    result, err := r.db.Exec(query, shortCode)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return ErrShortURLNotFound
    }
    return nil
}
