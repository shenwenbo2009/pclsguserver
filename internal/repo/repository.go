package repo

import (
	"time"

	"pclsguserver/internal/model"

	"gorm.io/gorm"
)

type SoftwareVersionRepo struct{ db *gorm.DB }

func NewSoftwareVersionRepo(db *gorm.DB) *SoftwareVersionRepo { return &SoftwareVersionRepo{db: db} }

func (r *SoftwareVersionRepo) Create(tx *gorm.DB, v *model.SoftwareVersion) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(v).Error
}

func (r *SoftwareVersionRepo) FindLatestPublished(swCode string) (*model.SoftwareVersion, error) {
	var v model.SoftwareVersion
	err := r.db.Where("swCode = ? AND status = ?", swCode, model.SoftwareStatusPublished).
		Order("createdAt DESC").First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SoftwareVersionRepo) GetByVersion(swCode, version string) (*model.SoftwareVersion, error) {
	var v model.SoftwareVersion
	err := r.db.Where("swCode = ? AND version = ?", swCode, version).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SoftwareVersionRepo) GetByID(id string) (*model.SoftwareVersion, error) {
	var v model.SoftwareVersion
	err := r.db.Where("id = ?", id).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SoftwareVersionRepo) ListBySwCode(swCode string) ([]model.SoftwareVersion, error) {
	var vs []model.SoftwareVersion
	err := r.db.Where("swCode = ?", swCode).Order("createdAt DESC").Find(&vs).Error
	return vs, err
}

func (r *SoftwareVersionRepo) ListAllPublished() ([]model.SoftwareVersion, error) {
	var vs []model.SoftwareVersion
	err := r.db.Where("status = ?", model.SoftwareStatusPublished).Find(&vs).Error
	return vs, err
}

func (r *SoftwareVersionRepo) ListAll(page, size int) ([]model.SoftwareVersion, int64, error) {
	var total int64
	if err := r.db.Model(&model.SoftwareVersion{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var vs []model.SoftwareVersion
	err := r.db.Order("createdAt DESC").Offset((page - 1) * size).Limit(size).Find(&vs).Error
	return vs, total, err
}

func (r *SoftwareVersionRepo) Update(tx *gorm.DB, v *model.SoftwareVersion) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Save(v).Error
}

func (r *SoftwareVersionRepo) Delete(tx *gorm.DB, id string) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Where("id = ?", id).Delete(&model.SoftwareVersion{}).Error
}

func (r *SoftwareVersionRepo) IncrementDownloadCount(swCode, version string) error {
	return r.db.Model(&model.SoftwareVersion{}).
		Where("swCode = ? AND version = ?", swCode, version).
		UpdateColumn("downloadCount", gorm.Expr("downloadCount + 1")).Error
}

type PluginRepo struct{ db *gorm.DB }

func NewPluginRepo(db *gorm.DB) *PluginRepo { return &PluginRepo{db: db} }

func (r *PluginRepo) Create(tx *gorm.DB, p *model.Plugin) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(p).Error
}

func (r *PluginRepo) GetByID(id string) (*model.Plugin, error) {
	var p model.Plugin
	err := r.db.Where("id = ?", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PluginRepo) Update(tx *gorm.DB, p *model.Plugin) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Save(p).Error
}

func (r *PluginRepo) Delete(tx *gorm.DB, id string) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Where("id = ?", id).Delete(&model.Plugin{}).Error
}

func (r *PluginRepo) ListPublished(page, size int, category string) ([]model.Plugin, int64, error) {
	q := r.db.Model(&model.Plugin{}).Where("status = ?", model.PluginStatusPublished)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var ps []model.Plugin
	err := q.Order("downloadCount DESC, createdAt DESC").Offset((page - 1) * size).Limit(size).Find(&ps).Error
	return ps, total, err
}

func (r *PluginRepo) ListAllPublished() ([]model.Plugin, error) {
	var ps []model.Plugin
	err := r.db.Where("status = ?", model.PluginStatusPublished).Find(&ps).Error
	return ps, err
}

