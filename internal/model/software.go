package model

import "time"

type SoftwareVersion struct {
	ID            string         `gorm:"column:id;primaryKey" json:"id"`
	SwCode        string         `gorm:"column:swCode;index:idx_sw_code" json:"swCode"`
	Version       string         `gorm:"column:version" json:"version"`
	SwUrl         string         `gorm:"column:swUrl" json:"swUrl"`
	Checksum      string         `gorm:"column:checksum" json:"checksum"`
	ReleaseNote   string         `gorm:"column:releaseNote" json:"releaseNote"`
	PublishedAt   *time.Time     `gorm:"column:publishedAt" json:"publishedAt"`
	Status        SoftwareStatus `gorm:"column:status;index:idx_sw_status,priority:2" json:"status"`
	Size          int64          `gorm:"column:size" json:"size"`
	DownloadCount int64          `gorm:"column:downloadCount" json:"downloadCount"`
	CreatedAt     time.Time      `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"column:updatedAt" json:"updatedAt"`
}

func (SoftwareVersion) TableName() string { return "software_versions" }
