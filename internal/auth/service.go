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
	Role      string `json:"role"`
}

func (s *Service) Signup(ctx context.Context, input SignupInput) (*AuthResult, error) {
	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("auth: hash password: %w", err)
	}

	var userID string
	err = s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'viewer') RETURNING id`,
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
		Role:      "viewer",
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
			Role:      "viewer",
		},
	}, nil
}

type LoginInput struct {
	Email     string
	Password  string
	ProjectID string
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	var userID, passwordHash, role string
	err := s.pool.QueryRow(ctx,
		`SELECT id, password_hash, role FROM users WHERE email = $1`,
		input.Email,
	).Scan(&userID, &passwordHash, &role)
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
		Role:      role,
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
			Role:      role,
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
