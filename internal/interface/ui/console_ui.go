package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ConsoleUI 控制台用户界面
type ConsoleUI struct {
	mu              sync.RWMutex
	progressTracker *ProgressTracker
	isRunning       bool
	updateInterval  time.Duration
	lastOutput      string
	colorEnabled    bool
	width           int
}

// ConsoleUIOptions 控制台UI选项
type ConsoleUIOptions struct {
	UpdateInterval time.Duration // 更新间隔
	ColorEnabled   bool          // 是否启用颜色
	Width          int           // 进度条宽度
}

// NewConsoleUI 创建控制台UI
func NewConsoleUI(options *ConsoleUIOptions) *ConsoleUI {
	if options == nil {
		options = &ConsoleUIOptions{
			UpdateInterval: 100 * time.Millisecond,
			ColorEnabled:   true,
			Width:          50,
		}
	}

	return &ConsoleUI{
		updateInterval: options.UpdateInterval,
		colorEnabled:   options.ColorEnabled,
		width:          options.Width,
	}
}

// SetProgressTracker 设置进度跟踪器
func (c *ConsoleUI) SetProgressTracker(tracker *ProgressTracker) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.progressTracker = tracker
}

// Start 开始显示进度
func (c *ConsoleUI) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return
	}

	c.isRunning = true
	go c.updateLoop()
}

// Stop 停止显示进度
func (c *ConsoleUI) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isRunning = false
}

// updateLoop 更新循环
func (c *ConsoleUI) updateLoop() {
	ticker := time.NewTicker(c.updateInterval)
	defer ticker.Stop()

	for {
		c.mu.RLock()
		running := c.isRunning
		tracker := c.progressTracker
		c.mu.RUnlock()

		if !running {
			break
		}

		if tracker != nil && tracker.IsActive() {
			c.renderProgress()
		}

		<-ticker.C
	}
}

// renderProgress 渲染进度
func (c *ConsoleUI) renderProgress() {
	if c.progressTracker == nil {
		return
	}

	output := c.buildProgressOutput()

	c.mu.Lock()
	defer c.mu.Unlock()

	if output != c.lastOutput {
		c.clearLastOutput()
		fmt.Print(output)
		c.lastOutput = output
	}
}

// buildProgressOutput 构建进度输出
func (c *ConsoleUI) buildProgressOutput() string {
	info := c.progressTracker.GetProgress()

	var lines []string

	// 阶段信息
	if info.Stage != "" {
		stageLine := c.colorize(fmt.Sprintf("📋 阶段: %s", info.Stage), ColorBlue)
		lines = append(lines, stageLine)
	}

	// 进度条
	progressBar := c.progressTracker.RenderProgressBar(c.width)
	progressLine := c.colorize(progressBar, ColorGreen)

	// 添加旋转器
	if c.progressTracker.IsActive() && info.Progress < 100 {
		spinner := c.progressTracker.RenderSpinner()
		progressLine = fmt.Sprintf("%s %s", spinner, progressLine)
	}

	lines = append(lines, progressLine)

	// 状态消息
	if info.Message != "" {
		messageLine := c.colorize(fmt.Sprintf("💬 %s", info.Message), ColorYellow)
		lines = append(lines, messageLine)
	}

	// 时间信息
	elapsed := info.ElapsedTime.Truncate(time.Second)
	timeInfo := fmt.Sprintf("⏱️  已用时间: %v", elapsed)

	if info.ETA > 0 && info.Progress < 100 {
		eta := info.ETA.Truncate(time.Second)
		timeInfo += fmt.Sprintf(", 预计剩余: %v", eta)
	}

	timeLine := c.colorize(timeInfo, ColorCyan)
	lines = append(lines, timeLine)

	return strings.Join(lines, "\n") + "\n"
}

// clearLastOutput 清除上次输出
func (c *ConsoleUI) clearLastOutput() {
	if c.lastOutput == "" {
		return
	}

	lines := strings.Count(c.lastOutput, "\n")
	for i := 0; i < lines; i++ {
		fmt.Print("\033[1A\033[2K") // 向上移动一行并清除
	}
}

