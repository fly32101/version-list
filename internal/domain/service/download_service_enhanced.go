package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// EnhancedDownloadServiceImpl 增强的下载服务实现
type EnhancedDownloadServiceImpl struct {
	*DownloadServiceImpl
	retryStrategy *RetryStrategy
	stats         *DownloadStats
}

// NewEnhancedDownloadService 创建增强的下载服务
func NewEnhancedDownloadService(options *DownloadOptions) EnhancedDownloadService {
	baseService := NewDownloadService(options).(*DownloadServiceImpl)

	return &EnhancedDownloadServiceImpl{
		DownloadServiceImpl: baseService,
		retryStrategy:       DefaultRetryStrategy(),
	}
}

// SetRetryStrategy 设置重试策略
func (e *EnhancedDownloadServiceImpl) SetRetryStrategy(strategy *RetryStrategy) {
	e.retryStrategy = strategy
}

// DownloadWithStats 带统计信息的下载
func (e *EnhancedDownloadServiceImpl) DownloadWithStats(url, destPath string, progress ProgressCallback) (*DownloadStats, error) {
	stats := &DownloadStats{
		URL:       url,
		FilePath:  destPath,
		StartTime: time.Now(),
		MinSpeed:  float64(^uint64(0) >> 1), // 最大值
	}
	e.stats = stats

	// 获取文件大小
	if size, err := e.GetFileSize(url); err == nil {
		stats.FileSize = size
	}

	// 检查是否支持断点续传
	if supportsResume, err := e.SupportsResume(url); err == nil {
		stats.SupportsResume = supportsResume
	}

	// 创建增强的进度回调
	enhancedProgress := func(downloaded, total int64, speed float64) {
		// 更新统计信息
		stats.Downloaded = downloaded
		if speed > stats.MaxSpeed {
			stats.MaxSpeed = speed
		}
		if speed < stats.MinSpeed && speed > 0 {
			stats.MinSpeed = speed
		}

		// 调用原始进度回调
		if progress != nil {
			progress(downloaded, total, speed)
		}
	}

	// 执行下载
	err := e.downloadWithRetry(url, destPath, enhancedProgress)

	// 完成统计
	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)
	if stats.Duration.Seconds() > 0 {
		stats.AverageSpeed = float64(stats.Downloaded) / stats.Duration.Seconds()
	}

	return stats, err
}

