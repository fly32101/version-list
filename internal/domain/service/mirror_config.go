package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MirrorConfig 镜像配置管理
type MirrorConfig struct {
	CustomMirrors []Mirror                   `json:"custom_mirrors"`
	TestResults   map[string]MirrorTestCache `json:"test_results"`
	LastUpdated   time.Time                  `json:"last_updated"`
	mu            sync.RWMutex
}

// MirrorTestCache 镜像测试结果缓存
type MirrorTestCache struct {
	Available    bool          `json:"available"`
	ResponseTime time.Duration `json:"response_time"`
	TestedAt     time.Time     `json:"tested_at"`
	Error        string        `json:"error,omitempty"`
}

// NewMirrorConfig 创建镜像配置实例
func NewMirrorConfig() *MirrorConfig {
	return &MirrorConfig{
		CustomMirrors: make([]Mirror, 0),
		TestResults:   make(map[string]MirrorTestCache),
		LastUpdated:   time.Now(),
	}
}

// LoadFromFile 从文件加载配置
func (c *MirrorConfig) LoadFromFile(configPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，使用默认配置
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// SaveToFile 保存配置到文件
func (c *MirrorConfig) SaveToFile(configPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	c.LastUpdated = time.Now()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// AddCustomMirror 添加自定义镜像
func (c *MirrorConfig) AddCustomMirror(mirror Mirror) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已存在
	for i, existing := range c.CustomMirrors {
		if existing.Name == mirror.Name {
			c.CustomMirrors[i] = mirror
			return
		}
	}

	c.CustomMirrors = append(c.CustomMirrors, mirror)
}

// RemoveCustomMirror 移除自定义镜像
func (c *MirrorConfig) RemoveCustomMirror(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, mirror := range c.CustomMirrors {
		if mirror.Name == name {
			c.CustomMirrors = append(c.CustomMirrors[:i], c.CustomMirrors[i+1:]...)
			return true
		}
	}

	return false
}

// GetCustomMirrors 获取自定义镜像列表
func (c *MirrorConfig) GetCustomMirrors() []Mirror {
	c.mu.RLock()
	defer c.mu.RUnlock()

	mirrors := make([]Mirror, len(c.CustomMirrors))
	copy(mirrors, c.CustomMirrors)
	return mirrors
}

// CacheTestResult 缓存测试结果
func (c *MirrorConfig) CacheTestResult(mirrorName string, result *MirrorTestResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cache := MirrorTestCache{
		Available:    result.Available,
		ResponseTime: result.ResponseTime,
		TestedAt:     time.Now(),
	}

	if result.Error != nil {
		cache.Error = result.Error.Error()
	}

	c.TestResults[mirrorName] = cache
}

// GetCachedResult 获取缓存的测试结果
func (c *MirrorConfig) GetCachedResult(mirrorName string, maxAge time.Duration) (*MirrorTestCache, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cache, exists := c.TestResults[mirrorName]
	if !exists {
		return nil, false
	}

	// 检查缓存是否过期
	if time.Since(cache.TestedAt) > maxAge {
		return nil, false
	}

	return &cache, true
}

// ClearExpiredCache 清理过期缓存
func (c *MirrorConfig) ClearExpiredCache(maxAge time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for name, cache := range c.TestResults {
		if now.Sub(cache.TestedAt) > maxAge {
			delete(c.TestResults, name)
		}
	}
}
