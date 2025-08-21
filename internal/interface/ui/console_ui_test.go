package ui

import (
	"strings"
	"testing"
	"time"
)

func TestConsoleUI_Creation(t *testing.T) {
	// 测试默认选项
	ui := NewConsoleUI(nil)
	if ui == nil {
		t.Fatal("NewConsoleUI() 返回 nil")
	}

	// 测试自定义选项
	options := &ConsoleUIOptions{
		UpdateInterval: 50 * time.Millisecond,
		ColorEnabled:   false,
		Width:          30,
	}

	ui = NewConsoleUI(options)
	if ui == nil {
		t.Fatal("NewConsoleUI() 返回 nil")
	}

	if ui.updateInterval != options.UpdateInterval {
		t.Errorf("UpdateInterval = %v, 期望 %v", ui.updateInterval, options.UpdateInterval)
	}

	if ui.colorEnabled != options.ColorEnabled {
		t.Errorf("ColorEnabled = %v, 期望 %v", ui.colorEnabled, options.ColorEnabled)
	}

	if ui.width != options.Width {
		t.Errorf("Width = %d, 期望 %d", ui.width, options.Width)
	}
}

func TestConsoleUI_ProgressTracker(t *testing.T) {
	ui := NewConsoleUI(nil)
	tracker := NewProgressTracker([]string{"test"})

	// 测试设置进度跟踪器
	ui.SetProgressTracker(tracker)

	ui.mu.RLock()
	if ui.progressTracker != tracker {
		t.Error("进度跟踪器设置失败")
	}
	ui.mu.RUnlock()
}

func TestConsoleUI_StartStop(t *testing.T) {
	ui := NewConsoleUI(&ConsoleUIOptions{
		UpdateInterval: 10 * time.Millisecond,
		ColorEnabled:   false,
		Width:          20,
	})

	// 测试初始状态
	if ui.isRunning {
		t.Error("新创建的UI不应该处于运行状态")
	}

	// 测试启动
	ui.Start()

	// 等待一小段时间让goroutine启动
	time.Sleep(20 * time.Millisecond)

	ui.mu.RLock()
	running := ui.isRunning
	ui.mu.RUnlock()

	if !running {
		t.Error("启动后UI应该处于运行状态")
	}

	// 测试停止
	ui.Stop()

	// 等待一小段时间让goroutine停止
	time.Sleep(20 * time.Millisecond)

	ui.mu.RLock()
	running = ui.isRunning
	ui.mu.RUnlock()

	if running {
		t.Error("停止后UI不应该处于运行状态")
	}
}

func TestConsoleUI_Colorize(t *testing.T) {
	// 测试启用颜色
	ui := NewConsoleUI(&ConsoleUIOptions{
		ColorEnabled: true,
	})

	text := "测试文本"
	colored := ui.colorize(text, ColorRed)

	// 应该包含颜色代码（如果终端支持）
	if ui.isTerminalSupported() {
		if !strings.Contains(colored, string(ColorRed)) {
			t.Error("启用颜色时应该包含颜色代码")
		}

		if !strings.Contains(colored, string(ColorReset)) {
			t.Error("启用颜色时应该包含重置代码")
		}
	}

	// 测试禁用颜色
	ui.colorEnabled = false
	uncolored := ui.colorize(text, ColorRed)

	if uncolored != text {
		t.Errorf("禁用颜色时应该返回原文本, 得到: %s", uncolored)
	}
}

func TestConsoleUI_BuildProgressOutput(t *testing.T) {
	ui := NewConsoleUI(&ConsoleUIOptions{
		ColorEnabled: false, // 禁用颜色以便测试
		Width:        20,
	})

	tracker := NewProgressTracker([]string{"测试阶段"})
	tracker.Start()
	tracker.UpdateProgress("测试阶段", 50.0, "正在处理...")

	ui.SetProgressTracker(tracker)

	output := ui.buildProgressOutput()

	// 验证输出包含关键信息
	if !strings.Contains(output, "测试阶段") {
		t.Error("输出应该包含阶段名称")
	}

	if !strings.Contains(output, "50.0%") {
		t.Error("输出应该包含进度百分比")
	}

	if !strings.Contains(output, "正在处理...") {
		t.Error("输出应该包含消息")
	}

	if !strings.Contains(output, "已用时间") {
		t.Error("输出应该包含时间信息")
	}
}

