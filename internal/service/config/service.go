package configsvc

import (
	"errors"
	"strings"
	"time"

	"pclsguserver/internal/audit"
	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
)

var (
	ErrConfigNotFound       = errors.New("config not found")
	ErrInvalidConfigValue   = errors.New("invalid config value")
	ErrConfirmationRequired = errors.New("confirmation required")
)

type Service struct {
	repo  *repo.ConfigRepo
	audit *audit.Service
}

func NewService(r *repo.ConfigRepo, a *audit.Service) *Service {
	return &Service{repo: r, audit: a}
}

type ConfigItem struct {
	Key           string `json:"key"`
	Value         string `json:"value"`
	Sensitive     bool   `json:"sensitive"`
	HotReloadable bool   `json:"hotReloadable"`
	Description   string `json:"description"`
}

func (s *Service) List() ([]ConfigItem, error) {
	cfgs, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	items := make([]ConfigItem, len(cfgs))
	for i, c := range cfgs {
		items[i] = ConfigItem{
			Key:           c.Key,
			Value:         maskValue(c.Value, c.Sensitive),
			Sensitive:     c.Sensitive,
			HotReloadable: c.HotReloadable,
			Description:   c.Description,
		}
	}
	return items, nil
}

func maskValue(val string, sensitive bool) string {
	if !sensitive || len(val) <= 8 {
		return val
	}
	return val[:4] + "***" + val[len(val)-4:]
}

func (s *Service) Update(key, value string, confirm bool, operator, requestID, sourceIP string) (*ConfigItem, error) {
	c, err := s.repo.GetByKey(key)
	if err != nil {
		return nil, ErrConfigNotFound
	}
	if !validateConfigValue(key, value) {
		return nil, ErrInvalidConfigValue
	}
	if c.Sensitive && !confirm {
		return nil, ErrConfirmationRequired
	}
	oldValue := c.Value
	c.Value = value
	c.UpdatedAt = time.Now()
	if err := s.repo.Upsert(nil, c); err != nil {
		return nil, err
	}
	detail := "config " + key + " changed"
	if c.Sensitive {
		detail = "config " + key + " changed (sensitive, value masked)"
	} else {
		detail = "config " + key + " changed from " + oldValue + " to " + value
	}
	_ = s.audit.Record(nil, audit.Record{
		Operator:   operator,
		Action:     model.AuditActionConfigChanged,
		TargetType: model.AuditTargetConfig,
		TargetID:   key,
		Result:     model.AuditResultSuccess,
		Detail:     detail,
		SourceIP:   sourceIP,
		RequestID:  requestID,
	})
	return &ConfigItem{
		Key:           c.Key,
		Value:         maskValue(c.Value, c.Sensitive),
		Sensitive:     c.Sensitive,
		HotReloadable: c.HotReloadable,
		Description:   c.Description,
	}, nil
}

func validateConfigValue(key, value string) bool {
	if value == "" {
		return false
	}
	if strings.Contains(key, "rps") || strings.Contains(key, "Threshold") {
		for _, c := range value {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	if strings.Contains(key, "retentionDays") {
		for _, c := range value {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}
