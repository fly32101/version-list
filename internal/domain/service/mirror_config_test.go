package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMirrorConfig(t *testing.T) {
	config := NewMirrorConfig()
	assert.NotNil(t, config)
	assert.Empty(t, config.CustomMirrors)
	assert.Empty(t, config.TestResults)
	assert.False(t, config.LastUpdated.IsZero())
}

func TestMirrorConfig_AddCustomMirror(t *testing.T) {
	config := NewMirrorConfig()

	mirror := Mirror{
		Name:        "custom",
		BaseURL:     "https://custom.example.com/golang/",
		Description: "自定义镜像",
		Region:      "custom",
		Priority:    10,
	}

	config.AddCustomMirror(mirror)

	customMirrors := config.GetCustomMirrors()
	assert.Len(t, customMirrors, 1)
	assert.Equal(t, "custom", customMirrors[0].Name)
}

func TestMirrorConfig_AddCustomMirror_Update(t *testing.T) {
	config := NewMirrorConfig()

	// 添加镜像
	mirror1 := Mirror{
		Name:        "custom",
		BaseURL:     "https://custom1.example.com/golang/",
		Description: "自定义镜像1",
		Region:      "custom",
		Priority:    10,
	}
	config.AddCustomMirror(mirror1)

	// 更新同名镜像
	mirror2 := Mirror{
		Name:        "custom",
		BaseURL:     "https://custom2.example.com/golang/",
		Description: "自定义镜像2",
		Region:      "custom",
		Priority:    11,
	}
	config.AddCustomMirror(mirror2)

	customMirrors := config.GetCustomMirrors()
	assert.Len(t, customMirrors, 1)
	assert.Equal(t, "https://custom2.example.com/golang/", customMirrors[0].BaseURL)
	assert.Equal(t, 11, customMirrors[0].Priority)
}

func TestMirrorConfig_RemoveCustomMirror(t *testing.T) {
	config := NewMirrorConfig()

	mirror := Mirror{
		Name:        "custom",
		BaseURL:     "https://custom.example.com/golang/",
		Description: "自定义镜像",
		Region:      "custom",
		Priority:    10,
	}
	config.AddCustomMirror(mirror)

	// 移除存在的镜像
	removed := config.RemoveCustomMirror("custom")
	assert.True(t, removed)
	assert.Empty(t, config.GetCustomMirrors())

	// 移除不存在的镜像
	removed = config.RemoveCustomMirror("nonexistent")
	assert.False(t, removed)
}

func TestMirrorConfig_CacheTestResult(t *testing.T) {
	config := NewMirrorConfig()

	result := &MirrorTestResult{
		Mirror:       Mirror{Name: "test"},
		Available:    true,
		ResponseTime: 100 * time.Millisecond,
		Error:        nil,
	}

	config.CacheTestResult("test", result)

	cached, exists := config.GetCachedResult("test", 1*time.Minute)
	require.True(t, exists)
	assert.True(t, cached.Available)
	assert.Equal(t, 100*time.Millisecond, cached.ResponseTime)
	assert.Empty(t, cached.Error)
}

func TestMirrorConfig_CacheTestResult_WithError(t *testing.T) {
	config := NewMirrorConfig()

	result := &MirrorTestResult{
		Mirror:       Mirror{Name: "test"},
		Available:    false,
		ResponseTime: 0,
		Error:        assert.AnError,
	}

	config.CacheTestResult("test", result)

	cached, exists := config.GetCachedResult("test", 1*time.Minute)
	require.True(t, exists)
	assert.False(t, cached.Available)
	assert.NotEmpty(t, cached.Error)
}

func TestMirrorConfig_GetCachedResult_Expired(t *testing.T) {
	config := NewMirrorConfig()

	result := &MirrorTestResult{
		Mirror:       Mirror{Name: "test"},
		Available:    true,
		ResponseTime: 100 * time.Millisecond,
	}

	config.CacheTestResult("test", result)

	// 立即获取应该存在
	cached, exists := config.GetCachedResult("test", 1*time.Minute)
	assert.True(t, exists)
	assert.NotNil(t, cached)

	// 等待一小段时间，然后使用很短的过期时间应该不存在
	time.Sleep(2 * time.Millisecond)
	cached, exists = config.GetCachedResult("test", 1*time.Millisecond)
	assert.False(t, exists)
	assert.Nil(t, cached)
}

