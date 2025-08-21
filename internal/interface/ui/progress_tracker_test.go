package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestProgressTracker_Basic(t *testing.T) {
	stages := []string{"stage1", "stage2", "stage3"}
	tracker := NewProgressTracker(stages)

	// 测试初始状态
	if tracker.IsActive() {
		t.Error("新创建的跟踪器不应该处于活跃状态")
	}

	// 测试启动
	tracker.Start()
	if !tracker.IsActive() {
		t.Error("启动后跟踪器应该处于活跃状态")
	}

	// 测试停止
	tracker.Stop()
	if tracker.IsActive() {
		t.Error("停止后跟踪器不应该处于活跃状态")
	}
}

func TestProgressTracker_Progress(t *testing.T) {
	stages := []string{"download", "extract", "configure"}
	tracker := NewProgressTracker(stages)
	tracker.Start()

	// 测试设置进度
	tracker.SetProgress(50.0)
	info := tracker.GetProgress()

	if info.Progress != 50.0 {
		t.Errorf("进度 = %.1f, 期望 %.1f", info.Progress, 50.0)
	}

	// 测试进度边界
	tracker.SetProgress(-10.0)
	info = tracker.GetProgress()
	if info.Progress != 0.0 {
		t.Errorf("负进度应该被限制为0, 得到 %.1f", info.Progress)
	}

	tracker.SetProgress(150.0)
	info = tracker.GetProgress()
	if info.Progress != 100.0 {
		t.Errorf("超过100的进度应该被限制为100, 得到 %.1f", info.Progress)
	}
}

func TestProgressTracker_Stages(t *testing.T) {
	stages := []string{"检测", "下载", "解压"}
	tracker := NewProgressTracker(stages)
	tracker.Start()

	// 测试设置阶段
	tracker.SetStage("下载")
	info := tracker.GetProgress()

	if info.Stage != "下载" {
		t.Errorf("阶段 = %s, 期望 %s", info.Stage, "下载")
	}

	// 测试更新进度和阶段
	tracker.UpdateProgress("解压", 75.0, "正在解压文件...")
	info = tracker.GetProgress()

	if info.Stage != "解压" {
		t.Errorf("阶段 = %s, 期望 %s", info.Stage, "解压")
	}

	if info.Progress != 75.0 {
		t.Errorf("进度 = %.1f, 期望 %.1f", info.Progress, 75.0)
	}

	if info.Message != "正在解压文件..." {
		t.Errorf("消息 = %s, 期望 %s", info.Message, "正在解压文件...")
	}
}

func TestProgressTracker_RenderProgressBar(t *testing.T) {
	tracker := NewProgressTracker([]string{"test"})
	tracker.Start()

	// 测试0%进度
	tracker.SetProgress(0)
	bar := tracker.RenderProgressBar(10)

	if !strings.Contains(bar, "0.0%") {
		t.Errorf("进度条应该显示0.0%%, 得到: %s", bar)
	}

	// 测试50%进度
	tracker.SetProgress(50)
	bar = tracker.RenderProgressBar(10)

	if !strings.Contains(bar, "50.0%") {
		t.Errorf("进度条应该显示50.0%%, 得到: %s", bar)
	}

	// 测试100%进度
	tracker.SetProgress(100)
	bar = tracker.RenderProgressBar(10)

	if !strings.Contains(bar, "100.0%") {
		t.Errorf("进度条应该显示100.0%%, 得到: %s", bar)
	}

	// 验证进度条格式
	if !strings.Contains(bar, "[") || !strings.Contains(bar, "]") {
		t.Errorf("进度条应该包含方括号, 得到: %s", bar)
	}
}

func TestProgressTracker_RenderSpinner(t *testing.T) {
	tracker := NewProgressTracker([]string{"test"})
	tracker.Start()

	// 测试旋转器
	spinner1 := tracker.RenderSpinner()
	spinner2 := tracker.RenderSpinner()

	if spinner1 == "" {
		t.Error("旋转器不应该为空")
	}

	// 旋转器应该会变化（虽然可能相同）
	if len(spinner1) == 0 || len(spinner2) == 0 {
		t.Error("旋转器应该有内容")
	}
}

func TestProgressTracker_TimeCalculation(t *testing.T) {
	tracker := NewProgressTracker([]string{"test"})
	tracker.Start()

	// 等待一小段时间
	time.Sleep(10 * time.Millisecond)

	tracker.SetProgress(25.0)
	info := tracker.GetProgress()

	// 验证已用时间
	if info.ElapsedTime <= 0 {
		t.Error("已用时间应该大于0")
	}

	// 验证ETA计算
	if info.ETA <= 0 {
		t.Error("ETA应该大于0（当进度在0-100之间时）")
	}

	// 测试100%进度时的ETA
	tracker.SetProgress(100.0)
	info = tracker.GetProgress()

	if info.ETA != 0 {
		t.Error("100%进度时ETA应该为0")
	}
}

func TestProgressTracker_CalculateOverallProgress(t *testing.T) {
	stages := []string{"stage1", "stage2", "stage3"}
	tracker := NewProgressTracker(stages)
	tracker.Start()

	// 定义阶段权重
	weights := map[string]float64{
		"stage1": 1.0,
		"stage2": 2.0,
		"stage3": 1.0,
	}

	// 测试第一阶段50%进度
	tracker.UpdateProgress("stage1", 50.0, "")
	overall := tracker.CalculateOverallProgress(weights)

	expected := (1.0 * 0.5) / 4.0 * 100.0 // 12.5%
	if overall < expected-1 || overall > expected+1 {
		t.Errorf("总体进度 = %.1f, 期望约 %.1f", overall, expected)
	}

	// 测试第二阶段开始
	tracker.UpdateProgress("stage2", 0.0, "")
	overall = tracker.CalculateOverallProgress(weights)

	expected = (1.0) / 4.0 * 100.0 // 25%
	if overall < expected-1 || overall > expected+1 {
		t.Errorf("总体进度 = %.1f, 期望约 %.1f", overall, expected)
	}
}

func TestProgressTracker_RenderFullStatus(t *testing.T) {
	tracker := NewProgressTracker([]string{"测试阶段"})
	tracker.Start()

	tracker.UpdateProgress("测试阶段", 60.0, "正在处理...")

	status := tracker.RenderFullStatus(20)

	// 验证状态包含关键信息
	if !strings.Contains(status, "测试阶段") {
		t.Error("完整状态应该包含阶段名称")
	}

	if !strings.Contains(status, "60.0%") {
		t.Error("完整状态应该包含进度百分比")
	}

	if !strings.Contains(status, "正在处理...") {
		t.Error("完整状态应该包含消息")
	}

	if !strings.Contains(status, "已用时间") {
		t.Error("完整状态应该包含时间信息")
	}
}

func TestProgressTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewProgressTracker([]string{"test"})
	tracker.Start()

	// 并发访问测试
	done := make(chan bool, 2)

	// 写入goroutine
	go func() {
		for i := 0; i < 100; i++ {
			tracker.SetProgress(float64(i))
			tracker.SetMessage(fmt.Sprintf("Progress %d", i))
		}
		done <- true
	}()

	// 读取goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = tracker.GetProgress()
			_ = tracker.RenderProgressBar(20)
		}
		done <- true
	}()

	// 等待完成
	<-done
	<-done

	// 如果没有panic，说明并发访问是安全的
	t.Log("并发访问测试通过")
}
