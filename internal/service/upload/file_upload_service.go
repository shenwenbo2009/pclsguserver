package upload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pclsguserver/internal/audit"
	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
	"pclsguserver/internal/store"
)

type FileUploadInput struct {
	Purpose  string
	Filename string
	FileSize int64
}

type FileUploadResult struct {
	FileId     string `json:"fileId"`
	URL        string `json:"url"`
	Checksum   string `json:"checksum"`
	Size       int64  `json:"size"`
	Purpose    string `json:"purpose"`
	UploadedAt string `json:"uploadedAt"`
}

type FileUploadService struct {
	validator      *FileValidator
	store          store.BinaryStore
	fileRecordRepo *repo.FileRecordRepo
	audit          *audit.Service
	db             *gorm.DB
	maxSize        int64
}

func NewFileUploadService(validator *FileValidator, st store.BinaryStore, fileRecordRepo *repo.FileRecordRepo, auditSvc *audit.Service, db *gorm.DB, maxSize int64) *FileUploadService {
	return &FileUploadService{
		validator:      validator,
		store:          st,
		fileRecordRepo: fileRecordRepo,
		audit:          auditSvc,
		db:             db,
		maxSize:        maxSize,
	}
}

func (s *FileUploadService) Upload(ctx context.Context, in FileUploadInput, reader io.Reader, uploaderIdentity, requestID, sourceIP string) (*FileUploadResult, error) {
	validPurposes := map[string]bool{
		string(model.FilePurposeAvatar):     true,
		string(model.FilePurposeAttachment): true,
		string(model.FilePurposeScreenshot): true,
		string(model.FilePurposeDocument):   true,
		string(model.FilePurposeOther):      true,
	}
	if !validPurposes[in.Purpose] {
		return nil, fmt.Errorf("%w: purpose 非法", ErrInvalidPurpose)
	}

	if err := s.validator.ValidateFileType(in.Filename); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFileType, err)
	}

	if err := s.validator.ValidateSize(in.FileSize, s.maxSize); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFileTooLarge, err)
	}

	cleanFilename, err := s.validator.SanitizeFilename(in.Filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFilename, err)
	}

	randomUUID := uuid.NewString()
	path := store.FilePath(in.Purpose, randomUUID, cleanFilename)

	h := sha256.New()
	tee := io.TeeReader(reader, h)
	size, err := s.store.Save(ctx, path, tee)
	if err != nil {
		_ = s.store.Delete(ctx, path)
		return nil, fmt.Errorf("存储失败: %w", err)
	}
	checksum := hex.EncodeToString(h.Sum(nil))

	fileId := uuid.NewString()
	url := "/files/" + in.Purpose + "/" + randomUUID + "/" + cleanFilename

	record := &model.FileRecord{
		ID:               fileId,
		Purpose:          model.FilePurpose(in.Purpose),
		OriginalFilename: cleanFilename,
		StoredPath:       path,
		URL:              url,
		Checksum:         checksum,
		Size:             size,
		UploaderIdentity: uploaderIdentity,
		CreatedAt:        time.Now(),
	}

	tx := s.db.Begin()
	if err := s.fileRecordRepo.Create(tx, record); err != nil {
		tx.Rollback()
		_ = s.store.Delete(ctx, path)
		return nil, fmt.Errorf("创建记录失败: %w", err)
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   uploaderIdentity,
		Action:     model.AuditActionFileUploaded,
		TargetType: model.AuditTargetFile,
		TargetID:   fileId,
		Result:     model.AuditResultSuccess,
		Detail:     "upload " + in.Purpose + "/" + cleanFilename,
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

	return &FileUploadResult{
		FileId:     fileId,
		URL:        url,
		Checksum:   checksum,
		Size:       size,
		Purpose:    in.Purpose,
		UploadedAt: time.Now().Format(time.RFC3339),
	}, nil
}
