package postgres

import (
	"auth-service/internal/domain/models"
	"auth-service/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"math/rand/v2"
	"strings"
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

	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id`

	username, err := s.generateUniqueUsername()
	if err != nil {
		return 0, fmt.Errorf("internal error, try later")
	}
	var id int64
	err = s.db.QueryRowContext(ctx, query, username, email, passHash).Scan(&id)
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

	query := `SELECT id, username, email, password, role_id FROM users WHERE email = $1`

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

func (s *Storage) GetUserByID(ctx context.Context, userID int64) (models.User, error) {
	const op = "storage.postgres.GetUserByID"

	query := `SELECT id, username, email, password, role_id FROM users WHERE id = $1`

	var user models.User
	err := s.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) GetUsers(ctx context.Context, roleID *int64, nameStartsWith *string) ([]models.User, error) {
	const op = "storage.postgres.GetUsers"

	var args []interface{}
	var conditions []string
	argIndex := 1

	query := `SELECT id, username, email, role_id FROM users WHERE 1=1`

	if roleID != nil {
		conditions = append(conditions, fmt.Sprintf(" AND role_id = $%d", argIndex))
		args = append(args, *roleID)
		argIndex++
	}

	if nameStartsWith != nil {
		conditions = append(conditions, fmt.Sprintf(" AND username LIKE $%d", argIndex))
		args = append(args, *nameStartsWith+"%")
		argIndex++
	}

	fullQuery := query + strings.Join(conditions, "")

	var users []models.User
	err := s.db.SelectContext(ctx, &users, fullQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func (s *Storage) DeleteUserByUsername(ctx context.Context, username string) error {
	const op = "storage.postgres.DeleteUserByUsername"

	query := `DELETE FROM users WHERE username = $1`

	res, err := s.db.ExecContext(ctx, query, username)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if affected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}

	return nil
}

func (s *Storage) generateUniqueUsername() (string, error) {
	for {
		adjectives := []string{
			"ephemeral", "quixotic", "luminous", "serendipitous", "nebulous",
			"effervescent", "obstreperous", "surreptitious", "perspicuous", "phantasmagorical",
		}
		nouns := []string{
			"nebula", "quasar", "wisp", "aether", "rune",
			"spectre", "chasm", "vortex", "shimmer", "enigma",
		}

		adj := adjectives[rand.IntN(len(adjectives))]
		noun := nouns[rand.IntN(len(nouns))]
		number := rand.IntN(1000)
		newUsername := fmt.Sprintf("%s-%s-%d", adj, noun, number)

		var exists bool
		err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", newUsername).Scan(&exists)
		if err != nil {
			return "", err
		}

		if !exists {
			return newUsername, nil
		}
	}
}
