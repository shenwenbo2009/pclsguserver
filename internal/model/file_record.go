package model

import "time"

type FileRecord struct {
	ID               string      `gorm:"column:id;primaryKey" json:"id"`
	Purpose          FilePurpose `gorm:"column:purpose" json:"purpose"`
	OriginalFilename string      `gorm:"column:originalFilename" json:"originalFilename"`
	StoredPath       string      `gorm:"column:storedPath" json:"storedPath"`
	URL              string      `gorm:"column:url" json:"url"`
	Checksum         string      `gorm:"column:checksum" json:"checksum"`
	Size             int64       `gorm:"column:size" json:"size"`
	UploaderIdentity string      `gorm:"column:uploaderIdentity" json:"uploaderIdentity"`
	CreatedAt        time.Time   `gorm:"column:createdAt" json:"createdAt"`
}

func (FileRecord) TableName() string { return "file_records" }
