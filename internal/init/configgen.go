package init

import (
	"fmt"
	"os"
	"time"

	"pclsguserver/internal/config"
)

func GenerateConfigYaml(cfg InitConfig) ([]byte, error) {
	c := buildConfigFromInitConfig(cfg)
	return configToYaml(c)
}

func WriteConfigYaml(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func buildConfigFromInitConfig(cfg InitConfig) *config.Config {
	dsn := cfg.SqlitePath
	if cfg.DbDriver == "postgres" {
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword, cfg.DbName)
	}
	if cfg.DbDriver == "mysql" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DbUser, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbName)
	}
	if dsn == "" && cfg.DbDriver == "sqlite" {
		dsn = "./data/meta.db"
	}

	storagePath := cfg.StoragePath
	if storagePath == "" {
		storagePath = "./data"
	}

	return &config.Config{
		Server: config.ServerConfig{
			Port:         8080,
			DataDir:      "./data",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		DB: config.DBConfig{
			Driver: cfg.DbDriver,
			DSN:    dsn,
		},
		Storage: config.StorageConfig{
			Driver:     cfg.StorageDriver,
			Path:       storagePath,
			S3Bucket:   cfg.S3Bucket,
			S3Region:   cfg.S3Region,
			S3Endpoint: cfg.S3Endpoint,
		},
		JWT: config.JWTConfig{
			Secret:     cfg.JwtSecret,
			AccessTtl:  2 * time.Hour,
			RefreshTtl: 168 * time.Hour,
		},
		RateLimit: config.RateLimitConfig{
			LoginThreshold: 5,
			LoginWindow:    5 * time.Minute,
			GlobalRPS:      100,
			DownloadRPS:    50,
		},
		Audit: config.AuditConfig{
			RetentionDays: 180,
		},
		Index: config.IndexConfig{
			RebuildInterval: 5 * time.Minute,
			FormatVersion:   1,
			DegradePolicy:   "db",
		},
		Metrics: config.MetricsConfig{
			AggregateInterval: 5 * time.Minute,
		},
		Admin: config.AdminConfig{
			SuperUsername: cfg.AdminUsername,
		},
	}
}

func configToYaml(c *config.Config) ([]byte, error) {
	var buf []byte
	buf = append(buf, []byte("server:\n")...)
	buf = append(buf, []byte(fmt.Sprintf("  port: %d\n", c.Server.Port))...)
	buf = append(buf, []byte("  dataDir: "+c.Server.DataDir+"\n")...)
	buf = append(buf, []byte("  readTimeout: "+c.Server.ReadTimeout.String()+"\n")...)
	buf = append(buf, []byte("  writeTimeout: "+c.Server.WriteTimeout.String()+"\n")...)

	buf = append(buf, []byte("db:\n")...)
	buf = append(buf, []byte("  driver: "+c.DB.Driver+"\n")...)
	buf = append(buf, []byte("  dsn: "+c.DB.DSN+"\n")...)

	buf = append(buf, []byte("storage:\n")...)
	buf = append(buf, []byte("  driver: "+c.Storage.Driver+"\n")...)
	buf = append(buf, []byte("  path: "+c.Storage.Path+"\n")...)
	if c.Storage.S3Bucket != "" {
		buf = append(buf, []byte("  s3Bucket: "+c.Storage.S3Bucket+"\n")...)
		buf = append(buf, []byte("  s3Region: "+c.Storage.S3Region+"\n")...)
		buf = append(buf, []byte("  s3Endpoint: "+c.Storage.S3Endpoint+"\n")...)
	}

	buf = append(buf, []byte("jwt:\n")...)
	buf = append(buf, []byte("  secret: "+c.JWT.Secret+"\n")...)
	buf = append(buf, []byte("  accessTtl: "+c.JWT.AccessTtl.String()+"\n")...)
	buf = append(buf, []byte("  refreshTtl: "+c.JWT.RefreshTtl.String()+"\n")...)

	buf = append(buf, []byte("rateLimit:\n")...)
	buf = append(buf, []byte(fmt.Sprintf("  loginThreshold: %d\n", c.RateLimit.LoginThreshold))...)
	buf = append(buf, []byte("  loginWindow: "+c.RateLimit.LoginWindow.String()+"\n")...)
	buf = append(buf, []byte(fmt.Sprintf("  globalRps: %d\n", c.RateLimit.GlobalRPS))...)
	buf = append(buf, []byte(fmt.Sprintf("  downloadRps: %d\n", c.RateLimit.DownloadRPS))...)

	buf = append(buf, []byte("audit:\n")...)
	buf = append(buf, []byte(fmt.Sprintf("  retentionDays: %d\n", c.Audit.RetentionDays))...)

	buf = append(buf, []byte("index:\n")...)
	buf = append(buf, []byte("  rebuildInterval: "+c.Index.RebuildInterval.String()+"\n")...)
	buf = append(buf, []byte(fmt.Sprintf("  formatVersion: %d\n", c.Index.FormatVersion))...)
	buf = append(buf, []byte("  degradePolicy: "+c.Index.DegradePolicy+"\n")...)

	buf = append(buf, []byte("metrics:\n")...)
	buf = append(buf, []byte("  aggregateInterval: "+c.Metrics.AggregateInterval.String()+"\n")...)

	buf = append(buf, []byte("admin:\n")...)
	buf = append(buf, []byte("  superUsername: "+c.Admin.SuperUsername+"\n")...)

	return buf, nil
}