func (r *PluginRepo) ListAll(page, size int) ([]model.Plugin, int64, error) {
	var total int64
	if err := r.db.Model(&model.Plugin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var ps []model.Plugin
	err := r.db.Order("createdAt DESC").Offset((page - 1) * size).Limit(size).Find(&ps).Error
	return ps, total, err
}

func (r *PluginRepo) ListReviewing(page, size int) ([]model.Plugin, int64, error) {
	q := r.db.Model(&model.Plugin{}).Where("status = ?", model.PluginStatusReviewing)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var ps []model.Plugin
	err := q.Order("createdAt ASC").Offset((page - 1) * size).Limit(size).Find(&ps).Error
	return ps, total, err
}

func (r *PluginRepo) SearchFallback(keyword, category, name string, page, size int) ([]model.Plugin, int64, error) {
	q := r.db.Model(&model.Plugin{}).Where("status = ?", model.PluginStatusPublished)
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if keyword != "" {
		q = q.Where("name LIKE ? OR description LIKE ? OR keywords LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var ps []model.Plugin
	err := q.Order("downloadCount DESC").Offset((page - 1) * size).Limit(size).Find(&ps).Error
	return ps, total, err
}

func (r *PluginRepo) IncrementDownloadCount(id string) error {
	return r.db.Model(&model.Plugin{}).Where("id = ?", id).
		UpdateColumn("downloadCount", gorm.Expr("downloadCount + 1")).Error
}

type ReviewRecordRepo struct{ db *gorm.DB }

func NewReviewRecordRepo(db *gorm.DB) *ReviewRecordRepo { return &ReviewRecordRepo{db: db} }

func (r *ReviewRecordRepo) Create(tx *gorm.DB, rr *model.ReviewRecord) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(rr).Error
}

func (r *ReviewRecordRepo) ListByPlugin(pluginId string) ([]model.ReviewRecord, error) {
	var rrs []model.ReviewRecord
	err := r.db.Where("pluginId = ?", pluginId).Order("reviewedAt DESC").Find(&rrs).Error
	return rrs, err
}

type AdminUserRepo struct{ db *gorm.DB }

func NewAdminUserRepo(db *gorm.DB) *AdminUserRepo { return &AdminUserRepo{db: db} }

func (r *AdminUserRepo) GetByUsername(username string) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.Preload("Roles").Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AdminUserRepo) GetByID(id string) (*model.AdminUser, error) {
	var u model.AdminUser
	err := r.db.Preload("Roles").Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AdminUserRepo) Create(tx *gorm.DB, u *model.AdminUser) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(u).Error
}

func (r *AdminUserRepo) Update(tx *gorm.DB, u *model.AdminUser) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Save(u).Error
}

func (r *AdminUserRepo) Delete(tx *gorm.DB, id string) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Where("id = ?", id).Delete(&model.AdminUser{}).Error
}

func (r *AdminUserRepo) List(page, size int) ([]model.AdminUser, int64, error) {
	var total int64
	if err := r.db.Model(&model.AdminUser{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var us []model.AdminUser
	err := r.db.Preload("Roles").Order("createdAt DESC").Offset((page - 1) * size).Limit(size).Find(&us).Error
	return us, total, err
}

func (r *AdminUserRepo) AssignRoles(tx *gorm.DB, userID string, roleIDs []string) error {
	if tx == nil {
		tx = r.db
	}
	if err := tx.Where("userId = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return err
	}
	for _, rid := range roleIDs {
		if err := tx.Create(&model.UserRole{UserID: userID, RoleID: rid}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *AdminUserRepo) CountSuperAdmins() (int64, error) {
	var count int64
	err := r.db.Model(&model.UserRole{}).Where("roleId = ?", "role-super-admin").Count(&count).Error
	return count, err
}

type RoleRepo struct{ db *gorm.DB }

func NewRoleRepo(db *gorm.DB) *RoleRepo { return &RoleRepo{db: db} }

func (r *RoleRepo) List() ([]model.Role, error) {
	var rs []model.Role
	err := r.db.Order("builtin DESC, name ASC").Find(&rs).Error
	return rs, err
}

func (r *RoleRepo) GetByID(id string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepo) Create(tx *gorm.DB, role *model.Role) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(role).Error
}

func (r *RoleRepo) Update(tx *gorm.DB, role *model.Role) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Save(role).Error
}

func (r *RoleRepo) Delete(tx *gorm.DB, id string) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Where("id = ?", id).Delete(&model.Role{}).Error
}

func (r *RoleRepo) GetPermissionsByUser(userID string) ([]string, error) {
	var roles []model.Role
	err := r.db.Joins("JOIN user_roles ON user_roles.roleId = roles.id").
		Where("user_roles.userId = ?", userID).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	permSet := make(map[string]bool)
	for _, role := range roles {
		for _, p := range role.Permissions {
			permSet[p] = true
		}
	}
	var perms []string
	for p := range permSet {
		perms = append(perms, p)
	}
	return perms, nil
}

type AuditRepo struct{ db *gorm.DB }

func NewAuditRepo(db *gorm.DB) *AuditRepo { return &AuditRepo{db: db} }

func (r *AuditRepo) Append(tx *gorm.DB, a *model.AuditLog) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(a).Error
}

