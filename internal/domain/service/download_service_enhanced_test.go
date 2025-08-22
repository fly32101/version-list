package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEnhancedDownloadService_DownloadWithStats(t *testing.T) {
	testContent := "Hello, World! This is test content for enhanced download service."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "enhanced_test.txt")

	service := NewEnhancedDownloadService(&DownloadOptions{
		MaxRetries: 1,
		RetryDelay: 100 * time.Millisecond,
		Timeout:    5 * time.Second,
		ChunkSize:  10, // 小块大小以测试进度
	})

	var progressCalls int
	progressCallback := func(downloaded, total int64, speed float64) {
		progressCalls++
		t.Logf("进度: %d/%d 字节, 速度: %.2f B/s", downloaded, total, speed)
	}

	stats, err := service.DownloadWithStats(server.URL, destPath, progressCallback)
	if err != nil {
		t.Fatalf("DownloadWithStats() 返回错误: %v", err)
	}

	// 验证统计信息
	if stats == nil {
		t.Fatal("统计信息不能为空")
	}

	if stats.URL != server.URL {
		t.Errorf("URL不匹配: 得到 %s, 期望 %s", stats.URL, server.URL)
	}

	if stats.FilePath != destPath {
		t.Errorf("文件路径不匹配: 得到 %s, 期望 %s", stats.FilePath, destPath)
	}

	if stats.FileSize != int64(len(testContent)) {
		t.Errorf("文件大小不匹配: 得到 %d, 期望 %d", stats.FileSize, len(testContent))
	}

	if stats.Downloaded != int64(len(testContent)) {
		t.Errorf("下载字节数不匹配: 得到 %d, 期望 %d", stats.Downloaded, len(testContent))
	}

	if !stats.SupportsResume {
		t.Error("应该支持断点续传")
	}

	if stats.AverageSpeed <= 0 {
		t.Error("平均速度应该大于0")
	}

	if stats.Duration <= 0 {
		t.Error("下载时间应该大于0")
	}

	t.Logf("下载统计:\n%s", stats.String())

	// 验证文件内容
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("读取下载文件失败: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("下载内容不匹配")
	}

	if progressCalls == 0 {
		t.Error("进度回调未被调用")
	}
}

