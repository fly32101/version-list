package main

import (
	"context"
	"fmt"
	"log"

	"version-list/internal/domain/service"
)

func main() {
	fmt.Println("=== Go版本管理器镜像功能示例 ===")
	fmt.Println()

	// 创建镜像服务
	mirrorService := service.NewMirrorService()

	// 1. 获取所有可用镜像
	fmt.Println("1. 获取所有可用镜像:")
	mirrors := mirrorService.GetAvailableMirrors()
	for _, mirror := range mirrors {
		fmt.Printf("   - %s: %s (%s)\n", mirror.Name, mirror.Description, mirror.Region)
		fmt.Printf("     URL: %s\n", mirror.BaseURL)
		fmt.Printf("     优先级: %d\n", mirror.Priority)
		fmt.Println()
	}

	// 2. 根据名称获取镜像
	fmt.Println("2. 根据名称获取镜像:")
	mirror, err := mirrorService.GetMirrorByName("goproxy-cn")
	if err != nil {
		log.Printf("获取镜像失败: %v", err)
	} else {
		fmt.Printf("   找到镜像: %s - %s\n", mirror.Name, mirror.Description)
	}
	fmt.Println()

	// 3. 测试镜像速度
	fmt.Println("3. 测试镜像速度:")
	ctx := context.Background()

	// 测试前3个镜像的速度
	testMirrors := mirrors[:3]
	for _, m := range testMirrors {
		fmt.Printf("   正在测试 %s...", m.Name)
		result, err := mirrorService.TestMirrorSpeed(ctx, m)
		if err != nil {
			fmt.Printf(" 错误: %v\n", err)
			continue
		}

		if result.Available {
			fmt.Printf(" 可用 (响应时间: %v)\n", result.ResponseTime)
		} else {
			fmt.Printf(" 不可用")
			if result.Error != nil {
				fmt.Printf(" (错误: %v)", result.Error)
			}
			fmt.Println()
		}
	}
	fmt.Println()

	// 4. 自动选择最快镜像
	fmt.Println("4. 自动选择最快镜像:")
	fmt.Println("   正在测试所有镜像速度...")

	fastest, err := mirrorService.SelectFastestMirror(ctx, mirrors)
	if err != nil {
		log.Printf("选择最快镜像失败: %v", err)
	} else {
		fmt.Printf("   最快镜像: %s (%s)\n", fastest.Name, fastest.Description)
	}
	fmt.Println()

	// 5. 验证镜像可用性
	fmt.Println("5. 验证镜像可用性:")
	if mirror != nil {
		fmt.Printf("   正在验证 %s...", mirror.Name)
		err := mirrorService.ValidateMirror(ctx, *mirror)
		if err != nil {
			fmt.Printf(" 验证失败: %v\n", err)
		} else {
			fmt.Printf(" 验证通过\n")
		}
	}
	fmt.Println()

	// 6. 演示系统检测器的镜像功能
	fmt.Println("6. 系统检测器镜像功能:")
	systemDetector := service.NewSystemDetector()

	version := "1.21.0"

	// 使用不同镜像构建下载URL
	mirrorNames := []string{"official", "goproxy-cn", "aliyun"}
	for _, mirrorName := range mirrorNames {
		systemInfo, err := systemDetector.GetSystemInfoWithMirror(version, mirrorName)
		if err != nil {
			log.Printf("获取系统信息失败: %v", err)
			continue
		}

		fmt.Printf("   %s: %s\n", mirrorName, systemInfo.URL)
	}
	fmt.Println()

	fmt.Println("=== 镜像功能示例完成 ===")
}
