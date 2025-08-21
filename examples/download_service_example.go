package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"version-list/internal/domain/service"
)

func main() {
	fmt.Println("=== 下载服务示例 ===")

	// 创建下载服务
	downloadService := service.NewDownloadService(&service.DownloadOptions{
		MaxRetries: 3,
		RetryDelay: 2 * time.Second,
		Timeout:    30 * time.Second,
		ChunkSize:  8192,
		UserAgent:  "go-version-example/1.0",
	})

	// 测试URL（使用一个小文件进行测试）
	testURL := "https://golang.org/robots.txt"

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	destPath := filepath.Join(tempDir, "robots.txt")

	fmt.Printf("下载URL: %s\n", testURL)
	fmt.Printf("目标路径: %s\n", destPath)

	// 获取文件大小
	fmt.Println("\n检查文件大小...")
	size, err := downloadService.GetFileSize(testURL)
	if err != nil {
		log.Printf("获取文件大小失败: %v", err)
	} else {
		fmt.Printf("文件大小: %d 字节\n", size)
	}

	// 检查是否支持断点续传
	fmt.Println("\n检查断点续传支持...")
	supportsResume, err := downloadService.SupportsResume(testURL)
	if err != nil {
		log.Printf("检查断点续传支持失败: %v", err)
	} else {
		fmt.Printf("支持断点续传: %t\n", supportsResume)
	}

	// 下载文件
	fmt.Println("\n开始下载...")

	var lastProgress time.Time
	progressCallback := func(downloaded, total int64, speed float64) {
		now := time.Now()
		if now.Sub(lastProgress) >= 100*time.Millisecond {
			percentage := float64(downloaded) / float64(total) * 100
			speedKB := speed / 1024
			fmt.Printf("\r进度: %.1f%% (%d/%d 字节) 速度: %.1f KB/s",
				percentage, downloaded, total, speedKB)
			lastProgress = now
		}
	}

	startTime := time.Now()
	err = downloadService.Download(testURL, destPath, progressCallback)
	duration := time.Since(startTime)

	if err != nil {
		log.Fatalf("下载失败: %v", err)
	}

	fmt.Printf("\n下载完成! 耗时: %v\n", duration)

	// 验证下载的文件
	fileInfo, err := os.Stat(destPath)
	if err != nil {
		log.Fatalf("获取文件信息失败: %v", err)
	}

	fmt.Printf("下载文件大小: %d 字节\n", fileInfo.Size())

	// 读取并显示文件内容的前几行
	content, err := os.ReadFile(destPath)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}

	fmt.Println("\n文件内容预览:")
	lines := string(content)
	if len(lines) > 200 {
		lines = lines[:200] + "..."
	}
	fmt.Println(lines)

	fmt.Println("\n=== 示例完成 ===")
}
