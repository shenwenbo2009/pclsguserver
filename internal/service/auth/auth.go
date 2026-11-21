package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pclsguserver/internal/config"
	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
)

type Claims struct {
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type TokenStore struct {
	blacklist map[string]time.Time
}

func NewTokenStore() *TokenStore {
	return &TokenStore{blacklist: make(map[string]time.Time)}
}

func (s *TokenStore) Revoke(token string, exp time.Time) {
	s.blacklist[token] = exp
}

func (s *TokenStore) IsRevoked(token string) bool {
	exp, ok := s.blacklist[token]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		delete(s.blacklist, token)
		return false
	}
	return true
}

func (s *TokenStore) Cleanup() {
	now := time.Now()
	for k, exp := range s.blacklist {
		if now.After(exp) {
			delete(s.blacklist, k)
		}
	}
}

type AuthService struct {
	userRepo  *repo.AdminUserRepo
	roleRepo  *repo.RoleRepo
	auditRepo *repo.AuditRepo
	cfg       *config.Config
	tokens    *TokenStore
}

func NewAuthService(userRepo *repo.AdminUserRepo, roleRepo *repo.RoleRepo, auditRepo *repo.AuditRepo, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		auditRepo: auditRepo,
		cfg:       cfg,
		tokens:    NewTokenStore(),
	}
}

func (s *AuthService) TokenStore() *TokenStore { return s.tokens }

func (s *AuthService) signToken(user *model.AdminUser) (*TokenPair, error) {
	now := time.Now()
	roleIDs := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roleIDs = append(roleIDs, r.ID)
	}
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    roleIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWT.AccessTtl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := RefreshClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWT.RefreshTtl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		Token:        tokenStr,
		RefreshToken: refreshStr,
		ExpiresAt:    now.Add(s.cfg.JWT.AccessTtl),
	}, nil
}

func (s *AuthService) ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	if s.tokens.IsRevoked(tokenStr) {
		return nil, errors.New("token revoked")
	}
	return claims, nil
}

func (s *AuthService) ParseRefreshToken(tokenStr string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	return claims, nil
}

type LoginResult struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDisabled    = errors.New("account disabled")
	ErrAccountLocked      = errors.New("account locked")
)

func (s *AuthService) Login(username, password, clientIP string) (*LoginResult, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.Status == model.AccountStatusDisabled {
		return nil, ErrAccountDisabled
	}
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, ErrAccountLocked
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		user.FailCount++
		if user.FailCount >= s.cfg.RateLimit.LoginThreshold {
			lockedUntil := time.Now().Add(s.cfg.RateLimit.LoginWindow)
			user.LockedUntil = &lockedUntil
			user.Status = model.AccountStatusLocked
		}
		_ = s.userRepo.Update(nil, user)
		return nil, ErrInvalidCredentials
	}
	now := time.Now()
	user.LastLoginAt = &now
	user.FailCount = 0
	user.LockedUntil = nil
	if user.Status == model.AccountStatusLocked {
		user.Status = model.AccountStatusActive
	}
	if err := s.userRepo.Update(nil, user); err != nil {
		return nil, err
	}
	pair, err := s.signToken(user)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Token:        pair.Token,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt,
	}, nil
}

func (s *AuthService) Refresh(refreshToken string) (*LoginResult, error) {
	claims, err := s.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.Status != model.AccountStatusActive {
		return nil, ErrAccountDisabled
	}
	pair, err := s.signToken(user)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Token:        pair.Token,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt,
	}, nil
}

func (s *AuthService) Logout(token string) {
	claims, err := s.ParseToken(token)
	if err != nil {
		return
	}
	s.tokens.Revoke(token, claims.ExpiresAt.Time)
}
