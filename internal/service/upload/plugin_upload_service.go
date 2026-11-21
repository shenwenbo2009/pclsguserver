package upload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pclsguserver/internal/audit"
	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
	"pclsguserver/internal/store"
)

var semverRegex = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`)

type PluginUploadInput struct {
	Name        string
	Version     string
	Author      string
	Category    string
	Description string
	Keywords    []string
	Filename    string
	FileSize    int64
}

type PluginUploadResult struct {
	PluginId   string `json:"pluginId"`
	Status     string `json:"status"`
	Checksum   string `json:"checksum"`
	Size       int64  `json:"size"`
	UploadedAt string `json:"uploadedAt"`
}

type PluginUploadService struct {
	validator  *FileValidator
	store      store.BinaryStore
	pluginRepo *repo.PluginRepo
	audit      *audit.Service
	db         *gorm.DB
	maxSize    int64
}

func NewPluginUploadService(validator *FileValidator, st store.BinaryStore, pluginRepo *repo.PluginRepo, auditSvc *audit.Service, db *gorm.DB, maxSize int64) *PluginUploadService {
	return &PluginUploadService{
		validator:  validator,
		store:      st,
		pluginRepo: pluginRepo,
		audit:      auditSvc,
		db:         db,
		maxSize:    maxSize,
	}
}

func (s *PluginUploadService) Upload(ctx context.Context, in PluginUploadInput, reader io.Reader, uploaderIdentity, requestID, sourceIP string) (*PluginUploadResult, error) {
	if len(in.Name) < 1 || len(in.Name) > 100 {
		return nil, fmt.Errorf("%w: name 长度需在 1-100 字符之间", ErrInvalidParam)
	}
	if !semverRegex.MatchString(in.Version) {
		return nil, fmt.Errorf("%w: version 需符合 semver 格式", ErrInvalidParam)
	}
	if len(in.Author) < 1 || len(in.Author) > 100 {
		return nil, fmt.Errorf("%w: author 长度需在 1-100 字符之间", ErrInvalidParam)
	}
	if in.Category == "" {
		return nil, fmt.Errorf("%w: category 不能为空", ErrInvalidParam)
	}

	if err := s.validator.ValidatePackageType(in.Filename); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFileType, err)
	}

	if err := s.validator.ValidateSize(in.FileSize, s.maxSize); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFileTooLarge, err)
	}

	var count int64
	s.db.Model(&model.Plugin{}).Where("name = ? AND version = ? AND status IN ?", in.Name, in.Version, []model.PluginStatus{model.PluginStatusReviewing, model.PluginStatusPublished}).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("%w: 插件 %s@%s 已存在", ErrPluginAlreadyExists, in.Name, in.Version)
	}

	pluginId := uuid.NewString()
	path := store.PluginPath(pluginId, in.Version)

	h := sha256.New()
	tee := io.TeeReader(reader, h)
	size, err := s.store.Save(ctx, path, tee)
	if err != nil {
		_ = s.store.Delete(ctx, path)
		return nil, fmt.Errorf("存储失败: %w", err)
	}
	checksum := hex.EncodeToString(h.Sum(nil))

	p := &model.Plugin{
		ID:          pluginId,
		Name:        in.Name,
		Version:     in.Version,
		Author:      in.Author,
		Category:    in.Category,
		Keywords:    in.Keywords,
		Description: in.Description,
		Checksum:    checksum,
		Status:      model.PluginStatusReviewing,
		Size:        size,
		SubmitterId: uploaderIdentity,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tx := s.db.Begin()
	if err := s.pluginRepo.Create(tx, p); err != nil {
		tx.Rollback()
		_ = s.store.Delete(ctx, path)
		return nil, fmt.Errorf("创建记录失败: %w", err)
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   uploaderIdentity,
		Action:     model.AuditActionPluginUploaded,
		TargetType: model.AuditTargetPlugin,
		TargetID:   pluginId,
		Result:     model.AuditResultSuccess,
		Detail:     "upload " + in.Name + "@" + in.Version,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	}); err != nil {
		tx.Rollback()
		_ = s.store.Delete(ctx, path)
		return nil, fmt.Errorf("审计日志失败: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		_ = s.store.Delete(ctx, path)
		return nil, fmt.Errorf("事务提交失败: %w", err)
	}

	return &PluginUploadResult{
		PluginId:   pluginId,
		Status:     string(model.PluginStatusReviewing),
		Checksum:   checksum,
		Size:       size,
		UploadedAt: time.Now().Format(time.RFC3339),
	}, nil
}
