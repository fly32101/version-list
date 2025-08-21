package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// ProgressCallback 进度回调函数
type ProgressCallback func(downloaded, total int64, speed float64)

// DownloadProgress 下载进度信息
type DownloadProgress struct {
	Downloaded int64   // 已下载字节数
	Total      int64   // 总字节数
	Speed      float64 // 下载速度 (bytes/sec)
	Percentage float64 // 完成百分比
	ETA        int64   // 预计剩余时间（秒）
}

// DownloadService 下载服务接口
type DownloadService interface {
	Download(url, destPath string, progress ProgressCallback) error
	DownloadWithContext(ctx context.Context, url, destPath string, progress ProgressCallback) error
	GetFileSize(url string) (int64, error)
	SupportsResume(url string) (bool, error)
}

// DownloadServiceImpl 下载服务实现
type DownloadServiceImpl struct {
	client     *http.Client
	maxRetries int
	retryDelay time.Duration
	chunkSize  int64
	userAgent  string
}

// DownloadOptions 下载选项
type DownloadOptions struct {
	MaxRetries int           // 最大重试次数
	RetryDelay time.Duration // 重试延迟
	Timeout    time.Duration // 超时时间
	ChunkSize  int64         // 分块大小
	UserAgent  string        // User-Agent
}

// NewDownloadService 创建下载服务实例
func NewDownloadService(options *DownloadOptions) DownloadService {
	if options == nil {
		options = &DownloadOptions{
			MaxRetries: 3,
			RetryDelay: 2 * time.Second,
			Timeout:    30 * time.Minute,
			ChunkSize:  8192, // 8KB
			UserAgent:  "go-version-manager/1.0",
		}
	}

	client := &http.Client{
		Timeout: options.Timeout,
	}

	return &DownloadServiceImpl{
		client:     client,
		maxRetries: options.MaxRetries,
		retryDelay: options.RetryDelay,
		chunkSize:  options.ChunkSize,
		userAgent:  options.UserAgent,
	}
}

// Download 下载文件
func (d *DownloadServiceImpl) Download(url, destPath string, progress ProgressCallback) error {
	return d.DownloadWithContext(context.Background(), url, destPath, progress)
}

// DownloadWithContext 带上下文的下载文件
func (d *DownloadServiceImpl) DownloadWithContext(ctx context.Context, url, destPath string, progress ProgressCallback) error {
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 检查是否支持断点续传
	var startByte int64 = 0
	if fileInfo, err := os.Stat(destPath); err == nil {
		// 文件已存在，检查是否支持断点续传
		supportsResume, err := d.SupportsResume(url)
		if err != nil {
			return fmt.Errorf("检查断点续传支持失败: %v", err)
		}

		if supportsResume {
			startByte = fileInfo.Size()
		} else {
			// 不支持断点续传，删除现有文件
			if err := os.Remove(destPath); err != nil {
				return fmt.Errorf("删除现有文件失败: %v", err)
			}
		}
	}

	var lastErr error
	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(d.retryDelay):
			}
		}

		err := d.downloadWithResume(ctx, url, destPath, startByte, progress)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < d.maxRetries {
			// 重试前检查文件大小，更新startByte
			if fileInfo, statErr := os.Stat(destPath); statErr == nil {
				startByte = fileInfo.Size()
			}
		}
	}

	return fmt.Errorf("下载失败，已重试 %d 次: %v", d.maxRetries, lastErr)
}

// downloadWithResume 支持断点续传的下载
func (d *DownloadServiceImpl) downloadWithResume(ctx context.Context, url, destPath string, startByte int64, progress ProgressCallback) error {
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", d.userAgent)

	// 设置Range头支持断点续传
	if startByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
	}

	// 发送请求
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if startByte > 0 && resp.StatusCode != http.StatusPartialContent {
		if resp.StatusCode == http.StatusOK {
			// 服务器不支持断点续传，重新开始下载
			startByte = 0
			if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("删除现有文件失败: %v", err)
			}
		} else {
			return fmt.Errorf("HTTP响应错误: %d %s", resp.StatusCode, resp.Status)
		}
	} else if startByte == 0 && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP响应错误: %d %s", resp.StatusCode, resp.Status)
	}

	// 获取文件总大小
	var totalSize int64
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			totalSize = startByte + size
		}
	}

	// 打开目标文件
	var file *os.File
	if startByte > 0 {
		file, err = os.OpenFile(destPath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		file, err = os.Create(destPath)
	}
	if err != nil {
		return fmt.Errorf("打开目标文件失败: %v", err)
	}
	defer file.Close()

	// 开始下载
	return d.copyWithProgress(ctx, file, resp.Body, startByte, totalSize, progress)
}

// copyWithProgress 带进度的文件复制
func (d *DownloadServiceImpl) copyWithProgress(ctx context.Context, dst io.Writer, src io.Reader, startByte, totalSize int64, progress ProgressCallback) error {
	buffer := make([]byte, d.chunkSize)
	downloaded := startByte
	startTime := time.Now()
	lastProgressTime := startTime

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := src.Read(buffer)
		if n > 0 {
			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("写入文件失败: %v", writeErr)
			}
			downloaded += int64(n)

			// 更新进度
			if progress != nil && time.Since(lastProgressTime) >= 100*time.Millisecond {
				elapsed := time.Since(startTime).Seconds()
				speed := float64(downloaded-startByte) / elapsed
				progress(downloaded, totalSize, speed)
				lastProgressTime = time.Now()
			}
		}

		if err == io.EOF {
			// 最后一次进度更新
			if progress != nil {
				elapsed := time.Since(startTime).Seconds()
				speed := float64(downloaded-startByte) / elapsed
				progress(downloaded, totalSize, speed)
			}
			break
		}

		if err != nil {
			return fmt.Errorf("读取数据失败: %v", err)
		}
	}

	return nil
}

// GetFileSize 获取远程文件大小
func (d *DownloadServiceImpl) GetFileSize(url string) (int64, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建HEAD请求失败: %v", err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HEAD请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HEAD请求响应错误: %d %s", resp.StatusCode, resp.Status)
	}

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("服务器未返回Content-Length")
	}

	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("解析Content-Length失败: %v", err)
	}

	return size, nil
}

// SupportsResume 检查是否支持断点续传
func (d *DownloadServiceImpl) SupportsResume(url string) (bool, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, fmt.Errorf("创建HEAD请求失败: %v", err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("HEAD请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HEAD请求响应错误: %d %s", resp.StatusCode, resp.Status)
	}

	// 检查是否支持Range请求
	acceptRanges := resp.Header.Get("Accept-Ranges")
	return acceptRanges == "bytes", nil
}
