package model

import "time"

type Plugin struct {
	ID            string       `gorm:"column:id;primaryKey" json:"id"`
	Name          string       `gorm:"column:name;type:varchar(255);index:idx_plugin_name" json:"name"`
	Version       string       `gorm:"column:version" json:"version"`
	Author        string       `gorm:"column:author" json:"author"`
	Category      string       `gorm:"column:category;type:varchar(100);index:idx_plugin_category" json:"category"`
	Keywords      []string     `gorm:"column:keywords;type:text;serializer:json" json:"keywords"`
	Description   string       `gorm:"column:description;type:text" json:"description"`
	DownloadUrl   string       `gorm:"column:downloadUrl" json:"downloadUrl"`
	Checksum      string       `gorm:"column:checksum" json:"checksum"`
	Status        PluginStatus `gorm:"column:status;index:idx_plugin_status" json:"status"`
	PublishedAt   *time.Time   `gorm:"column:publishedAt" json:"publishedAt"`
	Size          int64        `gorm:"column:size" json:"size"`
	DownloadCount int64        `gorm:"column:downloadCount" json:"downloadCount"`
	RejectReason  string       `gorm:"column:rejectReason;type:text" json:"rejectReason"`
	SubmitterId   string       `gorm:"column:submitterId" json:"submitterId"`
	CreatedAt     time.Time    `gorm:"column:createdAt" json:"createdAt"`
	UpdatedAt     time.Time    `gorm:"column:updatedAt" json:"updatedAt"`
}

func (Plugin) TableName() string { return "plugins" }

type ReviewRecord struct {
	ID         string           `gorm:"column:id;primaryKey" json:"id"`
	PluginID   string           `gorm:"column:pluginId;type:varchar(255);index:idx_review_plugin" json:"pluginId"`
	Reviewer   string           `gorm:"column:reviewer" json:"reviewer"`
	Conclusion ReviewConclusion `gorm:"column:conclusion" json:"conclusion"`
	Reason     string           `gorm:"column:reason;type:text" json:"reason"`
	Comment    string           `gorm:"column:comment;type:text" json:"comment"`
	ReviewedAt time.Time        `gorm:"column:reviewedAt" json:"reviewedAt"`
}

func (ReviewRecord) TableName() string { return "review_records" }
