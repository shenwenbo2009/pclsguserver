package model

import "time"

type UploadToken struct {
	ID               string            `gorm:"column:id;primaryKey" json:"id"`
	TokenHash        string            `gorm:"column:tokenHash;uniqueIndex" json:"-"`
	Salt             string            `gorm:"column:salt" json:"-"`
	Name             string            `gorm:"column:name" json:"name"`
	UploaderIdentity string            `gorm:"column:uploaderIdentity" json:"uploaderIdentity"`
	Permissions      []string          `gorm:"column:permissions;serializer:json" json:"permissions"`
	Status           UploadTokenStatus `gorm:"column:status" json:"status"`
	CreatedAt        time.Time         `gorm:"column:createdAt" json:"createdAt"`
	DisabledAt       *time.Time        `gorm:"column:disabledAt" json:"disabledAt"`
}

func (UploadToken) TableName() string { return "upload_tokens" }
