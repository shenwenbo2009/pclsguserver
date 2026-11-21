package upload

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
)

type UploadTokenService struct {
	tokenRepo *repo.UploadTokenRepo
}

func NewUploadTokenService(tokenRepo *repo.UploadTokenRepo) *UploadTokenService {
	return &UploadTokenService{tokenRepo: tokenRepo}
}

func (s *UploadTokenService) Create(name, uploaderIdentity string, permissions []string) (string, *model.UploadToken, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", nil, fmt.Errorf("生成 Token 失败: %w", err)
	}
	token := "utk_" + hex.EncodeToString(tokenBytes)

	saltBytes := make([]byte, 32)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", nil, fmt.Errorf("生成盐值失败: %w", err)
	}
	salt := hex.EncodeToString(saltBytes)

	hash := sha256.Sum256([]byte(salt + token))
	tokenHash := hex.EncodeToString(hash[:])

	record := &model.UploadToken{
		ID:               generateTokenID(),
		TokenHash:        tokenHash,
		Salt:             salt,
		Name:             name,
		UploaderIdentity: uploaderIdentity,
		Permissions:      permissions,
		Status:           model.UploadTokenStatusActive,
		CreatedAt:        time.Now(),
	}

	if err := s.tokenRepo.Create(nil, record); err != nil {
		return "", nil, fmt.Errorf("保存 Token 失败: %w", err)
	}

	return token, record, nil
}

func (s *UploadTokenService) Validate(token string) (*model.UploadToken, error) {
	if token == "" {
		return nil, fmt.Errorf("Token 为空")
	}

	tokens, _, err := s.tokenRepo.List(1, 1000)
	if err != nil {
		return nil, fmt.Errorf("查询 Token 失败: %w", err)
	}

	for _, t := range tokens {
		hash := sha256.Sum256([]byte(t.Salt + token))
		if hex.EncodeToString(hash[:]) == t.TokenHash {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("Token 无效")
}

func (s *UploadTokenService) Disable(id string) error {
	token, err := s.tokenRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("Token 不存在: %w", err)
	}
	token.Status = model.UploadTokenStatusDisabled
	now := time.Now()
	token.DisabledAt = &now
	return s.tokenRepo.Update(nil, token)
}

func (s *UploadTokenService) List(page, size int) ([]model.UploadToken, int64, error) {
	return s.tokenRepo.List(page, size)
}

func generateTokenID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "utk-" + hex.EncodeToString(b)
}