func (r *AuditRepo) Query(from, to time.Time, action, operator, targetType string, page, size int) ([]model.AuditLog, int64, error) {
	q := r.db.Model(&model.AuditLog{})
	if !from.IsZero() {
		q = q.Where("occurredAt >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("occurredAt <= ?", to)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if operator != "" {
		q = q.Where("operator = ?", operator)
	}
	if targetType != "" {
		q = q.Where("targetType = ?", targetType)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []model.AuditLog
	err := q.Order("occurredAt DESC").Offset((page - 1) * size).Limit(size).Find(&logs).Error
	return logs, total, err
}

func (r *AuditRepo) CleanBefore(t time.Time) error {
	return r.db.Where("occurredAt < ?", t).Delete(&model.AuditLog{}).Error
}

type MetricsRepo struct{ db *gorm.DB }

func NewMetricsRepo(db *gorm.DB) *MetricsRepo { return &MetricsRepo{db: db} }

func (r *MetricsRepo) AppendRaw(tx *gorm.DB, e *model.MetricsEventRaw) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(e).Error
}

func (r *MetricsRepo) ListUnaggregated(limit int) ([]model.MetricsEventRaw, error) {
	var events []model.MetricsEventRaw
	err := r.db.Where("aggregated = ?", false).Limit(limit).Find(&events).Error
	return events, err
}

func (r *MetricsRepo) MarkAggregated(ids []string) error {
	return r.db.Model(&model.MetricsEventRaw{}).Where("id IN ?", ids).Update("aggregated", true).Error
}

func (r *MetricsRepo) UpsertAggregated(tx *gorm.DB, e *model.MetricsEvent) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Save(e).Error
}

func (r *MetricsRepo) QueryAggregated(dimension, subjectKey string, from, to time.Time, granularity string) ([]model.MetricsEvent, error) {
	var events []model.MetricsEvent
	q := r.db.Model(&model.MetricsEvent{})
	if dimension != "" {
		q = q.Where("dimension = ?", dimension)
	}
	if subjectKey != "" {
		q = q.Where("subjectKey = ?", subjectKey)
	}
	if granularity != "" {
		q = q.Where("granularity = ?", granularity)
	}
	if !from.IsZero() {
		q = q.Where("bucketStart >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("bucketStart <= ?", to)
	}
	err := q.Order("bucketStart ASC").Find(&events).Error
	return events, err
}

type ConfigRepo struct{ db *gorm.DB }

func NewConfigRepo(db *gorm.DB) *ConfigRepo { return &ConfigRepo{db: db} }

func (r *ConfigRepo) GetAll() ([]model.SystemConfig, error) {
	var cfgs []model.SystemConfig
	err := r.db.Order("config_key ASC").Find(&cfgs).Error
	return cfgs, err
}

func (r *ConfigRepo) GetByKey(key string) (*model.SystemConfig, error) {
	var c model.SystemConfig
	err := r.db.Where("config_key = ?", key).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ConfigRepo) Upsert(tx *gorm.DB, c *model.SystemConfig) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Save(c).Error
}
