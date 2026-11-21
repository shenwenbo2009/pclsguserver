package repo

import (
	"pclsguserver/internal/model"

	"gorm.io/gorm"
)

type UploadTokenRepo struct{ db *gorm.DB }

func NewUploadTokenRepo(db *gorm.DB) *UploadTokenRepo { return &UploadTokenRepo{db: db} }

func (r *UploadTokenRepo) Create(tx *gorm.DB, token *model.UploadToken) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(token).Error
}

func (r *UploadTokenRepo) GetByTokenHash(tokenHash string) (*model.UploadToken, error) {
	var t model.UploadToken
	err := r.db.Where("tokenHash = ?", tokenHash).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *UploadTokenRepo) GetByID(id string) (*model.UploadToken, error) {
	var t model.UploadToken
	err := r.db.Where("id = ?", id).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *UploadTokenRepo) Update(tx *gorm.DB, token *model.UploadToken) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Save(token).Error
}

func (r *UploadTokenRepo) List(page, size int) ([]model.UploadToken, int64, error) {
	var total int64
	if err := r.db.Model(&model.UploadToken{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tokens []model.UploadToken
	err := r.db.Order("createdAt DESC").Offset((page - 1) * size).Limit(size).Find(&tokens).Error
	return tokens, total, err
}

type FileRecordRepo struct{ db *gorm.DB }

func NewFileRecordRepo(db *gorm.DB) *FileRecordRepo { return &FileRecordRepo{db: db} }

func (r *FileRecordRepo) Create(tx *gorm.DB, record *model.FileRecord) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(record).Error
}

func (r *FileRecordRepo) GetByID(id string) (*model.FileRecord, error) {
	var f model.FileRecord
	err := r.db.Where("id = ?", id).First(&f).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FileRecordRepo) Delete(tx *gorm.DB, id string) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Where("id = ?", id).Delete(&model.FileRecord{}).Error
}
