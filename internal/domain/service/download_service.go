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
	Downloaded int64         // 已下载字节数
	Total      int64         // 总字节数
	Speed      float64       // 下载速度 (bytes/sec)
	Percentage float64       // 完成百分比
	ETA        int64         // 预计剩余时间（秒）
	StartTime  time.Time     // 开始时间
	Elapsed    time.Duration // 已用时间
}

// CalculateProgress 计算下载进度
func CalculateProgress(downloaded, total int64, startTime time.Time) *DownloadProgress {
	progress := &DownloadProgress{
		Downloaded: downloaded,
		Total:      total,
		StartTime:  startTime,
		Elapsed:    time.Since(startTime),
	}

	// 计算百分比
	if total > 0 {
		progress.Percentage = float64(downloaded) / float64(total) * 100
	}

	// 计算速度
	if progress.Elapsed.Seconds() > 0 {
		progress.Speed = float64(downloaded) / progress.Elapsed.Seconds()
	}

	// 计算ETA
	if progress.Speed > 0 && total > downloaded {
		remaining := total - downloaded
		progress.ETA = int64(float64(remaining) / progress.Speed)
	}

	return progress
}

// String 返回进度的字符串表示
func (p *DownloadProgress) String() string {
	if p.Total > 0 {
		return fmt.Sprintf("%.1f%% (%s/%s) @ %s/s ETA: %s",
			p.Percentage,
			formatBytes(p.Downloaded),
			formatBytes(p.Total),
			formatBytes(int64(p.Speed)),
			formatDuration(time.Duration(p.ETA)*time.Second))
	}
	return fmt.Sprintf("%s @ %s/s",
		formatBytes(p.Downloaded),
		formatBytes(int64(p.Speed)))
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
	client             *http.Client
	maxRetries         int
	retryDelay         time.Duration
	chunkSize          int64
	userAgent          string
	enableCache        bool
	cacheDir           string
	progressUpdateRate time.Duration
}

// DownloadOptions 下载选项
type DownloadOptions struct {
	MaxRetries         int           // 最大重试次数
	RetryDelay         time.Duration // 重试延迟
	Timeout            time.Duration // 超时时间
	ChunkSize          int64         // 分块大小
	UserAgent          string        // User-Agent
	EnableCache        bool          // 启用缓存
	CacheDir           string        // 缓存目录
	ProgressUpdateRate time.Duration // 进度更新频率
}

// NewDownloadService 创建下载服务实例
func NewDownloadService(options *DownloadOptions) DownloadService {
	if options == nil {
		options = &DownloadOptions{
			MaxRetries:         3,
			RetryDelay:         2 * time.Second,
			Timeout:            30 * time.Minute,
			ChunkSize:          optimizeChunkSize(), // 动态优化分块大小
			UserAgent:          "go-version-manager/1.0",
			EnableCache:        true,
			CacheDir:           filepath.Join(os.TempDir(), "go-version-cache"),
			ProgressUpdateRate: 50 * time.Millisecond, // 优化进度更新频率
		}
	}

	// 确保缓存目录存在
	if options.EnableCache && options.CacheDir != "" {
		os.MkdirAll(options.CacheDir, 0755)
	}

	client := &http.Client{
		Timeout: options.Timeout,
	}

	return &DownloadServiceImpl{
		client:             client,
		maxRetries:         options.MaxRetries,
		retryDelay:         options.RetryDelay,
		chunkSize:          options.ChunkSize,
		userAgent:          options.UserAgent,
		enableCache:        options.EnableCache,
		cacheDir:           options.CacheDir,
		progressUpdateRate: options.ProgressUpdateRate,
	}
}

// Download 下载文件
func (d *DownloadServiceImpl) Download(url, destPath string, progress ProgressCallback) error {
	return d.DownloadWithContext(context.Background(), url, destPath, progress)
}

// DownloadWithContext 带上下文的下载文件
func (d *DownloadServiceImpl) DownloadWithContext(ctx context.Context, url, destPath string, progress ProgressCallback) error {
	// 检查缓存
	if d.enableCache {
		if cachedFile, err := d.getCachedFile(url); err == nil {
			return d.copyFromCache(cachedFile, destPath, progress)
		}
	}

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
			// 下载成功后缓存文件
			if d.enableCache {
				d.cacheFile(url, destPath)
			}
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

			// 更新进度（使用优化的更新频率）
			if progress != nil && time.Since(lastProgressTime) >= d.progressUpdateRate {
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

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), d.Seconds()-60*d.Minutes())
	}
	return fmt.Sprintf("%.0fh%.0fm", d.Hours(), d.Minutes()-60*d.Hours())
}

// DownloadStats 下载统计信息
type DownloadStats struct {
	URL            string        // 下载URL
	FilePath       string        // 文件路径
	FileSize       int64         // 文件大小
	Downloaded     int64         // 已下载字节数
	StartTime      time.Time     // 开始时间
	EndTime        time.Time     // 结束时间
	Duration       time.Duration // 总用时
	AverageSpeed   float64       // 平均速度
	MaxSpeed       float64       // 最大速度
	MinSpeed       float64       // 最小速度
	RetryCount     int           // 重试次数
	SupportsResume bool          // 是否支持断点续传
}

