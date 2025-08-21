package main

import (
	"fmt"
	"log"
	"version-list/internal/domain/service"
)

func main() {
	// 创建系统检测服务
	detector := service.NewSystemDetector()

	// 检测当前系统信息
	fmt.Println("=== 系统检测服务示例 ===")

	os, err := detector.DetectOS()
	if err != nil {
		log.Fatalf("检测操作系统失败: %v", err)
	}
	fmt.Printf("操作系统: %s\n", os)

	arch, err := detector.DetectArch()
	if err != nil {
		log.Fatalf("检测CPU架构失败: %v", err)
	}
	fmt.Printf("CPU架构: %s\n", arch)

	// 获取Go 1.21.0的完整系统信息
	version := "1.21.0"
	systemInfo, err := detector.GetSystemInfo(version)
	if err != nil {
		log.Fatalf("获取系统信息失败: %v", err)
	}

	fmt.Printf("\n=== Go %s 下载信息 ===\n", version)
	fmt.Printf("文件名: %s\n", systemInfo.Filename)
	fmt.Printf("下载URL: %s\n", systemInfo.URL)

	// 测试其他版本
	versions := []string{"1.20.5", "1.19.10", "1.22.0"}
	fmt.Printf("\n=== 其他版本下载信息 ===\n")
	for _, v := range versions {
		info, err := detector.GetSystemInfo(v)
		if err != nil {
			fmt.Printf("版本 %s: 错误 - %v\n", v, err)
			continue
		}
		fmt.Printf("版本 %s: %s\n", v, info.Filename)
	}
}
