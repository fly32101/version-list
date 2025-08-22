package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"	"version-list/internal/domain/service"
	"version-list/internal/interface/ui""
)

func main() {
	fmt.Println("=== Go Version Manager 性能优化示例 ===")

	// 演示优化的下载服务
	demonstrateOptimizedDownload()

	// 演示并行解压
	demonstrateParallelExtraction()

	// 演示进度显示优化
	demonstrateOptimizedProgress()
}

// demonstrateOptimizedDownload 演示优化的下载服务
func demonstrateOptimizedDownload() {
	fmt.Println("\n1. 优化的下载服务演示")
	fmt.Println("特性：缓存机制、动态缓冲区、优化的进度更新频率")

	// 创建优化的下载选项
	options := &service.DownloadOptions{
		MaxRetries:         3,
		RetryDelay:         2 * time.Second,
		Timeout:            30 * time.Minute,
		ChunkSize:          64 * 1024, // 64KB 优化的分块大小
		UserAgent:          "go-version-manager/1.0",
		EnableCache:        true,
		CacheDir:           filepath.Join(os.TempDir(), "go-version-cache"),
		ProgressUpdateRate: 50 * time.Millisecond, // 优化的更新频率
	}

	downloadService := service.NewDownloadService(options)

	// 模拟下载（使用一个小文件进行演示）
	testURL := "https://httpbin.org/bytes/1024" // 1KB 测试文件
	destPath := filepath.Join(os.TempDir(), "test_download.bin")

	fmt.Printf("下载测试文件: %s\n", testURL)
	fmt.Printf("目标路径: %s\n", destPath)

	startTime := time.Now()

	// 进度回调
	progressCallback := func(downloaded, total int64, speed float64) {
		if total > 0 {
			percentage := float64(downloaded) / float64(total) * 100
			fmt.Printf("\r下载进度: %.1f%% (%s/%s) @ %s/s",
				percentage,
				formatBytes(downloaded),
				formatBytes(total),
				formatBytes(int64(speed)))
		}
	}

	err := downloadService.Download(testURL, destPath, progressCallback)
	if err != nil {
		log.Printf("下载失败: %v", err)
		return
	}

	duration := time.Since(startTime)
	fmt.Printf("\n下载完成，用时: %v\n", duration)

	// 第二次下载（演示缓存效果）
	fmt.Println("\n第二次下载（演示缓存效果）:")
	startTime = time.Now()

	err = downloadService.Download(testURL, destPath+"_cached", progressCallback)
	if err != nil {
		log.Printf("缓存下载失败: %v", err)
		return
	}

	cachedDuration := time.Since(startTime)
	fmt.Printf("\n缓存下载完成，用时: %v\n", cachedDuration)
	fmt.Printf("性能提升: %.2fx\n", float64(duration)/float64(cachedDuration))

	// 清理测试文件
	os.Remove(destPath)
	os.Remove(destPath + "_cached")
}

// demonstrateParallelExtraction 演示并行解压
func demonstrateParallelExtraction() {
	fmt.Println("\n2. 并行解压演示")
	fmt.Printf("CPU 核心数: %d\n", runtime.NumCPU())

	// 创建优化的解压选项
	options := &service.ArchiveExtractorOptions{
		PreservePermissions: true,
		OverwriteExisting:   true,
		ParallelWorkers:     runtime.NumCPU(), // 使用所有CPU核心
		BufferSize:          64 * 1024,        // 64KB 缓冲区
	}

	extractor := service.NewArchiveExtractor(options)

	fmt.Printf("并行工作协程数: %d\n", runtime.NumCPU())
	fmt.Printf("缓冲区大小: %s\n", formatBytes(64*1024))

	// 演示解压进度回调
	progressCallback := func(progress service.ExtractionProgress) {
		fmt.Printf("\r解压进度: %.1f%% (%d/%d 文件) 当前: %s",
			progress.Percentage,
			progress.ProcessedFiles,
			progress.TotalFiles,
			progress.CurrentFile)
	}

	fmt.Println("解压器已配置，支持以下优化:")
	fmt.Println("- 并行文件处理")
	fmt.Println("- 优化的缓冲区大小")
	fmt.Println("- 智能工作负载分配")

	// 注意：这里只是演示配置，实际解压需要真实的压缩文件
	_ = extractor
	_ = progressCallback
}

// demonstrateOptimizedProgress 演示进度显示优化
func demonstrateOptimizedProgress() {
	fmt.Println("\n3. 进度显示优化演示")

	// 创建进度跟踪器
	stages := []string{"检测系统", "下载文件", "解压文件", "配置环境"}
	tracker := ui.NewProgressTracker(stages)

	tracker.Start()

	fmt.Println("优化特性:")
	fmt.Println("- 限制更新频率减少CPU使用")
	fmt.Println("- 智能渲染控制")
	fmt.Println("- 内存优化的进度计算")

	// 模拟各个阶段的进度
	for i, stage := range stages {
		tracker.SetStage(stage)
		fmt.Printf("\n当前阶段: %s\n", stage)

		// 模拟阶段进度
		for progress := 0.0; progress <= 100.0; progress += 10.0 {
			tracker.SetProgress(progress)
			tracker.SetMessage(fmt.Sprintf("处理中... %.0f%%", progress))

			// 只在需要时渲染
			if tracker.ShouldRender() {
				status := tracker.RenderFullStatus(50)
				if status != "" {
					fmt.Print("\r" + status)
				}
			}

			time.Sleep(50 * time.Millisecond) // 模拟工作
		}

		// 阶段完成
		overallProgress := float64(i+1) / float64(len(stages)) * 100
		fmt.Printf("\n阶段 '%s' 完成 (总进度: %.1f%%)\n", stage, overallProgress)
	}

	tracker.Stop()
	fmt.Println("\n所有阶段完成!")

	// 显示性能统计
	info := tracker.GetProgress()
	fmt.Printf("总用时: %v\n", info.ElapsedTime)
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
