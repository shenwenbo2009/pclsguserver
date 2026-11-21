package index

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"pclsguserver/internal/model"
	"pclsguserver/internal/pkg/logger"
	"pclsguserver/internal/repo"
)

var ErrRebuildInProgress = errors.New("rebuild in progress")

type IndexEvent struct {
	Type     string
	PluginID string
}

type Builder struct {
	engine        *Engine
	pluginRepo    *repo.PluginRepo
	softwareRepo  *repo.SoftwareVersionRepo
	dataDir       string
	formatVersion int
	rebuildMu     sync.Mutex
	rebuilding    atomic.Bool
	lastBuiltAt   time.Time
	lastTotal     int
	updateCh      chan IndexEvent
}

func NewBuilder(engine *Engine, pluginRepo *repo.PluginRepo, softwareRepo *repo.SoftwareVersionRepo, dataDir string, formatVersion int) *Builder {
	b := &Builder{
		engine:        engine,
		pluginRepo:    pluginRepo,
		softwareRepo:  softwareRepo,
		dataDir:       dataDir,
		formatVersion: formatVersion,
		updateCh:      make(chan IndexEvent, 256),
	}
	go b.processUpdates()
	return b
}

func (b *Builder) Events() chan<- IndexEvent { return b.updateCh }

func (b *Builder) processUpdates() {
	for ev := range b.updateCh {
		_ = ev
		_, _ = b.BuildFull()
	}
}

type RebuildResult struct {
	Success    bool   `json:"success"`
	Total      int    `json:"total"`
	DurationMs int64  `json:"durationMs"`
	Error      string `json:"error,omitempty"`
}

func (b *Builder) BuildFull() (*RebuildResult, error) {
	if !b.rebuilding.CompareAndSwap(false, true) {
		return nil, ErrRebuildInProgress
	}
	b.rebuildMu.Lock()
	defer func() {
		b.rebuildMu.Unlock()
		b.rebuilding.Store(false)
	}()

	start := time.Now()

	plugins, err := b.pluginRepo.ListAllPublished()
	if err != nil {
		return &RebuildResult{Success: false, Error: err.Error()}, err
	}
	softwareUpdates, err := b.softwareRepo.ListAllPublished()
	if err != nil {
		return &RebuildResult{Success: false, Error: err.Error()}, err
	}

	items := make([]*PluginIndexItem, 0, len(plugins))
	for i := range plugins {
		p := &plugins[i]
		items = append(items, &PluginIndexItem{
			ID:            p.ID,
			Name:          p.Name,
			Version:       p.Version,
			Author:        p.Author,
			Category:      p.Category,
			Keywords:      p.Keywords,
			Description:   p.Description,
			DownloadUrl:   p.DownloadUrl,
			Checksum:      p.Checksum,
			Size:          p.Size,
			DownloadCount: p.DownloadCount,
		})
	}

	mi := BuildMemoryIndex(items)

	swItems := make([]SoftwareIndexItem, 0, len(softwareUpdates))
	for _, sv := range softwareUpdates {
		swItems = append(swItems, SoftwareIndexItem{
			SwCode:        sv.SwCode,
			LatestVersion: sv.Version,
			SwUrl:         sv.SwUrl,
			Checksum:      sv.Checksum,
		})
	}

	idx := IndexJson{
		Version:         b.formatVersion,
		GeneratedAt:     time.Now(),
		Total:           len(items),
		Plugins:         toPluginItems(items),
		SoftwareUpdates: swItems,
	}

	if err := b.writeIndexJson(&idx); err != nil {
		logger.L().Error().Err(err).Msg("write index.json failed")
		return &RebuildResult{Success: false, Error: err.Error()}, err
	}

	b.engine.Store(mi)
	b.lastBuiltAt = time.Now()
	b.lastTotal = len(items)

	duration := time.Since(start)
	return &RebuildResult{
		Success:    true,
		Total:      len(items),
		DurationMs: duration.Milliseconds(),
	}, nil
}

func toPluginItems(items []*PluginIndexItem) []PluginIndexItem {
	result := make([]PluginIndexItem, len(items))
	for i, p := range items {
		result[i] = *p
	}
	return result
}

func (b *Builder) writeIndexJson(idx *IndexJson) error {
	if err := os.MkdirAll(b.dataDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := filepath.Join(b.dataDir, "index.json.tmp")
	finalPath := filepath.Join(b.dataDir, "index.json")
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, finalPath)
}

func (b *Builder) Status() (time.Time, int, bool) {
	return b.lastBuiltAt, b.lastTotal, b.rebuilding.Load()
}

func (b *Builder) LoadIndexJson() ([]byte, error) {
	return os.ReadFile(filepath.Join(b.dataDir, "index.json"))
}

func (b *Builder) NotifyPluginPublished(pluginID string) {
	select {
	case b.updateCh <- IndexEvent{Type: "plugin_published", PluginID: pluginID}:
	default:
	}
}

func (b *Builder) NotifyPluginOffline(pluginID string) {
	select {
	case b.updateCh <- IndexEvent{Type: "plugin_offline", PluginID: pluginID}:
	default:
	}
}

func (b *Builder) Stop() {
	close(b.updateCh)
}

func (b *Builder) ToPluginIndexItem(p *model.Plugin) *PluginIndexItem {
	return &PluginIndexItem{
		ID:            p.ID,
		Name:          p.Name,
		Version:       p.Version,
		Author:        p.Author,
		Category:      p.Category,
		Keywords:      p.Keywords,
		Description:   p.Description,
		DownloadUrl:   p.DownloadUrl,
		Checksum:      p.Checksum,
		Size:          p.Size,
		DownloadCount: p.DownloadCount,
	}
}