func TestRetryStrategy_CalculateDelay(t *testing.T) {
	strategy := &RetryStrategy{
		InitialDelay:  1 * time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
	}

	tests := []struct {
		attempt     int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{0, 0, 2 * time.Second},
		{1, 2 * time.Second, 4 * time.Second},
		{2, 4 * time.Second, 8 * time.Second},
		{3, 8 * time.Second, 10 * time.Second}, // 应该被限制在MaxDelay
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			delay := strategy.CalculateDelay(tt.attempt)
			if delay < tt.expectedMin || delay > tt.expectedMax {
				t.Errorf("CalculateDelay(%d) = %v, 期望在 %v 到 %v 之间",
					tt.attempt, delay, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestRetryStrategy_IsRetryableError(t *testing.T) {
	strategy := DefaultRetryStrategy()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "connection reset error",
			err:      fmt.Errorf("connection reset by peer"),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      fmt.Errorf("request timeout"),
			expected: true,
		},
		{
			name:     "network unreachable",
			err:      fmt.Errorf("network unreachable"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      fmt.Errorf("file not found"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

func TestDownloadProgress_CalculateProgress(t *testing.T) {
	startTime := time.Now().Add(-10 * time.Second) // 10秒前开始
	downloaded := int64(500)
	total := int64(1000)

	progress := CalculateProgress(downloaded, total, startTime)

	if progress.Downloaded != downloaded {
		t.Errorf("Downloaded = %d, 期望 %d", progress.Downloaded, downloaded)
	}

	if progress.Total != total {
		t.Errorf("Total = %d, 期望 %d", progress.Total, total)
	}

	if progress.Percentage != 50.0 {
		t.Errorf("Percentage = %.1f, 期望 50.0", progress.Percentage)
	}

	if progress.Speed <= 0 {
		t.Error("Speed 应该大于 0")
	}

	if progress.ETA <= 0 {
		t.Error("ETA 应该大于 0")
	}

	// 测试字符串表示
	progressStr := progress.String()
	if !strings.Contains(progressStr, "50.0%") {
		t.Errorf("进度字符串应该包含百分比: %s", progressStr)
	}
}

func TestDownloadManager_StartDownload(t *testing.T) {
	testContent := "Test content for download manager"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "manager_test.txt")

	manager := NewDownloadManager(&DownloadOptions{
		Timeout:   5 * time.Second,
		ChunkSize: 1024,
	})

	// 开始下载
	downloadCtx, err := manager.StartDownload(server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("StartDownload() 返回错误: %v", err)
	}

	if downloadCtx == nil {
		t.Fatal("下载上下文不能为空")
	}

	if downloadCtx.URL != server.URL {
		t.Errorf("URL不匹配: 得到 %s, 期望 %s", downloadCtx.URL, server.URL)
	}

	// 检查活跃下载
	activeDownloads := manager.GetActiveDownloads()
	if len(activeDownloads) != 1 {
		t.Errorf("活跃下载数量 = %d, 期望 1", len(activeDownloads))
	}

	// 等待下载完成
	err = manager.WaitForDownload(server.URL)
	if err != nil {
		t.Errorf("WaitForDownload() 返回错误: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("读取下载文件失败: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("下载内容不匹配")
	}

	// 检查统计信息
	if downloadCtx.Stats == nil {
		t.Error("统计信息不能为空")
	}
}

func TestDownloadManager_CancelDownload(t *testing.T) {
	// 创建一个慢速服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)

		// 慢速写入数据
		for i := 0; i < 10; i++ {
			w.Write([]byte("0123456789"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "cancel_test.txt")

	manager := NewDownloadManager(&DownloadOptions{
		Timeout:   10 * time.Second,
		ChunkSize: 10,
	})

	// 开始下载
	_, err := manager.StartDownload(server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("StartDownload() 返回错误: %v", err)
	}

	// 等待一小段时间让下载开始
	time.Sleep(50 * time.Millisecond)

	// 取消下载
	err = manager.CancelDownload(server.URL)
	if err != nil {
		t.Errorf("CancelDownload() 返回错误: %v", err)
	}

	// 等待下载结束（应该被取消）
	err = manager.WaitForDownload(server.URL)
	// 下载被取消，可能返回错误，这是正常的
	t.Logf("取消下载后的错误（正常）: %v", err)
}

func TestNetworkErrorClassifier_ClassifyNetworkError(t *testing.T) {
	classifier := &NetworkErrorClassifier{}

	tests := []struct {
		name         string
		err          error
		expectedType ErrorType
	}{
		{
			name:         "timeout error",
			err:          fmt.Errorf("request timeout"),
			expectedType: ErrorTypeTimeout,
		},
		{
			name:         "connection refused",
			err:          fmt.Errorf("connection refused"),
			expectedType: ErrorTypeNetwork,
		},
		{
			name:         "DNS error",
			err:          fmt.Errorf("no such host"),
			expectedType: ErrorTypeNetwork,
		},
		{
			name:         "TLS error",
			err:          fmt.Errorf("tls handshake failed"),
			expectedType: ErrorTypeNetwork,
		},
		{
			name:         "generic network error",
			err:          fmt.Errorf("network error"),
			expectedType: ErrorTypeNetwork,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installErr := classifier.ClassifyNetworkError(tt.err)
			if installErr == nil {
				t.Fatal("ClassifyNetworkError() 返回 nil")
			}

			if installErr.Type != tt.expectedType {
				t.Errorf("错误类型 = %v, 期望 %v", installErr.Type, tt.expectedType)
			}
		})
	}

	// 测试 nil 错误
	result := classifier.ClassifyNetworkError(nil)
	if result != nil {
		t.Error("ClassifyNetworkError(nil) 应该返回 nil")
	}
}

func TestDownloadValidator_ValidateDownload(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_validate.txt")
	testContent := "Hello, World!"

	// 创建测试文件
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	validator := NewDownloadValidator()

	// 测试正常验证
	err = validator.ValidateDownload(testFile, int64(len(testContent)), "")
	if err != nil {
		t.Errorf("ValidateDownload() 返回错误: %v", err)
	}

	// 测试文件大小不匹配
	err = validator.ValidateDownload(testFile, 999, "")
	if err == nil {
		t.Error("ValidateDownload() 应该返回大小不匹配错误")
	}

	// 测试文件不存在
	nonExistentFile := filepath.Join(tempDir, "non_existent.txt")
	err = validator.ValidateDownload(nonExistentFile, 0, "")
	if err == nil {
		t.Error("ValidateDownload() 应该返回文件不存在错误")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("bytes_%d", tt.bytes), func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %s, 期望 %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		contains string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m"},
		{3661 * time.Second, "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.duration.String(), func(t *testing.T) {
			result := formatDuration(tt.duration)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("formatDuration(%v) = %s, 应该包含 %s", tt.duration, result, tt.contains)
			}
		})
	}
}
