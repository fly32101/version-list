package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"version-list/internal/domain/service"
)

func main() {
	fmt.Println("=== Go Version Manager - 增强下载服务示例 ===\n")

	// 1. 创建增强的下载服务
	fmt.Println("1. 创建增强的下载服务:")
	downloadService := service.NewEnhancedDownloadService(&service.DownloadOptions{
		MaxRetries: 3,
		RetryDelay: 2 * time.Second,
		Timeout:    30 * time.Second,
		ChunkSize:  8192,
		UserAgent:  "go-version-manager-example/1.0",
	})
	fmt.Println("   ✅ 下载服务创建成功")
	fmt.Println()

	// 2. 设置自定义重试策略
	fmt.Println("2. 设置自定义重试策略:")
	retryStrategy := &service.RetryStrategy{
		MaxRetries:    5,
		InitialDelay:  1 * time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"connection reset",
			"timeout",
			"network unreachable",
		},
	}
	downloadService.SetRetryStrategy(retryStrategy)
	fmt.Printf("   最大重试次数: %d\n", retryStrategy.MaxRetries)
	fmt.Printf("   初始延迟: %v\n", retryStrategy.InitialDelay)
	fmt.Printf("   最大延迟: %v\n", retryStrategy.MaxDelay)
	fmt.Printf("   退避因子: %.1f\n", retryStrategy.BackoffFactor)
	fmt.Println()

	// 3. 测试下载进度计算
	fmt.Println("3. 下载进度计算示例:")
	startTime := time.Now().Add(-30 * time.Second) // 模拟30秒前开始
	progress := service.CalculateProgress(75*1024*1024, 100*1024*1024, startTime)
	fmt.Printf("   %s\n", progress.String())
	fmt.Printf("   已下载: %s\n", formatBytes(progress.Downloaded))
	fmt.Printf("   总大小: %s\n", formatBytes(progress.Total))
	fmt.Printf("   进度: %.1f%%\n", progress.Percentage)
	fmt.Printf("   速度: %s/s\n", formatBytes(int64(progress.Speed)))
	fmt.Printf("   预计剩余时间: %s\n", formatDuration(time.Duration(progress.ETA)*time.Second))
	fmt.Println()

	// 4. 创建下载管理器
	fmt.Println("4. 下载管理器示例:")
	_ = service.NewDownloadManager(&service.DownloadOptions{
		MaxRetries: 2,
		Timeout:    10 * time.Second,
		ChunkSize:  4096,
	})
	fmt.Println("   ✅ 下载管理器创建成功")
	fmt.Println()

	// 5. 测试小文件下载
	fmt.Println("5. 测试小文件下载:")
	tempDir := os.TempDir()
	testFile := filepath.Join(tempDir, "go-version-test-download.txt")
	defer os.Remove(testFile)

	// 使用一个小的测试文件
	testURL := "https://golang.org/robots.txt"
	fmt.Printf("   下载URL: %s\n", testURL)
	fmt.Printf("   目标文件: %s\n", testFile)

	// 创建进度回调
	var lastProgress float64
	progressCallback := func(downloaded, total int64, speed float64) {
		if total > 0 {
			currentProgress := float64(downloaded) / float64(total) * 100
			if currentProgress-lastProgress >= 10 || downloaded == total {
				fmt.Printf("   进度: %.1f%% (%s/%s) @ %s/s\n",
					currentProgress,
					formatBytes(downloaded),
					formatBytes(total),
					formatBytes(int64(speed)))
				lastProgress = currentProgress
			}
		}
	}

	// 执行下载
	stats, err := downloadService.DownloadWithStats(testURL, testFile, progressCallback)
	if err != nil {
		fmt.Printf("   ❌ 下载失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 下载成功")
		fmt.Printf("   下载统计:\n%s\n", indentText(stats.String(), "     "))
	}
	fmt.Println()

	// 6. 测试文件验证
	fmt.Println("6. 下载文件验证:")
	if err == nil {
		validator := service.NewDownloadValidator()

		// 获取文件信息
		fileInfo, statErr := os.Stat(testFile)
		if statErr != nil {
			fmt.Printf("   ❌ 获取文件信息失败: %v\n", statErr)
		} else {
			fmt.Printf("   文件大小: %s\n", formatBytes(fileInfo.Size()))

			// 验证下载
			validateErr := validator.ValidateDownload(testFile, fileInfo.Size(), "")
			if validateErr != nil {
				fmt.Printf("   ❌ 文件验证失败: %v\n", validateErr)
			} else {
				fmt.Println("   ✅ 文件验证通过")
			}
		}
	}
	fmt.Println()

	// 7. 网络错误分类示例
	fmt.Println("7. 网络错误分类示例:")
	classifier := &service.NetworkErrorClassifier{}

	testErrors := []error{
		fmt.Errorf("connection timeout"),
		fmt.Errorf("connection refused"),
		fmt.Errorf("no such host: example.invalid"),
		fmt.Errorf("tls handshake failed"),
	}

	for _, testErr := range testErrors {
		installErr := classifier.ClassifyNetworkError(testErr)
		if installErr != nil {
			fmt.Printf("   错误: %s\n", testErr.Error())
			fmt.Printf("   分类: %s\n", installErr.Type.String())
			fmt.Printf("   建议: %s\n", installErr.GetSuggestion())
			fmt.Println()
		}
	}

	// 8. 重试策略测试
	fmt.Println("8. 重试策略测试:")
	for attempt := 0; attempt < 4; attempt++ {
		delay := retryStrategy.CalculateDelay(attempt)
		fmt.Printf("   第 %d 次重试延迟: %v\n", attempt+1, delay)
	}
	fmt.Println()

	fmt.Println("=== 增强下载服务示例完成 ===")
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
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) - minutes*60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) - hours*60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

// indentText 为文本添加缩进
func indentText(text, indent string) string {
	lines := []string{}
	current := ""
	for _, r := range text {
		if r == '\n' {
			lines = append(lines, indent+current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, indent+current)
	}

	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
