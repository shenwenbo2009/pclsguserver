package init

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"pclsguserver/internal/pkg/logger"
	"pclsguserver/internal/repo"
)

type InitConfig struct {
	AdminUsername string `json:"adminUsername"`
	AdminPassword string `json:"adminPassword"`
	JwtSecret     string `json:"jwtSecret"`
	DbDriver      string `json:"dbDriver"`
	SqlitePath    string `json:"sqlitePath"`
	DbHost        string `json:"dbHost"`
	DbPort        string `json:"dbPort"`
	DbName        string `json:"dbName"`
	DbUser        string `json:"dbUser"`
	DbPassword    string `json:"dbPassword"`
	StorageDriver string `json:"storageDriver"`
	StoragePath   string `json:"storagePath"`
	S3Bucket      string `json:"s3Bucket"`
	S3Region      string `json:"s3Region"`
	S3Endpoint    string `json:"s3Endpoint"`
}

type InitExecutor struct {
	configPath string
	markerPath string
}

func NewInitExecutor(configPath string) *InitExecutor {
	return &InitExecutor{
		configPath: configPath,
		markerPath: markerFilePath(configPath),
	}
}

func (e *InitExecutor) Execute(cfg InitConfig) error {
	for step := 1; step <= 4; step++ {
		data := initConfigToStepData(cfg, step)
		valid, errs := ValidateStep(step, data)
		if !valid {
			return fmt.Errorf("step %d validation failed: %v", step, errs)
		}
	}

	yamlData, err := GenerateConfigYaml(cfg)
	if err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(e.configPath), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := WriteConfigYaml(e.configPath, yamlData); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	c := buildConfigFromInitConfig(cfg)
	dataDir := c.Server.DataDir
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		os.Remove(e.configPath)
		return fmt.Errorf("create data dir: %w", err)
	}

	db, err := repo.InitDB(c)
	if err != nil {
		os.Remove(e.configPath)
		return fmt.Errorf("init db: %w", err)
	}

	if err := repo.Migrate(db); err != nil {
		os.Remove(e.configPath)
		return fmt.Errorf("migrate: %w", err)
	}

	if err := repo.Seed(db, c); err != nil {
		os.Remove(e.configPath)
		return fmt.Errorf("seed: %w", err)
	}

	marker := InitMarker{
		CompletedAt:   time.Now(),
		Version:       1,
		AdminUsername: cfg.AdminUsername,
	}
	if err := CreateInitMarker(e.markerPath, marker); err != nil {
		os.Remove(e.configPath)
		return fmt.Errorf("create marker: %w", err)
	}

	logger.L().Info().Str("adminUser", cfg.AdminUsername).Msg("initialization completed")
	return nil
}

func initConfigToStepData(cfg InitConfig, step int) map[string]interface{} {
	switch step {
	case 1:
		return map[string]interface{}{
			"adminUsername":   cfg.AdminUsername,
			"adminPassword":   cfg.AdminPassword,
			"passwordConfirm": cfg.AdminPassword,
		}
	case 2:
		mode := "manual"
		if cfg.JwtSecret != "" && len(cfg.JwtSecret) >= 32 {
			mode = "auto"
		}
		return map[string]interface{}{
			"jwtMode":   mode,
			"jwtSecret": cfg.JwtSecret,
		}
	case 3:
		return map[string]interface{}{
			"dbDriver":   cfg.DbDriver,
			"sqlitePath": cfg.SqlitePath,
			"dbHost":     cfg.DbHost,
			"dbPort":     cfg.DbPort,
			"dbName":     cfg.DbName,
			"dbUser":     cfg.DbUser,
			"dbPassword": cfg.DbPassword,
		}
	case 4:
		return map[string]interface{}{
			"storageDriver": cfg.StorageDriver,
			"storagePath":   cfg.StoragePath,
			"s3Bucket":      cfg.S3Bucket,
			"s3Region":      cfg.S3Region,
			"s3Endpoint":    cfg.S3Endpoint,
		}
	default:
		return nil
	}
}

var ErrAlreadyInitialized = errors.New("already initialized")