// PrintSuccess 打印成功消息
func (c *ConsoleUI) PrintSuccess(message string) {
	c.Stop()
	c.clearLastOutput()

	successMsg := c.colorize(fmt.Sprintf("✅ %s", message), ColorGreen)
	fmt.Println(successMsg)
}

// PrintError 打印错误消息
func (c *ConsoleUI) PrintError(message string) {
	c.Stop()
	c.clearLastOutput()

	errorMsg := c.colorize(fmt.Sprintf("❌ %s", message), ColorRed)
	fmt.Println(errorMsg)
}

// PrintWarning 打印警告消息
func (c *ConsoleUI) PrintWarning(message string) {
	warningMsg := c.colorize(fmt.Sprintf("⚠️  %s", message), ColorYellow)
	fmt.Println(warningMsg)
}

// PrintInfo 打印信息消息
func (c *ConsoleUI) PrintInfo(message string) {
	infoMsg := c.colorize(fmt.Sprintf("ℹ️  %s", message), ColorBlue)
	fmt.Println(infoMsg)
}

// Color 颜色常量
type Color string

const (
	ColorReset  Color = "\033[0m"
	ColorRed    Color = "\033[31m"
	ColorGreen  Color = "\033[32m"
	ColorYellow Color = "\033[33m"
	ColorBlue   Color = "\033[34m"
	ColorPurple Color = "\033[35m"
	ColorCyan   Color = "\033[36m"
	ColorWhite  Color = "\033[37m"
	ColorBold   Color = "\033[1m"
)

// colorize 给文本添加颜色
func (c *ConsoleUI) colorize(text string, color Color) string {
	if !c.colorEnabled || !c.isTerminalSupported() {
		return text
	}
	return string(color) + text + string(ColorReset)
}

// isTerminalSupported 检查终端是否支持颜色
func (c *ConsoleUI) isTerminalSupported() bool {
	// 简单检查，实际实现可能需要更复杂的逻辑
	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

// InstallProgressUI 安装进度UI
type InstallProgressUI struct {
	consoleUI *ConsoleUI
	tracker   *ProgressTracker
	stages    []string
}

// NewInstallProgressUI 创建安装进度UI
func NewInstallProgressUI() *InstallProgressUI {
	stages := []string{
		"检测系统",
		"下载文件",
		"验证文件",
		"解压文件",
		"配置安装",
		"完成安装",
	}

	tracker := NewProgressTracker(stages)
	consoleUI := NewConsoleUI(&ConsoleUIOptions{
		UpdateInterval: 100 * time.Millisecond,
		ColorEnabled:   true,
		Width:          40,
	})

	consoleUI.SetProgressTracker(tracker)

	return &InstallProgressUI{
		consoleUI: consoleUI,
		tracker:   tracker,
		stages:    stages,
	}
}

// Start 开始安装进度显示
func (i *InstallProgressUI) Start() {
	i.tracker.Start()
	i.consoleUI.Start()
}

// Stop 停止安装进度显示
func (i *InstallProgressUI) Stop() {
	i.tracker.Stop()
	i.consoleUI.Stop()
}

// SetStage 设置当前阶段
func (i *InstallProgressUI) SetStage(stage string) {
	i.tracker.SetStage(stage)
}

// SetProgress 设置进度
func (i *InstallProgressUI) SetProgress(progress float64) {
	i.tracker.SetProgress(progress)
}

// SetMessage 设置消息
func (i *InstallProgressUI) SetMessage(message string) {
	i.tracker.SetMessage(message)
}

// UpdateProgress 更新进度
func (i *InstallProgressUI) UpdateProgress(stage string, progress float64, message string) {
	i.tracker.UpdateProgress(stage, progress, message)
}

// PrintSuccess 打印成功消息
func (i *InstallProgressUI) PrintSuccess(message string) {
	i.consoleUI.PrintSuccess(message)
}

// PrintError 打印错误消息
func (i *InstallProgressUI) PrintError(message string) {
	i.consoleUI.PrintError(message)
}

// PrintWarning 打印警告消息
func (i *InstallProgressUI) PrintWarning(message string) {
	i.consoleUI.PrintWarning(message)
}

// PrintInfo 打印信息消息
func (i *InstallProgressUI) PrintInfo(message string) {
	i.consoleUI.PrintInfo(message)
}
