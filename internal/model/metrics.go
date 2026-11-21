package model

import "time"

type MetricsEventRaw struct {
	ID         string           `gorm:"column:id;primaryKey" json:"id"`
	Dimension  MetricsDimension `gorm:"column:dimension" json:"dimension"`
	SubjectKey string           `gorm:"column:subjectKey" json:"subjectKey"`
	ClientID   string           `gorm:"column:clientId" json:"clientId"`
	OccurredAt time.Time        `gorm:"column:occurredAt;index:idx_metrics_raw_occurred" json:"occurredAt"`
	Aggregated bool             `gorm:"column:aggregated" json:"aggregated"`
}

func (MetricsEventRaw) TableName() string { return "metrics_event_raw" }

type MetricsEvent struct {
	ID                string           `gorm:"column:id;primaryKey" json:"id"`
	Dimension         MetricsDimension `gorm:"column:dimension;uniqueIndex:idx_metrics_agg" json:"dimension"`
	SubjectKey        string           `gorm:"column:subjectKey;uniqueIndex:idx_metrics_agg,priority:2" json:"subjectKey"`
	Count             int64            `gorm:"column:count" json:"count"`
	UniqueClientCount int              `gorm:"column:uniqueClientCount" json:"uniqueClientCount"`
	Granularity       Granularity      `gorm:"column:granularity;uniqueIndex:idx_metrics_agg,priority:3" json:"granularity"`
	BucketStart       time.Time        `gorm:"column:bucketStart;uniqueIndex:idx_metrics_agg,priority:4" json:"bucketStart"`
	BucketEnd         time.Time        `gorm:"column:bucketEnd" json:"bucketEnd"`
}

func (MetricsEvent) TableName() string { return "metrics_events" }
