package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	DB        DBConfig        `mapstructure:"db"`
	Storage   StorageConfig   `mapstructure:"storage"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"rateLimit"`
	Audit     AuditConfig     `mapstructure:"audit"`
	Index     IndexConfig     `mapstructure:"index"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	Admin     AdminConfig     `mapstructure:"admin"`
	Upload    UploadConfig    `mapstructure:"upload"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	DataDir      string        `mapstructure:"dataDir"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
}

type DBConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

type StorageConfig struct {
	Driver     string `mapstructure:"driver"`
	Path       string `mapstructure:"path"`
	S3Bucket   string `mapstructure:"s3Bucket"`
	S3Region   string `mapstructure:"s3Region"`
	S3Endpoint string `mapstructure:"s3Endpoint"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	AccessTtl  time.Duration `mapstructure:"accessTtl"`
	RefreshTtl time.Duration `mapstructure:"refreshTtl"`
}

type RateLimitConfig struct {
	LoginThreshold int           `mapstructure:"loginThreshold"`
	LoginWindow    time.Duration `mapstructure:"loginWindow"`
	GlobalRPS      int           `mapstructure:"globalRps"`
	DownloadRPS    int           `mapstructure:"downloadRps"`
}

type AuditConfig struct {
	RetentionDays int    `mapstructure:"retentionDays"`
	SiemWebhook   string `mapstructure:"siemWebhook"`
	SiemEnabled   bool   `mapstructure:"siemEnabled"`
}

type IndexConfig struct {
	RebuildInterval time.Duration `mapstructure:"rebuildInterval"`
	FormatVersion   int           `mapstructure:"formatVersion"`
	DegradePolicy   string        `mapstructure:"degradePolicy"`
}

type MetricsConfig struct {
	AggregateInterval time.Duration `mapstructure:"aggregateInterval"`
}

type AdminConfig struct {
	SuperUsername string `mapstructure:"superUsername"`
	SuperPassword string `mapstructure:"superPassword"`
}

type UploadConfig struct {
	PluginMaxSize      int64         `mapstructure:"pluginMaxSize"`
	SoftwareMaxSize    int64         `mapstructure:"softwareMaxSize"`
	FileMaxSize        int64         `mapstructure:"fileMaxSize"`
	TokenRateLimit     int           `mapstructure:"tokenRateLimit"`
	IPRateLimit        int           `mapstructure:"ipRateLimit"`
	RateLimitWindow    time.Duration `mapstructure:"rateLimitWindow"`
	PluginAllowedTypes []string      `mapstructure:"pluginAllowedTypes"`
	FileAllowedTypes   []string      `mapstructure:"fileAllowedTypes"`
}

func defaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.dataDir", "./data")
	viper.SetDefault("server.readTimeout", "30s")
	viper.SetDefault("server.writeTimeout", "30s")
	viper.SetDefault("db.driver", "sqlite")
	viper.SetDefault("db.dsn", "./data/meta.db")
	viper.SetDefault("storage.driver", "fs")
	viper.SetDefault("storage.path", "./data")
	viper.SetDefault("jwt.secret", "change-me-in-production")
	viper.SetDefault("jwt.accessTtl", "2h")
	viper.SetDefault("jwt.refreshTtl", "168h")
	viper.SetDefault("rateLimit.loginThreshold", 5)
	viper.SetDefault("rateLimit.loginWindow", "5m")
	viper.SetDefault("rateLimit.globalRps", 100)
	viper.SetDefault("rateLimit.downloadRps", 50)
	viper.SetDefault("audit.retentionDays", 180)
	viper.SetDefault("audit.siemEnabled", false)
	viper.SetDefault("index.rebuildInterval", "5m")
	viper.SetDefault("index.formatVersion", 1)
	viper.SetDefault("index.degradePolicy", "db")
	viper.SetDefault("metrics.aggregateInterval", "5m")
	viper.SetDefault("admin.superUsername", "admin")
	viper.SetDefault("admin.superPassword", "admin123")
	viper.SetDefault("upload.pluginMaxSize", 209715200)
	viper.SetDefault("upload.softwareMaxSize", 524288000)
	viper.SetDefault("upload.fileMaxSize", 20971520)
	viper.SetDefault("upload.tokenRateLimit", 10)
	viper.SetDefault("upload.ipRateLimit", 20)
	viper.SetDefault("upload.rateLimitWindow", "1m")
	viper.SetDefault("upload.pluginAllowedTypes", []string{".zip", ".tar.gz", ".tgz", ".7z", ".rar"})
	viper.SetDefault("upload.fileAllowedTypes", []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx", ".txt", ".csv", ".xlsx"})
}

func Load(path string) (*Config, error) {
	defaults()
	if path != "" {
		viper.SetConfigFile(path)
		if err := viper.ReadInConfig(); err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				return nil, fmt.Errorf("read config: %w", err)
			}
		}
	}
	viper.SetEnvPrefix("PCL")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be positive")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret must not be empty")
	}
	if c.JWT.AccessTtl <= 0 {
		c.JWT.AccessTtl = 2 * time.Hour
	}
	if c.JWT.RefreshTtl <= 0 {
		c.JWT.RefreshTtl = 168 * time.Hour
	}
	if c.Index.FormatVersion <= 0 {
		c.Index.FormatVersion = 1
	}
	if c.Audit.RetentionDays < 180 {
		c.Audit.RetentionDays = 180
	}
	return nil
}
