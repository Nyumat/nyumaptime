package auth

import (
	"context"
	"database/sql"
	"errors"

	"encore.dev/storage/sqldb"
	"golang.org/x/crypto/bcrypt"
)

// encore:service
type AuthService struct{}

type AuthorizedUser struct {
	ID       int64
	Username string
	Password string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func initService() (*AuthService, error) {
	return &AuthService{}, nil
}

func (s *AuthService) AuthorizeUser(ctx context.Context, username, password string) (bool, error) {
	var user AuthorizedUser
	err := sqldb.QueryRow(ctx, `
        SELECT id, username, password
        FROM authorized_users
        WHERE username = $1
    `, username).Scan(&user.ID, &user.Username, &user.Password)

	if err == sql.ErrNoRows {
		return false, errors.New("invalid username or password")
	} else if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false, errors.New("invalid username or password")
	}

	return true, nil
}

// encore:api public method=POST path=/register
func (s *AuthService) RegisterUser(ctx context.Context, req *RegisterRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = sqldb.Exec(ctx, `
        INSERT INTO authorized_users (username, password)
        VALUES ($1, $2)
    `, req.Username, string(hashedPassword))
	if err != nil {
		return err
	}

	return nil
}

// encore:api public method=POST path=/login
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	success, err := s.AuthorizeUser(ctx, req.Username, req.Password)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &LoginResponse{
		Success: success,
		Message: "Login successful",
	}, nil
}
