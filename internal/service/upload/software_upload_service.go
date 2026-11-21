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

var swCodeRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type SoftwareUploadInput struct {
	SwCode      string
	Version     string
	ReleaseNote string
	SwUrl       string
	Filename    string
	FileSize    int64
}

type SoftwareUploadResult struct {
	VersionId  string `json:"versionId"`
	Status     string `json:"status"`
	Checksum   string `json:"checksum"`
	Size       int64  `json:"size"`
	UploadedAt string `json:"uploadedAt"`
}

type SoftwareUploadService struct {
	validator *FileValidator
	store     store.BinaryStore
	swRepo    *repo.SoftwareVersionRepo
	audit     *audit.Service
	db        *gorm.DB
	maxSize   int64
}

func NewSoftwareUploadService(validator *FileValidator, st store.BinaryStore, swRepo *repo.SoftwareVersionRepo, auditSvc *audit.Service, db *gorm.DB, maxSize int64) *SoftwareUploadService {
	return &SoftwareUploadService{
		validator: validator,
		store:     st,
		swRepo:    swRepo,
		audit:     auditSvc,
		db:        db,
		maxSize:   maxSize,
	}
}

func (s *SoftwareUploadService) Upload(ctx context.Context, in SoftwareUploadInput, reader io.Reader, uploaderIdentity, requestID, sourceIP string) (*SoftwareUploadResult, error) {
	if in.SwCode == "" || !swCodeRegex.MatchString(in.SwCode) {
		return nil, fmt.Errorf("%w: swCode 非空且仅含字母数字下划线", ErrInvalidParam)
	}
	if !semverRegex.MatchString(in.Version) {
		return nil, fmt.Errorf("%w: version 需符合 semver 格式", ErrInvalidParam)
	}

	if err := s.validator.ValidatePackageType(in.Filename); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFileType, err)
	}

	if err := s.validator.ValidateSize(in.FileSize, s.maxSize); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFileTooLarge, err)
	}

	var count int64
	s.db.Model(&model.SoftwareVersion{}).Where("swCode = ? AND version = ? AND status IN ?", in.SwCode, in.Version, []model.SoftwareStatus{model.SoftwareStatusDraft, model.SoftwareStatusPublished}).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("%w: 版本 %s@%s 已存在", ErrVersionAlreadyExists, in.SwCode, in.Version)
	}

	versionId := uuid.NewString()
	path := store.SwPath(in.SwCode, in.Version)

	h := sha256.New()
	tee := io.TeeReader(reader, h)
	size, err := s.store.Save(ctx, path, tee)
	if err != nil {
		_ = s.store.Delete(ctx, path)
		return nil, fmt.Errorf("存储失败: %w", err)
	}
	checksum := hex.EncodeToString(h.Sum(nil))

	v := &model.SoftwareVersion{
		ID:          versionId,
		SwCode:      in.SwCode,
		Version:     in.Version,
		SwUrl:       in.SwUrl,
		Checksum:    checksum,
		ReleaseNote: in.ReleaseNote,
		Status:      model.SoftwareStatusDraft,
		Size:        size,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tx := s.db.Begin()
	if err := s.swRepo.Create(tx, v); err != nil {
		tx.Rollback()
		_ = s.store.Delete(ctx, path)
		return nil, fmt.Errorf("创建记录失败: %w", err)
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   uploaderIdentity,
		Action:     model.AuditActionSoftwareUploaded,
		TargetType: model.AuditTargetSoftware,
		TargetID:   versionId,
		Result:     model.AuditResultSuccess,
		Detail:     "upload " + in.SwCode + "@" + in.Version,
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

	return &SoftwareUploadResult{
		VersionId:  versionId,
		Status:     string(model.SoftwareStatusDraft),
		Checksum:   checksum,
		Size:       size,
		UploadedAt: time.Now().Format(time.RFC3339),
	}, nil
}
