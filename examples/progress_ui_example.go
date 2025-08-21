package main

import (
	"fmt"
	"time"
	"version-list/internal/interface/ui"
)

func main() {
	fmt.Println("=== 进度显示UI示例 ===")

	// 创建安装进度UI
	progressUI := ui.NewInstallProgressUI()

	fmt.Println("开始模拟安装过程...")

	// 启动进度显示
	progressUI.Start()

	// 模拟安装过程
	simulateInstallation(progressUI)

	// 停止进度显示
	progressUI.Stop()

	fmt.Println("\n=== 示例完成 ===")
}

func simulateInstallation(progressUI *ui.InstallProgressUI) {
	stages := []struct {
		name     string
		duration time.Duration
		messages []string
	}{
		{
			name:     "检测系统",
			duration: 1 * time.Second,
			messages: []string{
				"检测操作系统...",
				"检测CPU架构...",
				"验证系统兼容性...",
			},
		},
		{
			name:     "下载文件",
			duration: 3 * time.Second,
			messages: []string{
				"连接到下载服务器...",
				"开始下载Go安装包...",
				"下载进行中...",
				"验证下载完整性...",
			},
		},
		{
			name:     "验证文件",
			duration: 500 * time.Millisecond,
			messages: []string{
				"验证文件签名...",
				"检查文件完整性...",
			},
		},
		{
			name:     "解压文件",
			duration: 2 * time.Second,
			messages: []string{
				"创建安装目录...",
				"解压安装包...",
				"设置文件权限...",
			},
		},
		{
			name:     "配置安装",
			duration: 1 * time.Second,
			messages: []string{
				"配置环境变量...",
				"创建符号链接...",
				"验证安装...",
			},
		},
		{
			name:     "完成安装",
			duration: 500 * time.Millisecond,
			messages: []string{
				"清理临时文件...",
				"更新版本记录...",
			},
		},
	}

	for _, stage := range stages {
		progressUI.SetStage(stage.name)
		progressUI.SetProgress(0)

		stepDuration := stage.duration / time.Duration(len(stage.messages))

		for i, message := range stage.messages {
			progressUI.SetMessage(message)

			// 模拟该步骤的进度
			progress := float64(i+1) / float64(len(stage.messages)) * 100
			progressUI.SetProgress(progress)

			time.Sleep(stepDuration)
		}
	}

	// 完成安装
	progressUI.SetProgress(100)
	progressUI.SetMessage("安装完成!")

	time.Sleep(500 * time.Millisecond)

	progressUI.PrintSuccess("Go 1.21.0 安装成功!")
}

// 演示基本进度跟踪器的使用
func demonstrateBasicProgressTracker() {
	fmt.Println("\n=== 基本进度跟踪器示例 ===")

	stages := []string{"准备", "处理", "完成"}
	tracker := ui.NewProgressTracker(stages)

	tracker.Start()

	for i, stage := range stages {
		tracker.SetStage(stage)

		for progress := 0; progress <= 100; progress += 10 {
			tracker.SetProgress(float64(progress))
			tracker.SetMessage(fmt.Sprintf("处理 %s: %d%%", stage, progress))

			// 显示进度条
			bar := tracker.RenderProgressBar(30)
			spinner := tracker.RenderSpinner()
			info := tracker.GetProgress()

			fmt.Printf("\r%s %s - %s (阶段 %d/%d)",
				spinner, bar, info.Message, i+1, len(stages))

			time.Sleep(50 * time.Millisecond)
		}
	}

	tracker.Stop()
	fmt.Println("\n基本进度跟踪器示例完成!")
}

// 演示控制台UI的使用
func demonstrateConsoleUI() {
	fmt.Println("\n=== 控制台UI示例 ===")

	// 创建控制台UI
	consoleUI := ui.NewConsoleUI(&ui.ConsoleUIOptions{
		UpdateInterval: 100 * time.Millisecond,
		ColorEnabled:   true,
		Width:          40,
	})

	// 创建进度跟踪器
	tracker := ui.NewProgressTracker([]string{"初始化", "执行", "清理"})
	consoleUI.SetProgressTracker(tracker)

	// 启动
	tracker.Start()
	consoleUI.Start()

	// 模拟工作
	for i := 0; i <= 100; i += 5 {
		stage := "初始化"
		if i > 33 {
			stage = "执行"
		}
		if i > 66 {
			stage = "清理"
		}

		tracker.UpdateProgress(stage, float64(i), fmt.Sprintf("进度: %d%%", i))
		time.Sleep(100 * time.Millisecond)
	}

	// 停止
	consoleUI.Stop()
	tracker.Stop()

	// 显示结果消息
	consoleUI.PrintSuccess("任务完成!")
	consoleUI.PrintInfo("所有操作已成功执行")
	consoleUI.PrintWarning("请注意保存您的工作")
}

func init() {
	// 可以在这里添加其他示例的调用
	// demonstrateBasicProgressTracker()
	// demonstrateConsoleUI()
}
