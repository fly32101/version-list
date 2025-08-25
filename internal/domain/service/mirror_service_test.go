package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMirrorService(t *testing.T) {
	service := NewMirrorService()
	assert.NotNil(t, service)

	mirrors := service.GetAvailableMirrors()
	assert.NotEmpty(t, mirrors)
	assert.Equal(t, "official", mirrors[0].Name) // 应该按优先级排序
}

func TestMirrorService_GetAvailableMirrors(t *testing.T) {
	service := NewMirrorService()
	mirrors := service.GetAvailableMirrors()

	// 验证返回的镜像数量
	assert.Len(t, mirrors, 5)

	// 验证按优先级排序
	for i := 1; i < len(mirrors); i++ {
		assert.GreaterOrEqual(t, mirrors[i].Priority, mirrors[i-1].Priority)
	}

	// 验证包含预期的镜像
	mirrorNames := make(map[string]bool)
	for _, mirror := range mirrors {
		mirrorNames[mirror.Name] = true
	}

	expectedMirrors := []string{"official", "goproxy-cn", "aliyun", "tencent", "huawei"}
	for _, name := range expectedMirrors {
		assert.True(t, mirrorNames[name], "应该包含镜像: %s", name)
	}
}

func TestMirrorService_GetMirrorByName(t *testing.T) {
	service := NewMirrorService()

	// 测试存在的镜像
	mirror, err := service.GetMirrorByName("official")
	require.NoError(t, err)
	assert.Equal(t, "official", mirror.Name)
	assert.Equal(t, "https://golang.org/dl/", mirror.BaseURL)

	// 测试不存在的镜像
	_, err = service.GetMirrorByName("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未找到名为 'nonexistent' 的镜像")
}

func TestMirrorService_TestMirrorSpeed(t *testing.T) {
	// 创建测试HTTP服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟延迟
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewMirrorService()
	mirror := Mirror{
		Name:        "test",
		BaseURL:     server.URL,
		Description: "测试镜像",
		Region:      "test",
		Priority:    1,
	}

	ctx := context.Background()
	result, err := service.TestMirrorSpeed(ctx, mirror)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Available)
	assert.Greater(t, result.ResponseTime, time.Duration(0))
	assert.Equal(t, mirror.Name, result.Mirror.Name)
}

func TestMirrorService_TestMirrorSpeed_Failure(t *testing.T) {
	service := NewMirrorService()
	mirror := Mirror{
		Name:        "test",
		BaseURL:     "http://nonexistent.example.com",
		Description: "测试镜像",
		Region:      "test",
		Priority:    1,
	}

	ctx := context.Background()
	result, err := service.TestMirrorSpeed(ctx, mirror)

	require.NoError(t, err) // 方法本身不应该返回错误
	assert.NotNil(t, result)
	assert.False(t, result.Available)
	assert.NotNil(t, result.Error)
}

func TestMirrorService_TestMirrorSpeed_Timeout(t *testing.T) {
	// 创建一个会超时的服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // 超过客户端超时时间
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewMirrorService()
	mirror := Mirror{
		Name:        "test",
		BaseURL:     server.URL,
		Description: "测试镜像",
		Region:      "test",
		Priority:    1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := service.TestMirrorSpeed(ctx, mirror)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Available)
	assert.NotNil(t, result.Error)
}
func TestMirrorService_ValidateMirror(t *testing.T) {
	// 创建可用的测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewMirrorService()

	// 测试可用镜像
	validMirror := Mirror{
		Name:        "test",
		BaseURL:     server.URL,
		Description: "测试镜像",
		Region:      "test",
		Priority:    1,
	}

	ctx := context.Background()
	err := service.ValidateMirror(ctx, validMirror)
	assert.NoError(t, err)

	// 测试不可用镜像
	invalidMirror := Mirror{
		Name:        "invalid_test",
		BaseURL:     "http://nonexistent.example.com",
		Description: "测试镜像",
		Region:      "test",
		Priority:    1,
	}

	err = service.ValidateMirror(ctx, invalidMirror)
	assert.Error(t, err)
}

func TestMirrorService_SelectFastestMirror(t *testing.T) {
	// 创建多个测试服务器，模拟不同的响应时间
	fastServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer fastServer.Close()

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	service := NewMirrorService()
	mirrors := []Mirror{
		{
			Name:        "slow",
			BaseURL:     slowServer.URL,
			Description: "慢镜像",
			Region:      "test",
			Priority:    2,
		},
		{
			Name:        "fast",
			BaseURL:     fastServer.URL,
			Description: "快镜像",
			Region:      "test",
			Priority:    1,
		},
	}

	ctx := context.Background()
	fastest, err := service.SelectFastestMirror(ctx, mirrors)

	require.NoError(t, err)
	assert.NotNil(t, fastest)
	assert.Equal(t, "fast", fastest.Name) // 应该选择更快的镜像
}

func TestMirrorService_SelectFastestMirror_NoAvailable(t *testing.T) {
	service := NewMirrorService()
	mirrors := []Mirror{
		{
			Name:        "unavailable",
			BaseURL:     "http://nonexistent.example.com",
			Description: "不可用镜像",
			Region:      "test",
			Priority:    1,
		},
	}

	ctx := context.Background()
	fastest, err := service.SelectFastestMirror(ctx, mirrors)

	assert.Error(t, err)
	assert.Nil(t, fastest)
	assert.Contains(t, err.Error(), "没有可用的镜像")
}

func TestMirrorService_SelectFastestMirror_EmptyList(t *testing.T) {
	service := NewMirrorService()
	mirrors := []Mirror{}

	ctx := context.Background()
	fastest, err := service.SelectFastestMirror(ctx, mirrors)

	assert.Error(t, err)
	assert.Nil(t, fastest)
	assert.Contains(t, err.Error(), "没有可用的镜像")
}

func TestDefaultMirrors(t *testing.T) {
	mirrors := getDefaultMirrors()

	assert.Len(t, mirrors, 5)

	// 验证官方镜像
	official := mirrors[0]
	assert.Equal(t, "official", official.Name)
	assert.Equal(t, "https://golang.org/dl/", official.BaseURL)
	assert.Equal(t, "global", official.Region)
	assert.Equal(t, 1, official.Priority)

	// 验证所有镜像都有必要的字段
	for _, mirror := range mirrors {
		assert.NotEmpty(t, mirror.Name)
		assert.NotEmpty(t, mirror.BaseURL)
		assert.NotEmpty(t, mirror.Description)
		assert.NotEmpty(t, mirror.Region)
		assert.Greater(t, mirror.Priority, 0)
	}
}
