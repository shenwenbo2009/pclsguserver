package repo

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pclsguserver/internal/config"
	"pclsguserver/internal/model"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}
	var db *gorm.DB
	var err error
	switch cfg.DB.Driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DB.DSN), gormCfg)
	case "mysql":
		if cfg.DB.DSN == "" {
			return nil, fmt.Errorf("MySQL DSN 不能为空")
		}
		db, err = gorm.Open(mysql.Open(cfg.DB.DSN), gormCfg)
	case "sqlite":
		if cfg.DB.DSN == "" {
			cfg.DB.DSN = "./data/meta.db"
		}
		db, err = gorm.Open(sqlite.Open(cfg.DB.DSN), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", cfg.DB.Driver)
	}
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := Migrate(db); err != nil {
		dropAllTables(db)
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := Seed(db, cfg); err != nil {
		dropAllTables(db)
		return nil, fmt.Errorf("seed: %w", err)
	}
	return db, nil
}

func dropAllTables(db *gorm.DB) {
	tables := []string{
		"software_versions", "plugins", "review_records",
		"admin_users", "roles", "user_roles",
		"audit_logs", "metrics_event_raw", "metrics_events",
		"system_configs", "upload_tokens", "file_records",
	}
	for _, t := range tables {
		db.Exec("DROP TABLE IF EXISTS " + t)
	}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.SoftwareVersion{},
		&model.Plugin{},
		&model.ReviewRecord{},
		&model.AdminUser{},
		&model.Role{},
		&model.UserRole{},
		&model.AuditLog{},
		&model.MetricsEventRaw{},
		&model.MetricsEvent{},
		&model.SystemConfig{},
		&model.UploadToken{},
		&model.FileRecord{},
	)
}

func Seed(db *gorm.DB, cfg *config.Config) error {
	if err := seedRoles(db); err != nil {
		return err
	}
	if err := seedSuperAdmin(db, cfg); err != nil {
		return err
	}
	return seedConfigs(db)
}

func seedRoles(db *gorm.DB) error {
	roles := []model.Role{
		{ID: "role-super-admin", Name: "超级管理员", Permissions: []string{"*"}, Builtin: true, Description: "拥有全部权限"},
		{ID: "role-ops", Name: "运营", Permissions: []string{"software.publish", "plugin.manage", "index.rebuild", "index.view", "metrics.view"}, Builtin: true, Description: "运营管理"},
		{ID: "role-reviewer", Name: "审核员", Permissions: []string{"plugin.review", "plugin.submit", "index.view"}, Builtin: true, Description: "插件审核"},
		{ID: "role-stats", Name: "统计查看员", Permissions: []string{"metrics.view", "audit.view"}, Builtin: true, Description: "统计查看"},
	}
	for _, r := range roles {
		var existing model.Role
		if err := db.Where("id = ?", r.ID).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&r).Error; err != nil {
			return fmt.Errorf("seed role %s: %w", r.Name, err)
		}
	}
	return nil
}

func seedSuperAdmin(db *gorm.DB, cfg *config.Config) error {
	username := cfg.Admin.SuperUsername
	if username == "" {
		username = "admin"
	}
	password := cfg.Admin.SuperPassword
	if password == "" {
		password = "admin123"
	}
	var existing model.AdminUser
	if err := db.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	admin := model.AdminUser{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hashed),
		Status:       model.AccountStatusActive,
		CreatedAt:    time.Now(),
	}
	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	return nil
}

func seedConfigs(db *gorm.DB) error {
	configs := []model.SystemConfig{
		{Key: "rateLimit.globalRps", Value: "100", Sensitive: false, HotReloadable: true, Description: "全局限流RPS", UpdatedAt: time.Now()},
		{Key: "rateLimit.downloadRps", Value: "50", Sensitive: false, HotReloadable: true, Description: "下载限流RPS", UpdatedAt: time.Now()},
		{Key: "rateLimit.loginThreshold", Value: "5", Sensitive: false, HotReloadable: true, Description: "登录失败锁定阈值", UpdatedAt: time.Now()},
		{Key: "audit.retentionDays", Value: "180", Sensitive: false, HotReloadable: false, Description: "审计保留天数", UpdatedAt: time.Now()},
		{Key: "index.rebuildInterval", Value: "5m", Sensitive: false, HotReloadable: true, Description: "索引重建间隔", UpdatedAt: time.Now()},
		{Key: "jwt.secret", Value: "change-me-in-production", Sensitive: true, HotReloadable: false, Description: "JWT密钥", UpdatedAt: time.Now()},
	}
	for _, c := range configs {
		var existing model.SystemConfig
		if err := db.Where("key = ?", c.Key).First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&c).Error; err != nil {
			return fmt.Errorf("seed config %s: %w", c.Key, err)
		}
	}
	return nil
}