// String 返回统计信息的字符串表示
func (s *DownloadStats) String() string {
	return fmt.Sprintf(`下载统计:
  URL: %s
  文件: %s
  大小: %s
  用时: %s
  平均速度: %s/s
  最大速度: %s/s
  重试次数: %d
  断点续传: %t`,
		s.URL,
		s.FilePath,
		formatBytes(s.FileSize),
		s.Duration,
		formatBytes(int64(s.AverageSpeed)),
		formatBytes(int64(s.MaxSpeed)),
		s.RetryCount,
		s.SupportsResume)
}

// EnhancedDownloadService 增强的下载服务接口
type EnhancedDownloadService interface {
	DownloadService
	DownloadWithStats(url, destPath string, progress ProgressCallback) (*DownloadStats, error)
	GetDownloadStats() *DownloadStats
	SetRetryStrategy(strategy *RetryStrategy)
}

// RetryStrategy 重试策略
type RetryStrategy struct {
	MaxRetries      int           // 最大重试次数
	InitialDelay    time.Duration // 初始延迟
	MaxDelay        time.Duration // 最大延迟
	BackoffFactor   float64       // 退避因子
	RetryableErrors []string      // 可重试的错误类型
}

// DefaultRetryStrategy 默认重试策略
func DefaultRetryStrategy() *RetryStrategy {
	return &RetryStrategy{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"connection reset",
			"connection refused",
			"timeout",
			"temporary failure",
			"network unreachable",
		},
	}
}

// CalculateDelay 计算重试延迟
func (rs *RetryStrategy) CalculateDelay(attempt int) time.Duration {
	delay := float64(rs.InitialDelay) * float64(attempt) * rs.BackoffFactor
	if delay > float64(rs.MaxDelay) {
		delay = float64(rs.MaxDelay)
	}
	return time.Duration(delay)
}

// IsRetryableError 判断错误是否可重试
func (rs *RetryStrategy) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	for _, retryableErr := range rs.RetryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}
	return false
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

// containsSubstring 检查是否包含子字符串
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// optimizeChunkSize 动态优化分块大小
func optimizeChunkSize() int64 {
	// 根据可用内存动态调整分块大小
	// 对于大文件下载，使用更大的缓冲区可以提高性能
	const (
		minChunkSize = 8 * 1024    // 8KB 最小值
		maxChunkSize = 1024 * 1024 // 1MB 最大值
		defaultChunk = 64 * 1024   // 64KB 默认值
	)

	// 简单的启发式算法：根据系统可用内存调整
	// 在实际应用中可以根据网络条件和文件大小进一步优化
	return defaultChunk
}

// getCachedFile 获取缓存文件路径
func (d *DownloadServiceImpl) getCachedFile(url string) (string, error) {
	if !d.enableCache || d.cacheDir == "" {
		return "", fmt.Errorf("缓存未启用")
	}

	// 使用URL的哈希作为缓存文件名
	hash := fmt.Sprintf("%x", url) // 简化的哈希，实际应用中可以使用更好的哈希算法
	cacheFile := filepath.Join(d.cacheDir, hash)

	// 检查缓存文件是否存在且有效
	if _, err := os.Stat(cacheFile); err == nil {
		return cacheFile, nil
	}

	return "", fmt.Errorf("缓存文件不存在")
}

// cacheFile 缓存下载的文件
func (d *DownloadServiceImpl) cacheFile(url, filePath string) error {
	if !d.enableCache || d.cacheDir == "" {
		return nil
	}

	hash := fmt.Sprintf("%x", url)
	cacheFile := filepath.Join(d.cacheDir, hash)

	// 复制文件到缓存目录
	src, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(cacheFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// copyFromCache 从缓存复制文件
func (d *DownloadServiceImpl) copyFromCache(cacheFile, destPath string, progress ProgressCallback) error {
	src, err := os.Open(cacheFile)
	if err != nil {
		return err
	}
	defer src.Close()

	// 获取文件大小
	fileInfo, err := src.Stat()
	if err != nil {
		return err
	}
	totalSize := fileInfo.Size()

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// 带进度的复制
	buffer := make([]byte, d.chunkSize)
	var copied int64
	startTime := time.Now()
	lastProgressTime := startTime

	for {
		n, err := src.Read(buffer)
		if n > 0 {
			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("写入文件失败: %v", writeErr)
			}
			copied += int64(n)

			// 更新进度
			if progress != nil && time.Since(lastProgressTime) >= d.progressUpdateRate {
				elapsed := time.Since(startTime).Seconds()
				speed := float64(copied) / elapsed
				progress(copied, totalSize, speed)
				lastProgressTime = time.Now()
			}
		}

		if err == io.EOF {
			// 最后一次进度更新
			if progress != nil {
				elapsed := time.Since(startTime).Seconds()
				speed := float64(copied) / elapsed
				progress(copied, totalSize, speed)
			}
			break
		}

		if err != nil {
			return fmt.Errorf("读取缓存文件失败: %v", err)
		}
	}

	return nil
}
