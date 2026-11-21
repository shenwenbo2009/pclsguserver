package apidoc

import (
	"sync"
	"time"
)

type SpecCache struct {
	mu       sync.RWMutex
	spec     *OpenAPISpec
	markdown string
	builtAt  time.Time
	version  string
}

var defaultCache = &SpecCache{}

func GetDefaultCache() *SpecCache {
	return defaultCache
}

func (c *SpecCache) GetSpec() *OpenAPISpec {
	c.mu.RLock()
	if c.spec != nil {
		spec := c.spec
		c.mu.RUnlock()
		return spec
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.spec != nil {
		return c.spec
	}
	c.spec = BuildSpec(c.version)
	if c.spec != nil {
		c.builtAt = time.Now()
	}
	return c.spec
}

func (c *SpecCache) GetMarkdown() string {
	c.mu.RLock()
	if c.markdown != "" {
		md := c.markdown
		c.mu.RUnlock()
		return md
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.markdown != "" {
		return c.markdown
	}
	spec := c.GetSpec()
	if spec == nil {
		return ""
	}
	c.markdown = GenerateMarkdown(spec)
	return c.markdown
}

func (c *SpecCache) Rebuild() {
	c.mu.Lock()
	defer c.mu.Unlock()

	newSpec := BuildSpec(c.version)
	if newSpec != nil {
		c.spec = newSpec
		c.builtAt = time.Now()
	}

	newMd := GenerateMarkdown(newSpec)
	if newMd != "" {
		c.markdown = newMd
	}
}

func (c *SpecCache) BuiltAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.builtAt
}

func (c *SpecCache) SetVersion(v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.version = v
}
