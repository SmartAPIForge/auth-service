package postgres

import (
	"auth-service/internal/domain/models"
	"auth-service/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB
}

func NewStorage(connString string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.postgres.SaveUser"

	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id`

	var id int64
	err := s.db.QueryRowContext(ctx, query, email, passHash).Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUser(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.GetUser"

	query := `SELECT id, email, password, role_id FROM users WHERE email = $1`

	var user models.User
	err := s.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
