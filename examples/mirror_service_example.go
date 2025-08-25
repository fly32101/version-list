package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"version-list/internal/domain/service"
)

func main() {
	// 创建镜像服务
	mirrorService := service.NewMirrorService()

	fmt.Println("=== 镜像服务示例 ===\n")

	// 1. 获取所有可用镜像
	fmt.Println("1. 获取所有可用镜像:")
	mirrors := mirrorService.GetAvailableMirrors()
	for i, mirror := range mirrors {
		fmt.Printf("   %d. %s (%s) - %s\n", i+1, mirror.Name, mirror.Region, mirror.Description)
		fmt.Printf("      URL: %s\n", mirror.BaseURL)
	}
	fmt.Println()

	// 2. 根据名称获取特定镜像
	fmt.Println("2. 获取特定镜像 (official):")
	officialMirror, err := mirrorService.GetMirrorByName("official")
	if err != nil {
		log.Printf("获取镜像失败: %v", err)
	} else {
		fmt.Printf("   名称: %s\n", officialMirror.Name)
		fmt.Printf("   URL: %s\n", officialMirror.BaseURL)
		fmt.Printf("   描述: %s\n", officialMirror.Description)
	}
	fmt.Println()

	// 3. 测试镜像速度
	fmt.Println("3. 测试镜像速度:")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, mirror := range mirrors[:3] { // 只测试前3个镜像
		fmt.Printf("   测试 %s...", mirror.Name)
		result, err := mirrorService.TestMirrorSpeed(ctx, mirror)
		if err != nil {
			fmt.Printf(" 错误: %v\n", err)
			continue
		}

		if result.Available {
			fmt.Printf(" 可用 (响应时间: %v)\n", result.ResponseTime)
		} else {
			fmt.Printf(" 不可用")
			if result.Error != nil {
				fmt.Printf(" - %v", result.Error)
			}
			fmt.Println()
		}
	}
	fmt.Println()

	// 4. 验证镜像可用性
	fmt.Println("4. 验证官方镜像可用性:")
	if officialMirror != nil {
		err = mirrorService.ValidateMirror(ctx, *officialMirror)
		if err != nil {
			fmt.Printf("   官方镜像不可用: %v\n", err)
		} else {
			fmt.Println("   官方镜像可用")
		}
	}
	fmt.Println()

	// 5. 自动选择最快镜像
	fmt.Println("5. 自动选择最快镜像:")
	fmt.Println("   正在测试所有镜像...")

	fastestMirror, err := mirrorService.SelectFastestMirror(ctx, mirrors)
	if err != nil {
		fmt.Printf("   选择最快镜像失败: %v\n", err)
	} else {
		fmt.Printf("   最快镜像: %s (%s)\n", fastestMirror.Name, fastestMirror.Description)
		fmt.Printf("   URL: %s\n", fastestMirror.BaseURL)
	}

	// 6. 演示配置功能
	fmt.Println("6. 演示配置功能:")

	// 创建带配置的服务
	configService := service.NewMirrorService()

	// 添加自定义镜像
	customMirror := service.Mirror{
		Name:        "custom-example",
		BaseURL:     "https://example.com/golang/",
		Description: "示例自定义镜像",
		Region:      "example",
		Priority:    100,
	}

	fmt.Printf("   添加自定义镜像: %s\n", customMirror.Name)

	// 获取更新后的镜像列表
	updatedMirrors := configService.GetAvailableMirrors()
	fmt.Printf("   总镜像数量: %d\n", len(updatedMirrors))

	fmt.Println("\n=== 示例完成 ===")
}
