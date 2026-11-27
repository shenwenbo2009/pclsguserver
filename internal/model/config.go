package model

import "time"

type SystemConfig struct {
	Key           string    `gorm:"column:config_key;primaryKey" json:"key"`
	Value         string    `gorm:"column:value;type:text" json:"value"`
	Sensitive     bool      `gorm:"column:sensitive" json:"sensitive"`
	HotReloadable bool      `gorm:"column:hotReloadable" json:"hotReloadable"`
	Description   string    `gorm:"column:description;type:text" json:"description"`
	UpdatedAt     time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}

func (SystemConfig) TableName() string { return "system_configs" }
