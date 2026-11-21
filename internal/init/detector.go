package init

import (
	"os"
	"path/filepath"

	"pclsguserver/internal/pkg/logger"
)

func IsFirstRun(configPath string) bool {
	configExists := false
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			configExists = true
		}
	}
	markerPath := markerFilePath(configPath)
	markerExists := MarkerExists(markerPath)

	if !configExists && !markerExists {
		return true
	}
	if markerExists && !configExists {
		logger.L().Warn().Msg("marker exists but config.yaml missing, treating as first run")
		return true
	}
	if configExists && !markerExists {
		logger.L().Warn().Msg("config.yaml exists but .initialized marker missing, treating as first run")
		return true
	}
	return false
}

func GetInitStatus(configPath string) (isFirstRun bool, initStatus string) {
	if IsFirstRun(configPath) {
		return true, "uninitialized"
	}
	return false, "completed"
}

func markerFilePath(configPath string) string {
	dir := filepath.Dir(configPath)
	if dir == "" || dir == "." {
		dir = "."
	}
	return filepath.Join(dir, ".initialized")
}