func TestMirrorConfig_ClearExpiredCache(t *testing.T) {
	config := NewMirrorConfig()

	// 添加一些测试结果
	result1 := &MirrorTestResult{Mirror: Mirror{Name: "test1"}, Available: true}
	result2 := &MirrorTestResult{Mirror: Mirror{Name: "test2"}, Available: true}

	config.CacheTestResult("test1", result1)
	config.CacheTestResult("test2", result2)

	// 验证都存在
	_, exists1 := config.GetCachedResult("test1", 1*time.Minute)
	_, exists2 := config.GetCachedResult("test2", 1*time.Minute)
	assert.True(t, exists1)
	assert.True(t, exists2)

	// 等待一小段时间，然后清理过期缓存
	time.Sleep(2 * time.Millisecond)
	config.ClearExpiredCache(1 * time.Millisecond)

	// 验证都被清理了
	_, exists1 = config.GetCachedResult("test1", 1*time.Minute)
	_, exists2 = config.GetCachedResult("test2", 1*time.Minute)
	assert.False(t, exists1)
	assert.False(t, exists2)
}
func TestMirrorConfig_SaveAndLoadFromFile(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "mirror_config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "mirror_config.json")

	// 创建配置并添加数据
	config1 := NewMirrorConfig()
	mirror := Mirror{
		Name:        "custom",
		BaseURL:     "https://custom.example.com/golang/",
		Description: "自定义镜像",
		Region:      "custom",
		Priority:    10,
	}
	config1.AddCustomMirror(mirror)

	result := &MirrorTestResult{
		Mirror:       Mirror{Name: "test"},
		Available:    true,
		ResponseTime: 100 * time.Millisecond,
	}
	config1.CacheTestResult("test", result)

	// 保存到文件
	err = config1.SaveToFile(configPath)
	require.NoError(t, err)

	// 验证文件存在
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// 从文件加载到新配置
	config2 := NewMirrorConfig()
	err = config2.LoadFromFile(configPath)
	require.NoError(t, err)

	// 验证数据一致
	customMirrors := config2.GetCustomMirrors()
	assert.Len(t, customMirrors, 1)
	assert.Equal(t, "custom", customMirrors[0].Name)

	cached, exists := config2.GetCachedResult("test", 1*time.Minute)
	assert.True(t, exists)
	assert.True(t, cached.Available)
}

func TestMirrorConfig_LoadFromFile_NotExists(t *testing.T) {
	config := NewMirrorConfig()

	// 加载不存在的文件应该成功（使用默认配置）
	err := config.LoadFromFile("/nonexistent/path/config.json")
	assert.NoError(t, err)
}

func TestMirrorConfig_LoadFromFile_InvalidJSON(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "invalid_config_*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// 写入无效JSON
	_, err = tempFile.WriteString("invalid json content")
	require.NoError(t, err)
	tempFile.Close()

	config := NewMirrorConfig()
	err = config.LoadFromFile(tempFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "解析配置文件失败")
}

func TestNewMirrorServiceWithConfig(t *testing.T) {
	// 创建临时配置文件
	tempDir, err := os.MkdirTemp("", "mirror_service_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// 创建配置
	config := NewMirrorConfig()
	mirror := Mirror{
		Name:        "custom",
		BaseURL:     "https://custom.example.com/golang/",
		Description: "自定义镜像",
		Region:      "custom",
		Priority:    10,
	}
	config.AddCustomMirror(mirror)
	err = config.SaveToFile(configPath)
	require.NoError(t, err)

	// 使用配置创建服务
	service, err := NewMirrorServiceWithConfig(configPath)
	require.NoError(t, err)
	assert.NotNil(t, service)

	// 验证自定义镜像被加载
	mirrors := service.GetAvailableMirrors()
	found := false
	for _, m := range mirrors {
		if m.Name == "custom" {
			found = true
			break
		}
	}
	assert.True(t, found, "应该包含自定义镜像")
}

func TestNewMirrorServiceWithConfig_InvalidPath(t *testing.T) {
	// 使用无效路径创建服务
	service, err := NewMirrorServiceWithConfig("/invalid/path/config.json")

	// 应该成功创建（使用默认配置）
	require.NoError(t, err)
	assert.NotNil(t, service)
}
