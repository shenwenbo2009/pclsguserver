package model

import "time"

type AuditLog struct {
	ID         string          `gorm:"column:id;primaryKey" json:"id"`
	Operator   string          `gorm:"column:operator;index:idx_audit_operator,priority:2" json:"operator"`
	Action     AuditAction     `gorm:"column:action;index:idx_audit_action,priority:2" json:"action"`
	TargetType AuditTargetType `gorm:"column:targetType" json:"targetType"`
	TargetID   string          `gorm:"column:targetId" json:"targetId"`
	Result     AuditResult     `gorm:"column:result" json:"result"`
	Detail     string          `gorm:"column:detail;type:text" json:"detail"`
	SourceIP   string          `gorm:"column:sourceIp" json:"sourceIp"`
	OccurredAt time.Time       `gorm:"column:occurredAt;index:idx_audit_occurred" json:"occurredAt"`
	RequestID  string          `gorm:"column:requestId" json:"requestId"`
}

func (AuditLog) TableName() string { return "audit_logs" }
