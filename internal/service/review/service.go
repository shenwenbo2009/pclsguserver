package review

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pclsguserver/internal/audit"
	"pclsguserver/internal/index"
	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
)

var (
	ErrPluginNotFound = errors.New("plugin not found")
	ErrNotReviewable  = errors.New("plugin not reviewable")
)

type Service struct {
	pluginRepo   *repo.PluginRepo
	reviewRepo   *repo.ReviewRecordRepo
	audit        *audit.Service
	indexBuilder *index.Builder
	db           *gorm.DB
}

func NewService(pr *repo.PluginRepo, rr *repo.ReviewRecordRepo, a *audit.Service, bld *index.Builder, db *gorm.DB) *Service {
	return &Service{pluginRepo: pr, reviewRepo: rr, audit: a, indexBuilder: bld, db: db}
}

type ReviewInput struct {
	PluginID string
	Pass     bool
	Reason   string
	Comment  string
}

func (s *Service) Review(in ReviewInput, reviewer, requestID, sourceIP string) (*model.Plugin, error) {
	p, err := s.pluginRepo.GetByID(in.PluginID)
	if err != nil {
		return nil, ErrPluginNotFound
	}
	if p.Status != model.PluginStatusReviewing {
		return nil, ErrNotReviewable
	}

	var conclusion model.ReviewConclusion
	var action model.AuditAction
	if in.Pass {
		conclusion = model.ReviewConclusionPass
		action = model.AuditActionApprove
		p.Status = model.PluginStatusPublished
		now := time.Now()
		p.PublishedAt = &now
	} else {
		conclusion = model.ReviewConclusionReject
		action = model.AuditActionReject
		p.Status = model.PluginStatusRejected
		p.RejectReason = in.Reason
	}
	p.UpdatedAt = time.Now()

	rr := &model.ReviewRecord{
		ID:         uuid.NewString(),
		PluginID:   in.PluginID,
		Reviewer:   reviewer,
		Conclusion: conclusion,
		Reason:     in.Reason,
		Comment:    in.Comment,
		ReviewedAt: time.Now(),
	}

	tx := s.db.Begin()
	if err := s.pluginRepo.Update(tx, p); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.reviewRepo.Create(tx, rr); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator:   reviewer,
		Action:     action,
		TargetType: model.AuditTargetPlugin,
		TargetID:   in.PluginID,
		Result:     model.AuditResultSuccess,
		Detail:     "review " + p.Name,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	}); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if in.Pass {
		s.indexBuilder.NotifyPluginPublished(in.PluginID)
	}
	return p, nil
}

func (s *Service) ListReviewing(page, size int) ([]model.Plugin, int64, error) {
	return s.pluginRepo.ListReviewing(page, size)
}

func (s *Service) ListByPlugin(pluginId string) ([]model.ReviewRecord, error) {
	return s.reviewRepo.ListByPlugin(pluginId)
}