// downloadWithRetry 带重试的下载
func (e *EnhancedDownloadServiceImpl) downloadWithRetry(url, destPath string, progress ProgressCallback) error {
	var lastErr error

	for attempt := 0; attempt <= e.retryStrategy.MaxRetries; attempt++ {
		if attempt > 0 {
			// 检查错误是否可重试
			if !e.retryStrategy.IsRetryableError(lastErr) {
				break
			}

			// 计算延迟时间
			delay := e.retryStrategy.CalculateDelay(attempt)
			time.Sleep(delay)

			if e.stats != nil {
				e.stats.RetryCount++
			}
		}

		err := e.DownloadServiceImpl.Download(url, destPath, progress)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("下载失败，已重试 %d 次: %v", e.retryStrategy.MaxRetries, lastErr)
}

// GetDownloadStats 获取下载统计信息
func (e *EnhancedDownloadServiceImpl) GetDownloadStats() *DownloadStats {
	return e.stats
}

// DownloadManager 下载管理器
type DownloadManager struct {
	service         EnhancedDownloadService
	activeDownloads map[string]*DownloadContext
}

// DownloadContext 下载上下文
type DownloadContext struct {
	URL       string
	DestPath  string
	Progress  *DownloadProgress
	Stats     *DownloadStats
	Cancel    context.CancelFunc
	Done      chan error
	StartTime time.Time
}

// NewDownloadManager 创建下载管理器
func NewDownloadManager(options *DownloadOptions) *DownloadManager {
	return &DownloadManager{
		service:         NewEnhancedDownloadService(options),
		activeDownloads: make(map[string]*DownloadContext),
	}
}

// StartDownload 开始下载
func (dm *DownloadManager) StartDownload(url, destPath string, progress ProgressCallback) (*DownloadContext, error) {
	// 检查是否已经在下载
	if _, exists := dm.activeDownloads[url]; exists {
		return nil, fmt.Errorf("URL %s 正在下载中", url)
	}

	_, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	downloadCtx := &DownloadContext{
		URL:       url,
		DestPath:  destPath,
		Cancel:    cancel,
		Done:      done,
		StartTime: time.Now(),
	}

	dm.activeDownloads[url] = downloadCtx

	// 启动下载协程
	go func() {
		defer func() {
			delete(dm.activeDownloads, url)
			close(done)
		}()

		// 创建带进度更新的回调
		progressCallback := func(downloaded, total int64, speed float64) {
			downloadCtx.Progress = CalculateProgress(downloaded, total, downloadCtx.StartTime)
			if progress != nil {
				progress(downloaded, total, speed)
			}
		}

		stats, err := dm.service.DownloadWithStats(url, destPath, progressCallback)
		downloadCtx.Stats = stats
		done <- err
	}()

	return downloadCtx, nil
}

// CancelDownload 取消下载
func (dm *DownloadManager) CancelDownload(url string) error {
	downloadCtx, exists := dm.activeDownloads[url]
	if !exists {
		return fmt.Errorf("URL %s 没有在下载中", url)
	}

	downloadCtx.Cancel()
	return nil
}

// GetActiveDownloads 获取活跃的下载
func (dm *DownloadManager) GetActiveDownloads() map[string]*DownloadContext {
	result := make(map[string]*DownloadContext)
	for url, ctx := range dm.activeDownloads {
		result[url] = ctx
	}
	return result
}

// WaitForDownload 等待下载完成
func (dm *DownloadManager) WaitForDownload(url string) error {
	downloadCtx, exists := dm.activeDownloads[url]
	if !exists {
		return fmt.Errorf("URL %s 没有在下载中", url)
	}

	return <-downloadCtx.Done
}

// NetworkErrorClassifier 网络错误分类器
type NetworkErrorClassifier struct{}

// ClassifyNetworkError 分类网络错误
func (nec *NetworkErrorClassifier) ClassifyNetworkError(err error) *InstallError {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// 超时错误
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") {
		return NewInstallError(ErrorTypeTimeout, "网络请求超时", err)
	}

	// 连接错误
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection aborted") {
		return NewInstallError(ErrorTypeNetwork, "网络连接失败", err)
	}

	// DNS错误
	if strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "dns") {
		return NewInstallError(ErrorTypeNetwork, "DNS解析失败", err)
	}

	// 网络不可达
	if strings.Contains(errStr, "network unreachable") ||
		strings.Contains(errStr, "host unreachable") {
		return NewInstallError(ErrorTypeNetwork, "网络不可达", err)
	}

	// TLS/SSL错误
	if strings.Contains(errStr, "tls") ||
		strings.Contains(errStr, "ssl") ||
		strings.Contains(errStr, "certificate") {
		return NewInstallError(ErrorTypeNetwork, "TLS/SSL连接错误", err)
	}

	// 默认网络错误
	return NewInstallError(ErrorTypeNetwork, "网络错误", err)
}

// DownloadValidator 下载验证器
type DownloadValidator struct {
	fileValidator FileValidator
}

// NewDownloadValidator 创建下载验证器
func NewDownloadValidator() *DownloadValidator {
	return &DownloadValidator{
		fileValidator: NewFileValidator(),
	}
}

// ValidateDownload 验证下载
func (dv *DownloadValidator) ValidateDownload(filePath string, expectedSize int64, checksum string) error {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewInstallError(ErrorTypeFileSystem, "下载文件不存在", nil)
		}
		return NewInstallError(ErrorTypeFileSystem, "无法获取文件信息", err)
	}

	// 验证文件大小
	if expectedSize > 0 && fileInfo.Size() != expectedSize {
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("文件大小不匹配: 期望 %d, 实际 %d", expectedSize, fileInfo.Size()), nil)
	}

	// 验证校验和
	if checksum != "" {
		if err := dv.fileValidator.ValidateChecksum(filePath, checksum, "sha256"); err != nil {
			return NewInstallError(ErrorTypeValidation, "文件校验和验证失败", err)
		}
	}

	return nil
}
