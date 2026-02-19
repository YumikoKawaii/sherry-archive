package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	"github.com/yumikokawaii/sherry-archive/pkg/password"
	"github.com/yumikokawaii/sherry-archive/pkg/token"
)

type AuthService struct {
	userRepo    repository.UserRepository
	tokenRepo   repository.RefreshTokenRepository
	tokenMgr    *token.Manager
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	tokenMgr *token.Manager,
) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, tokenMgr: tokenMgr}
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*model.User, *TokenPair, error) {
	// Check uniqueness
	if _, err := s.userRepo.GetByEmail(ctx, in.Email); !errors.Is(err, apperror.ErrNotFound) {
		if err == nil {
			return nil, nil, apperror.ErrConflict
		}
		return nil, nil, err
	}
	if _, err := s.userRepo.GetByUsername(ctx, in.Username); !errors.Is(err, apperror.ErrNotFound) {
		if err == nil {
			return nil, nil, apperror.ErrConflict
		}
		return nil, nil, err
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	u := &model.User{
		ID:           uuid.Must(uuid.NewV7()),
		Username:     in.Username,
		Email:        in.Email,
		PasswordHash: hash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, nil, err
	}

	pair, err := s.issueTokenPair(ctx, u.ID)
	if err != nil {
		return nil, nil, err
	}
	return u, pair, nil
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*model.User, *TokenPair, error) {
	u, err := s.userRepo.GetByEmail(ctx, in.Email)
	if errors.Is(err, apperror.ErrNotFound) {
		return nil, nil, apperror.ErrUnauthorized
	}
	if err != nil {
		return nil, nil, err
	}

	if !password.Verify(u.PasswordHash, in.Password) {
		return nil, nil, apperror.ErrUnauthorized
	}

	pair, err := s.issueTokenPair(ctx, u.ID)
	if err != nil {
		return nil, nil, err
	}
	return u, pair, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.tokenMgr.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	hash := hashToken(refreshToken)
	rt, err := s.tokenRepo.GetByHash(ctx, hash)
	if errors.Is(err, apperror.ErrNotFound) {
		return nil, apperror.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	if time.Now().After(rt.ExpiresAt) {
		_ = s.tokenRepo.DeleteByHash(ctx, hash)
		return nil, apperror.ErrTokenExpired
	}

	// Token rotation: delete old, issue new
	if err := s.tokenRepo.DeleteByHash(ctx, hash); err != nil {
		return nil, err
	}

	return s.issueTokenPair(ctx, claims.UserID)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	hash := hashToken(refreshToken)
	return s.tokenRepo.DeleteByHash(ctx, hash)
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	accessToken, err := s.tokenMgr.IssueAccessToken(userID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.tokenMgr.IssueRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	rt := &model.RefreshToken{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    userID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: now.Add(s.tokenMgr.RefreshExpiry()),
		CreatedAt: now,
	}
	if err := s.tokenRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}
