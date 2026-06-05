// Multiple resolvers share a single Cache instance to avoid redundant
// network fetches for index.yaml files.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	Dir string
	TTL time.Duration
}

func New(dir string, ttl time.Duration) *Cache {
	return &Cache{
		Dir: dir,
		TTL: ttl,
	}
}

func (c *Cache) Get(key string) ([]byte, error) {
	path := c.path(key)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if time.Since(info.ModTime()) > c.TTL {
		return nil, fmt.Errorf("cache expired for key %q", key)
	}
	return os.ReadFile(path)
}

func (c *Cache) Set(key string, data []byte) error {
	path := c.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Cache) Age(key string) (time.Duration, bool) {
	path := c.path(key)
	info, err := os.Stat(path)
	if err != nil {
		return 0, false
	}
	return time.Since(info.ModTime()), true
}

// path converts a cache key to a filesystem path.
func (c *Cache) path(key string) string {
	h := sha256.Sum256([]byte(key))
	name := hex.EncodeToString(h[:]) + ".cache"
	return filepath.Join(c.Dir, name)
}
