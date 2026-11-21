package init

import (
	"encoding/json"
	"os"
	"time"
)

type InitMarker struct {
	CompletedAt   time.Time `json:"completedAt"`
	Version       int       `json:"version"`
	AdminUsername string    `json:"adminUsername"`
}

func CreateInitMarker(markerPath string, marker InitMarker) error {
	if marker.Version == 0 {
		marker.Version = 1
	}
	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(markerPath, data, 0o644)
}

func ReadInitMarker(markerPath string) (*InitMarker, error) {
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return nil, err
	}
	var marker InitMarker
	if err := json.Unmarshal(data, &marker); err != nil {
		return nil, err
	}
	return &marker, nil
}

func MarkerExists(markerPath string) bool {
	if _, err := os.Stat(markerPath); err == nil {
		return true
	}
	return false
}
