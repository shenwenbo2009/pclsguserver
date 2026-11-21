package indexsvc

import (
	"errors"
	"time"

	"pclsguserver/internal/audit"
	"pclsguserver/internal/index"
	"pclsguserver/internal/model"
)

var ErrRebuildInProgress = errors.New("rebuild in progress")

type Service struct {
	builder *index.Builder
	audit   *audit.Service
}

func NewService(bld *index.Builder, a *audit.Service) *Service {
	return &Service{builder: bld, audit: a}
}

type RebuildResult struct {
	Success    bool   `json:"success"`
	Total      int    `json:"total"`
	DurationMs int64  `json:"durationMs"`
	Error      string `json:"error,omitempty"`
}

func (s *Service) Rebuild(operator, requestID, sourceIP string) (*RebuildResult, error) {
	result, err := s.builder.BuildFull()
	if err != nil {
		if errors.Is(err, index.ErrRebuildInProgress) {
			return nil, ErrRebuildInProgress
		}
		return &RebuildResult{Success: false, Error: err.Error()}, err
	}
	_ = s.audit.Record(nil, audit.Record{
		Operator:   operator,
		Action:     model.AuditActionIndexRebuild,
		TargetType: model.AuditTargetIndex,
		TargetID:   "index",
		Result:     model.AuditResultSuccess,
		Detail:     "rebuild index",
		SourceIP:   sourceIP,
		RequestID:  requestID,
	})
	return &RebuildResult{
		Success:    result.Success,
		Total:      result.Total,
		DurationMs: result.DurationMs,
	}, nil
}

type StatusResult struct {
	LastBuiltAt *time.Time `json:"lastBuiltAt"`
	Total       int        `json:"total"`
	Rebuilding  bool       `json:"rebuilding"`
}

func (s *Service) Status() *StatusResult {
	lastBuiltAt, total, rebuilding := s.builder.Status()
	var t *time.Time
	if !lastBuiltAt.IsZero() {
		t = &lastBuiltAt
	}
	return &StatusResult{
		LastBuiltAt: t,
		Total:       total,
		Rebuilding:  rebuilding,
	}
}
