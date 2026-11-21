package index

import (
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

type PluginIndexItem struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	Author        string   `json:"author"`
	Category      string   `json:"category"`
	Keywords      []string `json:"keywords"`
	Description   string   `json:"description"`
	DownloadUrl   string   `json:"downloadUrl"`
	Checksum      string   `json:"checksum"`
	Size          int64    `json:"size"`
	DownloadCount int64    `json:"downloadCount"`
}

type SoftwareIndexItem struct {
	SwCode        string `json:"swCode"`
	LatestVersion string `json:"latestVersion"`
	SwUrl         string `json:"swUrl"`
	Checksum      string `json:"checksum"`
}

type IndexJson struct {
	Version         int                 `json:"version"`
	GeneratedAt     time.Time           `json:"generatedAt"`
	Total           int                 `json:"total"`
	Plugins         []PluginIndexItem   `json:"plugins"`
	SoftwareUpdates []SoftwareIndexItem `json:"softwareUpdates"`
}

type MemoryIndex struct {
	byName     map[string][]*PluginIndexItem
	byKeyword  map[string][]*PluginIndexItem
	byCategory map[string][]*PluginIndexItem
	all        []*PluginIndexItem
	version    int
	builtAt    time.Time
}

func BuildMemoryIndex(plugins []*PluginIndexItem) *MemoryIndex {
	mi := &MemoryIndex{
		byName:     make(map[string][]*PluginIndexItem),
		byKeyword:  make(map[string][]*PluginIndexItem),
		byCategory: make(map[string][]*PluginIndexItem),
		all:        plugins,
		builtAt:    time.Now(),
	}
	for _, p := range plugins {
		mi.byName[strings.ToLower(p.Name)] = append(mi.byName[strings.ToLower(p.Name)], p)
		mi.byCategory[p.Category] = append(mi.byCategory[p.Category], p)
		for _, kw := range p.Keywords {
			lk := strings.ToLower(kw)
			mi.byKeyword[lk] = append(mi.byKeyword[lk], p)
		}
	}
	return mi
}

type SearchResult struct {
	Items []*PluginIndexItem
	Total int
}

func (mi *MemoryIndex) Query(keyword, category, name string, page, size int) SearchResult {
	var candidates []*PluginIndexItem

	hasKeyword := keyword != ""
	hasCategory := category != ""
	hasName := name != ""

	if !hasKeyword && !hasCategory && !hasName {
		candidates = mi.all
	} else if hasName && !hasKeyword && !hasCategory {
		candidates = mi.byName[strings.ToLower(name)]
	} else if !hasName && !hasKeyword && hasCategory {
		candidates = mi.byCategory[category]
	} else if !hasName && hasKeyword && !hasCategory {
		candidates = mi.byKeyword[strings.ToLower(keyword)]
	} else {
		sets := [][]*PluginIndexItem{}
		if hasName {
			sets = append(sets, mi.byName[strings.ToLower(name)])
		}
		if hasCategory {
			sets = append(sets, mi.byCategory[category])
		}
		if hasKeyword {
			sets = append(sets, mi.byKeyword[strings.ToLower(keyword)])
		}
		candidates = intersect(sets)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].DownloadCount != candidates[j].DownloadCount {
			return candidates[i].DownloadCount > candidates[j].DownloadCount
		}
		return candidates[i].Name < candidates[j].Name
	})

	total := len(candidates)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	start := (page - 1) * size
	if start >= total {
		return SearchResult{Items: nil, Total: total}
	}
	end := start + size
	if end > total {
		end = total
	}
	return SearchResult{Items: candidates[start:end], Total: total}
}

func intersect(sets [][]*PluginIndexItem) []*PluginIndexItem {
	if len(sets) == 0 {
		return nil
	}
	if len(sets) == 1 {
		return sets[0]
	}
	seen := make(map[string]*PluginIndexItem)
	for _, p := range sets[0] {
		seen[p.ID] = p
	}
	for _, s := range sets[1:] {
		next := make(map[string]*PluginIndexItem)
		for _, p := range s {
			if _, ok := seen[p.ID]; ok {
				next[p.ID] = p
			}
		}
		seen = next
	}
	result := make([]*PluginIndexItem, 0, len(seen))
	for _, p := range seen {
		result = append(result, p)
	}
	return result
}

type Engine struct {
	current atomic.Pointer[MemoryIndex]
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Load() *MemoryIndex {
	return e.current.Load()
}

func (e *Engine) Store(mi *MemoryIndex) {
	e.current.Store(mi)
}

func (e *Engine) IsReady() bool {
	return e.current.Load() != nil
}

func (e *Engine) Search(keyword, category, name string, page, size int) (*SearchResult, bool) {
	mi := e.current.Load()
	if mi == nil {
		return nil, false
	}
	r := mi.Query(keyword, category, name, page, size)
	return &r, true
}
