package plugin

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pclsguserver/internal/audit"
	"pclsguserver/internal/index"
	"pclsguserver/internal/metrics"
	"pclsguserver/internal/model"
	"pclsguserver/internal/pkg/checksum"
	"pclsguserver/internal/repo"
	"pclsguserver/internal/store"
)

var (
	ErrPluginNotFound   = errors.New("plugin not found")
	ErrInvalidStatus    = errors.New("invalid status")
	ErrIndexUnavailable = errors.New("index unavailable")
)

type Service struct {
	repo          *repo.PluginRepo
	store         store.BinaryStore
	metrics       *metrics.Service
	audit         *audit.Service
	indexEngine   *index.Engine
	indexBuilder  *index.Builder
	db            *gorm.DB
	degradePolicy string
}

func NewService(r *repo.PluginRepo, st store.BinaryStore, m *metrics.Service, a *audit.Service, eng *index.Engine, bld *index.Builder, db *gorm.DB, degrade string) *Service {
	return &Service{repo: r, store: st, metrics: m, audit: a, indexEngine: eng, indexBuilder: bld, db: db, degradePolicy: degrade}
}

type SubmitInput struct {
	Name        string
	Version     string
	Author      string
	Category    string
	Keywords    []string
	Description string
	DownloadUrl string
}

func (s *Service) Submit(ctx context.Context, in SubmitInput, reader io.Reader, submitterId, operator, requestID, sourceIP string) (*model.Plugin, error) {
	sum, size, err := checksum.ComputeSHA256(reader)
	if err != nil {
		return nil, err
	}
	p := &model.Plugin{
		ID:          uuid.NewString(),
		Name:        in.Name,
		Version:     in.Version,
		Author:      in.Author,
		Category:    in.Category,
		Keywords:    in.Keywords,
		Description: in.Description,
		DownloadUrl: in.DownloadUrl,
		Checksum:    sum,
		Status:      model.PluginStatusReviewing,
		Size:        size,
		SubmitterId: submitterId,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tx := s.db.Begin()
	if err := s.repo.Create(tx, p); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   operator,
		Action:     model.AuditActionPluginSubmitted,
		TargetType: model.AuditTargetPlugin,
		TargetID:   p.ID,
		Result:     model.AuditResultSuccess,
		Detail:     "submit " + in.Name + "@" + in.Version,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	}); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) SavePackage(ctx context.Context, id, version string, reader io.Reader) error {
	path := store.PluginPath(id, version)
	_, err := s.store.Save(ctx, path, reader)
	return err
}

func (s *Service) Offline(ctx context.Context, id, operator, requestID, sourceIP string) error {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return ErrPluginNotFound
	}
	if p.Status != model.PluginStatusPublished {
		return ErrInvalidStatus
	}
	p.Status = model.PluginStatusOffline
	p.UpdatedAt = time.Now()

	tx := s.db.Begin()
	if err := s.repo.Update(tx, p); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   operator,
		Action:     model.AuditActionOffline,
		TargetType: model.AuditTargetPlugin,
		TargetID:   id,
		Result:     model.AuditResultSuccess,
		Detail:     "offline " + p.Name,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	s.indexBuilder.NotifyPluginOffline(id)
	return nil
}

func (s *Service) Delete(ctx context.Context, id string, confirm bool, operator, requestID, sourceIP string) error {
	if !confirm {
		return errors.New("confirmation required")
	}
	p, err := s.repo.GetByID(id)
	if err != nil {
		return ErrPluginNotFound
	}
	tx := s.db.Begin()
	if err := s.repo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   operator,
		Action:     model.AuditActionDelete,
		TargetType: model.AuditTargetPlugin,
		TargetID:   id,
		Result:     model.AuditResultSuccess,
		Detail:     "delete " + p.Name,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	_ = s.store.Delete(ctx, store.PluginPath(id, p.Version))
	s.indexBuilder.NotifyPluginOffline(id)
	return nil
}

func (s *Service) Resubmit(ctx context.Context, id string, in SubmitInput, reader io.Reader, operator, requestID, sourceIP string) (*model.Plugin, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrPluginNotFound
	}
	if p.Status != model.PluginStatusRejected {
		return nil, ErrInvalidStatus
	}
	sum, size, err := checksum.ComputeSHA256(reader)
	if err != nil {
		return nil, err
	}
	p.Name = in.Name
	p.Version = in.Version
	p.Author = in.Author
	p.Category = in.Category
	p.Keywords = in.Keywords
	p.Description = in.Description
	p.DownloadUrl = in.DownloadUrl
	p.Checksum = sum
	p.Size = size
	p.Status = model.PluginStatusReviewing
	p.RejectReason = ""
	p.UpdatedAt = time.Now()

	tx := s.db.Begin()
	if err := s.repo.Update(tx, p); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   operator,
		Action:     model.AuditActionPluginSubmitted,
		TargetType: model.AuditTargetPlugin,
		TargetID:   id,
		Result:     model.AuditResultSuccess,
		Detail:     "resubmit " + in.Name,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	}); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListPublished(page, size int, category string) ([]model.Plugin, int64, error) {
	return s.repo.ListPublished(page, size, category)
}

type SearchResult struct {
	Items []model.Plugin `json:"items"`
	Total int64          `json:"total"`
}

func (s *Service) Search(keyword, category, name string, page, size int) (*SearchResult, error) {
	if s.indexEngine.IsReady() {
		r, ok := s.indexEngine.Search(keyword, category, name, page, size)
		if ok {
			items := make([]model.Plugin, 0, len(r.Items))
			for _, idx := range r.Items {
				p, err := s.repo.GetByID(idx.ID)
				if err == nil {
					items = append(items, *p)
				}
			}
			return &SearchResult{Items: items, Total: int64(r.Total)}, nil
		}
	}
	if s.degradePolicy == "db" {
		ps, total, err := s.repo.SearchFallback(keyword, category, name, page, size)
		if err != nil {
			return nil, err
		}
		return &SearchResult{Items: ps, Total: total}, nil
	}
	return nil, ErrIndexUnavailable
}

func (s *Service) GetPublished(id string) (*model.Plugin, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrPluginNotFound
	}
	if p.Status != model.PluginStatusPublished {
		return nil, ErrPluginNotFound
	}
	return p, nil
}

func (s *Service) GetForDownload(ctx context.Context, id, version, clientId string) (*model.Plugin, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, ErrPluginNotFound
	}
	if p.Status != model.PluginStatusPublished {
		return nil, ErrPluginNotFound
	}
	return p, nil
}

func (s *Service) ReadPackage(ctx context.Context, id, version string) (io.ReadCloser, error) {
	return s.store.Read(ctx, store.PluginPath(id, version))
}

func (s *Service) PackageSize(ctx context.Context, id, version string) (int64, error) {
	return s.store.Stat(ctx, store.PluginPath(id, version))
}

func (s *Service) IncrementDownloadCount(id, clientId string) {
	_ = s.repo.IncrementDownloadCount(id)
	s.metrics.RecordEvent(model.MetricsDimensionPluginDownload, id, clientId)
}

func (s *Service) ListReviewing(page, size int) ([]model.Plugin, int64, error) {
	return s.repo.ListReviewing(page, size)
}

func (s *Service) GetByID(id string) (*model.Plugin, error) {
	return s.repo.GetByID(id)
}
