package ui

import (
	"testing"
	"time"
)

// TestProgressUIIntegration 测试进度UI的集成功能
func TestProgressUIIntegration(t *testing.T) {
	// 创建安装进度UI
	progressUI := NewInstallProgressUI()

	// 测试完整的安装流程模拟
	t.Run("完整安装流程", func(t *testing.T) {
		progressUI.Start()

		// 模拟各个阶段
		stages := []struct {
			name     string
			progress float64
			message  string
		}{
			{"检测系统", 10, "检测操作系统和架构..."},
			{"下载文件", 30, "下载Go安装包..."},
			{"验证文件", 50, "验证文件完整性..."},
			{"解压文件", 70, "解压安装包..."},
			{"配置安装", 90, "配置环境变量..."},
			{"完成安装", 100, "安装完成"},
		}

		for _, stage := range stages {
			progressUI.UpdateProgress(stage.name, stage.progress, stage.message)

			// 验证进度更新
			info := progressUI.tracker.GetProgress()
			if info.Stage != stage.name {
				t.Errorf("阶段 = %s, 期望 %s", info.Stage, stage.name)
			}

			if info.Progress != stage.progress {
				t.Errorf("进度 = %.1f, 期望 %.1f", info.Progress, stage.progress)
			}

			if info.Message != stage.message {
				t.Errorf("消息 = %s, 期望 %s", info.Message, stage.message)
			}

			// 短暂等待以模拟实际操作
			time.Sleep(10 * time.Millisecond)
		}

		progressUI.Stop()

		// 验证最终状态
		if progressUI.tracker.IsActive() {
			t.Error("停止后跟踪器不应该处于活跃状态")
		}
	})

	t.Run("错误处理流程", func(t *testing.T) {
		progressUI.Start()

		// 模拟到一半出错
		progressUI.UpdateProgress("下载文件", 45, "下载中...")

		// 模拟错误
		progressUI.PrintError("下载失败: 网络连接超时")

		// 验证可以继续使用
		progressUI.PrintInfo("正在重试...")
		progressUI.UpdateProgress("下载文件", 50, "重新下载...")

		progressUI.Stop()
	})
}

// TestProgressTrackerPerformance 测试进度跟踪器的性能
func TestProgressTrackerPerformance(t *testing.T) {
	tracker := NewProgressTracker([]string{"test"})
	tracker.Start()

	// 测试大量更新的性能
	start := time.Now()

	for i := 0; i < 1000; i++ {
		tracker.SetProgress(float64(i % 101))
		tracker.SetMessage("Processing...")
		_ = tracker.GetProgress()
		_ = tracker.RenderProgressBar(50)
	}

	elapsed := time.Since(start)

	// 1000次操作应该在合理时间内完成
	if elapsed > time.Second {
		t.Errorf("性能测试耗时过长: %v", elapsed)
	}

	tracker.Stop()
}

// TestConsoleUIPerformance 测试控制台UI的性能
func TestConsoleUIPerformance(t *testing.T) {
	ui := NewConsoleUI(&ConsoleUIOptions{
		UpdateInterval: 1 * time.Millisecond, // 高频更新
		ColorEnabled:   false,
		Width:          30,
	})

	tracker := NewProgressTracker([]string{"performance_test"})
	ui.SetProgressTracker(tracker)

	tracker.Start()
	ui.Start()

	// 快速更新进度
	start := time.Now()

	for i := 0; i <= 100; i++ {
		tracker.SetProgress(float64(i))
		time.Sleep(1 * time.Millisecond)
	}

	elapsed := time.Since(start)

	ui.Stop()
	tracker.Stop()

	// 验证性能合理
	if elapsed > 5*time.Second {
		t.Errorf("UI性能测试耗时过长: %v", elapsed)
	}
}

// TestProgressUIMemoryUsage 测试进度UI的内存使用
func TestProgressUIMemoryUsage(t *testing.T) {
	// 创建多个UI实例以测试内存使用
	var uis []*InstallProgressUI

	for i := 0; i < 10; i++ {
		ui := NewInstallProgressUI()
		ui.Start()
		uis = append(uis, ui)
	}

	// 模拟使用
	for _, ui := range uis {
		for j := 0; j <= 100; j += 10 {
			ui.SetProgress(float64(j))
			ui.SetMessage("Testing memory usage...")
		}
	}

	// 清理
	for _, ui := range uis {
		ui.Stop()
	}

	// 如果没有内存泄漏，这个测试应该正常完成
	t.Log("内存使用测试完成")
}

// TestProgressUIEdgeCases 测试进度UI的边界情况
func TestProgressUIEdgeCases(t *testing.T) {
	ui := NewInstallProgressUI()

	t.Run("重复启动停止", func(t *testing.T) {
		// 多次启动和停止
		for i := 0; i < 5; i++ {
			ui.Start()
			ui.Stop()
		}

		// 应该能正常工作
		ui.Start()
		ui.SetProgress(50)
		info := ui.tracker.GetProgress()

		if info.Progress != 50 {
			t.Errorf("重复启动停止后进度设置失败: %.1f", info.Progress)
		}

		ui.Stop()
	})

	t.Run("极值测试", func(t *testing.T) {
		ui.Start()

		// 测试极大值
		ui.SetProgress(999999)
		info := ui.tracker.GetProgress()
		if info.Progress != 100 {
			t.Errorf("极大进度值应该被限制为100: %.1f", info.Progress)
		}

		// 测试极小值
		ui.SetProgress(-999999)
		info = ui.tracker.GetProgress()
		if info.Progress != 0 {
			t.Errorf("极小进度值应该被限制为0: %.1f", info.Progress)
		}

		// 测试空消息
		ui.SetMessage("")
		info = ui.tracker.GetProgress()
		if info.Message != "" {
			t.Error("空消息应该被正确设置")
		}

		ui.Stop()
	})

	t.Run("并发操作", func(t *testing.T) {
		ui.Start()

		// 并发更新进度
		done := make(chan bool, 3)

		go func() {
			for i := 0; i < 50; i++ {
				ui.SetProgress(float64(i * 2))
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 50; i++ {
				ui.SetMessage("Concurrent update")
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 50; i++ {
				_ = ui.tracker.GetProgress()
			}
			done <- true
		}()

		// 等待所有goroutine完成
		for i := 0; i < 3; i++ {
			<-done
		}

		ui.Stop()

		// 如果没有panic或死锁，说明并发安全
		t.Log("并发操作测试通过")
	})
}
