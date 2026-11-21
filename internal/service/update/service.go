package update

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pclsguserver/internal/audit"
	"pclsguserver/internal/metrics"
	"pclsguserver/internal/model"
	"pclsguserver/internal/pkg/checksum"
	"pclsguserver/internal/repo"
	"pclsguserver/internal/store"
)

var (
	ErrSwNotFound      = errors.New("software not found")
	ErrVersionNotFound = errors.New("version not found")
	ErrInvalidStatus   = errors.New("invalid status")
)

type Service struct {
	repo    *repo.SoftwareVersionRepo
	store   store.BinaryStore
	metrics *metrics.Service
	audit   *audit.Service
	db      *gorm.DB
}

func NewService(r *repo.SoftwareVersionRepo, st store.BinaryStore, m *metrics.Service, a *audit.Service, db *gorm.DB) *Service {
	return &Service{repo: r, store: st, metrics: m, audit: a, db: db}
}

type PublishInput struct {
	SwCode      string
	Version     string
	ReleaseNote string
	SwUrl       string
}

func (s *Service) Publish(ctx context.Context, in PublishInput, reader io.Reader, operator, requestID, sourceIP string) (*model.SoftwareVersion, error) {
	sum, size, err := checksum.ComputeSHA256(reader)
	if err != nil {
		return nil, err
	}
	path := store.SwPath(in.SwCode, in.Version)
	exists, _ := s.store.Exists(ctx, path)
	if exists {
		_ = s.store.Delete(ctx, path)
	}

	v := &model.SoftwareVersion{
		ID:          uuid.NewString(),
		SwCode:      in.SwCode,
		Version:     in.Version,
		SwUrl:       in.SwUrl,
		Checksum:    sum,
		ReleaseNote: in.ReleaseNote,
		Status:      model.SoftwareStatusPublished,
		Size:        size,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	now := time.Now()
	v.PublishedAt = &now

	tx := s.db.Begin()
	if err := s.repo.Create(tx, v); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   operator,
		Action:     model.AuditActionPublish,
		TargetType: model.AuditTargetSoftware,
		TargetID:   v.ID,
		Result:     model.AuditResultSuccess,
		Detail:     "publish " + in.SwCode + "@" + in.Version,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	}); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return v, nil
}

func (s *Service) SavePackage(ctx context.Context, swCode, version string, reader io.Reader) error {
	path := store.SwPath(swCode, version)
	_, err := s.store.Save(ctx, path, reader)
	return err
}

func (s *Service) Offline(ctx context.Context, id, operator, requestID, sourceIP string) error {
	v, err := s.repo.GetByID(id)
	if err != nil {
		return ErrSwNotFound
	}
	if v.Status != model.SoftwareStatusPublished {
		return ErrInvalidStatus
	}
	v.Status = model.SoftwareStatusOffline
	v.UpdatedAt = time.Now()

	tx := s.db.Begin()
	if err := s.repo.Update(tx, v); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   operator,
		Action:     model.AuditActionOffline,
		TargetType: model.AuditTargetSoftware,
		TargetID:   id,
		Result:     model.AuditResultSuccess,
		Detail:     "offline " + v.SwCode + "@" + v.Version,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *Service) Delete(ctx context.Context, id string, confirm bool, operator, requestID, sourceIP string) error {
	if !confirm {
		return errors.New("confirmation required")
	}
	v, err := s.repo.GetByID(id)
	if err != nil {
		return ErrSwNotFound
	}
	tx := s.db.Begin()
	if err := s.repo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   operator,
		Action:     model.AuditActionDelete,
		TargetType: model.AuditTargetSoftware,
		TargetID:   id,
		Result:     model.AuditResultSuccess,
		Detail:     "delete " + v.SwCode + "@" + v.Version,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	_ = s.store.Delete(ctx, store.SwPath(v.SwCode, v.Version))
	return nil
}

type CheckResult struct {
	SwCode         string `json:"swCode"`
	LatestVersion  string `json:"latestVersion"`
	CurrentVersion string `json:"currentVersion"`
	SwUrl          string `json:"swUrl"`
	Checksum       string `json:"checksum"`
	ReleaseNote    string `json:"releaseNote"`
	UpToDate       bool   `json:"upToDate"`
}

func (s *Service) Check(ctx context.Context, swCode, currentVersion, clientId string) (*CheckResult, error) {
	v, err := s.repo.FindLatestPublished(swCode)
	if err != nil {
		return nil, ErrSwNotFound
	}
	s.metrics.RecordEvent(model.MetricsDimensionSoftwareActivity, swCode+":"+v.Version, clientId)
	return &CheckResult{
		SwCode:         swCode,
		LatestVersion:  v.Version,
		CurrentVersion: currentVersion,
		SwUrl:          v.SwUrl,
		Checksum:       v.Checksum,
		ReleaseNote:    v.ReleaseNote,
		UpToDate:       currentVersion == v.Version,
	}, nil
}

func (s *Service) GetForDownload(ctx context.Context, swCode, version, clientId string) (*model.SoftwareVersion, error) {
	v, err := s.repo.GetByVersion(swCode, version)
	if err != nil {
		return nil, ErrVersionNotFound
	}
	if v.Status != model.SoftwareStatusPublished {
		return nil, ErrVersionNotFound
	}
	return v, nil
}

func (s *Service) ReadPackage(ctx context.Context, swCode, version string) (io.ReadCloser, error) {
	return s.store.Read(ctx, store.SwPath(swCode, version))
}

func (s *Service) ReadPackageRange(ctx context.Context, swCode, version string, offset, length int64) (io.ReadCloser, error) {
	return s.store.ReadRange(ctx, store.SwPath(swCode, version), offset, length)
}

func (s *Service) PackageSize(ctx context.Context, swCode, version string) (int64, error) {
	return s.store.Stat(ctx, store.SwPath(swCode, version))
}

func (s *Service) IncrementDownloadCount(swCode, version, clientId string) {
	_ = s.repo.IncrementDownloadCount(swCode, version)
	s.metrics.RecordEvent(model.MetricsDimensionSoftwareDownload, swCode+":"+version, clientId)
}

func (s *Service) ListBySwCode(swCode string) ([]model.SoftwareVersion, error) {
	return s.repo.ListBySwCode(swCode)
}
