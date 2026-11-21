package audit

import (
	"strings"
	"time"

	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repo *repo.AuditRepo
}

func NewService(auditRepo *repo.AuditRepo) *Service {
	return &Service{repo: auditRepo}
}

type Record struct {
	Operator   string
	Action     model.AuditAction
	TargetType model.AuditTargetType
	TargetID   string
	Result     model.AuditResult
	Detail     string
	SourceIP   string
	RequestID  string
}

func (s *Service) Record(tx *gorm.DB, r Record) error {
	log := &model.AuditLog{
		ID:         uuid.NewString(),
		Operator:   r.Operator,
		Action:     r.Action,
		TargetType: r.TargetType,
		TargetID:   r.TargetID,
		Result:     r.Result,
		Detail:     MaskSensitive(r.Detail),
		SourceIP:   r.SourceIP,
		OccurredAt: time.Now(),
		RequestID:  r.RequestID,
	}
	return s.repo.Append(tx, log)
}

func MaskSensitive(detail string) string {
	sensitiveKeys := []string{"password", "token", "secret", "apiKey", "privateKey"}
	result := detail
	for _, key := range sensitiveKeys {
		result = maskField(result, key)
	}
	return result
}

func maskField(s, key string) string {
	lower := strings.ToLower(s)
	idx := strings.Index(lower, strings.ToLower(key))
	if idx < 0 {
		return s
	}
	start := idx + len(key)
	for start < len(s) && (s[start] == ':' || s[start] == ' ' || s[start] == '"' || s[start] == '=') {
		start++
	}
	if start >= len(s) {
		return s
	}
	end := start
	for end < len(s) && s[end] != ',' && s[end] != '}' && s[end] != ' ' && s[end] != '\n' {
		end++
	}
	val := s[start:end]
	if len(val) <= 8 {
		return s
	}
	masked := val[:4] + "***" + val[len(val)-4:]
	return s[:start] + masked + s[end:]
}

func (s *Service) Query(from, to time.Time, action, operator, targetType string, page, size int) ([]model.AuditLog, int64, error) {
	return s.repo.Query(from, to, action, operator, targetType, page, size)
}

func (s *Service) CleanExpired(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return s.repo.CleanBefore(cutoff)
}