func TestInstallProgressUI_Creation(t *testing.T) {
	ui := NewInstallProgressUI()

	if ui == nil {
		t.Fatal("NewInstallProgressUI() 返回 nil")
	}

	if ui.consoleUI == nil {
		t.Error("ConsoleUI 不应该为 nil")
	}

	if ui.tracker == nil {
		t.Error("ProgressTracker 不应该为 nil")
	}

	if len(ui.stages) == 0 {
		t.Error("应该有预定义的阶段")
	}

	// 验证预定义阶段
	expectedStages := []string{
		"检测系统",
		"下载文件",
		"验证文件",
		"解压文件",
		"配置安装",
		"完成安装",
	}

	if len(ui.stages) != len(expectedStages) {
		t.Errorf("阶段数量 = %d, 期望 %d", len(ui.stages), len(expectedStages))
	}

	for i, expected := range expectedStages {
		if i < len(ui.stages) && ui.stages[i] != expected {
			t.Errorf("阶段[%d] = %s, 期望 %s", i, ui.stages[i], expected)
		}
	}
}

func TestInstallProgressUI_Operations(t *testing.T) {
	ui := NewInstallProgressUI()

	// 测试启动和停止
	ui.Start()

	if !ui.tracker.IsActive() {
		t.Error("启动后跟踪器应该处于活跃状态")
	}

	// 测试设置阶段
	ui.SetStage("下载文件")
	info := ui.tracker.GetProgress()

	if info.Stage != "下载文件" {
		t.Errorf("阶段 = %s, 期望 %s", info.Stage, "下载文件")
	}

	// 测试设置进度
	ui.SetProgress(75.0)
	info = ui.tracker.GetProgress()

	if info.Progress != 75.0 {
		t.Errorf("进度 = %.1f, 期望 %.1f", info.Progress, 75.0)
	}

	// 测试设置消息
	ui.SetMessage("正在下载...")
	info = ui.tracker.GetProgress()

	if info.Message != "正在下载..." {
		t.Errorf("消息 = %s, 期望 %s", info.Message, "正在下载...")
	}

	// 测试更新进度
	ui.UpdateProgress("解压文件", 90.0, "正在解压...")
	info = ui.tracker.GetProgress()

	if info.Stage != "解压文件" {
		t.Errorf("阶段 = %s, 期望 %s", info.Stage, "解压文件")
	}

	if info.Progress != 90.0 {
		t.Errorf("进度 = %.1f, 期望 %.1f", info.Progress, 90.0)
	}

	if info.Message != "正在解压..." {
		t.Errorf("消息 = %s, 期望 %s", info.Message, "正在解压...")
	}

	// 测试停止
	ui.Stop()

	if ui.tracker.IsActive() {
		t.Error("停止后跟踪器不应该处于活跃状态")
	}
}

func TestConsoleUI_MessageMethods(t *testing.T) {
	ui := NewConsoleUI(&ConsoleUIOptions{
		ColorEnabled: false, // 禁用颜色以便测试
	})

	// 这些方法主要是输出到控制台，我们只测试它们不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("消息方法不应该panic: %v", r)
		}
	}()

	ui.PrintSuccess("成功消息")
	ui.PrintError("错误消息")
	ui.PrintWarning("警告消息")
	ui.PrintInfo("信息消息")
}

func TestProgressTracker_EdgeCases(t *testing.T) {
	// 测试空阶段列表
	tracker := NewProgressTracker([]string{})
	tracker.Start()

	info := tracker.GetProgress()
	if info.Stage != "" {
		t.Error("空阶段列表时阶段应该为空")
	}

	// 测试nil阶段列表
	tracker = NewProgressTracker(nil)
	tracker.Start()

	info = tracker.GetProgress()
	if info.Stage != "" {
		t.Error("nil阶段列表时阶段应该为空")
	}

	// 测试设置不存在的阶段
	tracker = NewProgressTracker([]string{"stage1", "stage2"})
	tracker.Start()
	tracker.SetStage("nonexistent")

	info = tracker.GetProgress()
	if info.Stage != "nonexistent" {
		t.Error("应该能设置不存在的阶段")
	}
}

func TestConsoleUI_TerminalSupport(t *testing.T) {
	ui := NewConsoleUI(nil)

	// 测试终端支持检测（这个测试可能因环境而异）
	supported := ui.isTerminalSupported()

	// 我们只验证方法不会panic
	if supported {
		t.Log("检测到终端支持颜色")
	} else {
		t.Log("检测到终端不支持颜色")
	}
}
