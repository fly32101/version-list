package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadService_Basic(t *testing.T) {
	testContent := "Hello, World! This is test content."

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	// 创建临时文件
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test_download.txt")

	service := NewDownloadService(&DownloadOptions{
		MaxRetries: 1,
		RetryDelay: 100 * time.Millisecond,
		Timeout:    5 * time.Second,
		ChunkSize:  1024,
	})

	// 执行下载
	err := service.Download(server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("Download() 返回错误: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("读取下载文件失败: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("下载内容不匹配: 得到 %q, 期望 %q", string(content), testContent)
	}
}

func TestDownloadService_GetFileSize(t *testing.T) {
	testContent := "Hello, World!"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			t.Errorf("期望 HEAD 请求, 得到 %s", r.Method)
		}

		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewDownloadService(nil)
	size, err := service.GetFileSize(server.URL)
	if err != nil {
		t.Fatalf("GetFileSize() 返回错误: %v", err)
	}

	expectedSize := int64(len(testContent))
	if size != expectedSize {
		t.Errorf("GetFileSize() = %d, 期望 %d", size, expectedSize)
	}
}

func TestDownloadService_SupportsResume(t *testing.T) {
	// 测试支持断点续传的服务器
	serverWithResume := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)
	}))
	defer serverWithResume.Close()

	service := NewDownloadService(nil)
	supports, err := service.SupportsResume(serverWithResume.URL)
	if err != nil {
		t.Fatalf("SupportsResume() 返回错误: %v", err)
	}
	if !supports {
		t.Error("SupportsResume() = false, 期望 true")
	}

	// 测试不支持断点续传的服务器
	serverWithoutResume := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer serverWithoutResume.Close()

	supports, err = service.SupportsResume(serverWithoutResume.URL)
	if err != nil {
		t.Fatalf("SupportsResume() 返回错误: %v", err)
	}
	if supports {
		t.Error("SupportsResume() = true, 期望 false")
	}
}

func TestDownloadService_WithProgress(t *testing.T) {
	testContent := "Hello, World! This is test content for progress testing."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "test_progress.txt")

	service := NewDownloadService(&DownloadOptions{
		ChunkSize: 10, // 小块大小以确保多次进度回调
	})

	var progressCalls int
	progressCallback := func(downloaded, total int64, speed float64) {
		progressCalls++
		if downloaded > total {
			t.Errorf("下载字节数 %d 超过总大小 %d", downloaded, total)
		}
		if speed < 0 {
			t.Errorf("下载速度不能为负数: %f", speed)
		}
	}

	err := service.Download(server.URL, destPath, progressCallback)
	if err != nil {
		t.Fatalf("Download() 返回错误: %v", err)
	}

	if progressCalls == 0 {
		t.Error("进度回调未被调用")
	}

	// 验证文件内容
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("读取下载文件失败: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("下载内容不匹配")
	}
}
