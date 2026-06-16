package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-piazzano/cosmoria/internal/core"
)

type Service struct {
	pool *pgxpool.Pool
	cfg  *core.Config
}

func NewService(pool *pgxpool.Pool, cfg *core.Config) *Service {
	return &Service{pool: pool, cfg: cfg}
}

type SignupInput struct {
	Email     string
	Password  string
	ProjectID string
}

type AuthResult struct {
	Token string
	User  UserDTO
}

type UserDTO struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	ProjectID string `json:"project_id"`
}

func (s *Service) Signup(ctx context.Context, input SignupInput) (*AuthResult, error) {
	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("auth: hash password: %w", err)
	}

	var userID string
	err = s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		input.Email, string(hash),
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("auth: create user: %w", err)
	}

	expiry, err := s.resolveExpiry(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	claims := Claims{
		UserID:    userID,
		ProjectID: input.ProjectID,
	}

	token, err := GenerateToken(claims, s.cfg.JWTSecret, expiry)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token: token,
		User: UserDTO{
			ID:        userID,
			Email:     input.Email,
			ProjectID: input.ProjectID,
		},
	}, nil
}

type LoginInput struct {
	Email     string
	Password  string
	ProjectID string
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	var userID, passwordHash string
	err := s.pool.QueryRow(ctx,
		`SELECT id, password_hash FROM users WHERE email = $1`,
		input.Email,
	).Scan(&userID, &passwordHash)
	if err != nil {
		return nil, fmt.Errorf("auth: invalid credentials")
	}

	if !CheckPassword(input.Password, []byte(passwordHash)) {
		return nil, fmt.Errorf("auth: invalid credentials")
	}

	expiry, err := s.resolveExpiry(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}

	claims := Claims{
		UserID:    userID,
		ProjectID: input.ProjectID,
	}

	token, err := GenerateToken(claims, s.cfg.JWTSecret, expiry)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token: token,
		User: UserDTO{
			ID:        userID,
			Email:     input.Email,
			ProjectID: input.ProjectID,
		},
	}, nil
}

func (s *Service) resolveExpiry(ctx context.Context, projectID string) (int64, error) {
	var projectExpiry *int64
	err := s.pool.QueryRow(ctx,
		`SELECT jwt_expiry FROM projects WHERE id = $1`,
		projectID,
	).Scan(&projectExpiry)
	if err != nil {
		return 0, fmt.Errorf("auth: resolve project: %w", err)
	}

	if projectExpiry != nil && *projectExpiry > 0 {
		return *projectExpiry, nil
	}

	return s.cfg.JWTExpiry, nil
}

func (s *Service) GetByID(ctx context.Context, userID string) (*UserDTO, error) {
	var u UserDTO
	err := s.pool.QueryRow(ctx,
		`SELECT id, email FROM users WHERE id = $1`,
		userID,
	).Scan(&u.ID, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("auth: get user: %w", err)
	}
	return &u, nil
}

type UpdateEmailInput struct {
	UserID string
	Email  string
}

func (s *Service) UpdateEmail(ctx context.Context, input UpdateEmailInput) (*UserDTO, error) {
	var u UserDTO
	err := s.pool.QueryRow(ctx,
		`UPDATE users SET email = $1 WHERE id = $2
		 RETURNING id, email`,
		input.Email, input.UserID,
	).Scan(&u.ID, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("auth: update email: %w", err)
	}
	return &u, nil
}

type UpdatePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

func (s *Service) UpdatePassword(ctx context.Context, input UpdatePasswordInput) error {
	var passwordHash string
	err := s.pool.QueryRow(ctx,
		`SELECT password_hash FROM users WHERE id = $1`,
		input.UserID,
	).Scan(&passwordHash)
	if err != nil {
		return fmt.Errorf("auth: user not found")
	}

	if !CheckPassword(input.CurrentPassword, []byte(passwordHash)) {
		return fmt.Errorf("auth: invalid current password")
	}

	newHash, err := HashPassword(input.NewPassword)
	if err != nil {
		return fmt.Errorf("auth: hash password: %w", err)
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`,
		string(newHash), input.UserID,
	)
	if err != nil {
		return fmt.Errorf("auth: update password: %w", err)
	}
	return nil
}
