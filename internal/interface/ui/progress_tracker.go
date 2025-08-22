package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressTracker 进度跟踪器
type ProgressTracker struct {
	mu             sync.RWMutex
	currentStage   string
	stages         []string
	stageIndex     int
	progress       float64
	message        string
	startTime      time.Time
	lastUpdate     time.Time
	lastRender     time.Time
	isActive       bool
	showSpinner    bool
	spinnerIndex   int
	updateInterval time.Duration // 更新间隔优化
}

// ProgressStage 进度阶段
type ProgressStage struct {
	Name        string
	Description string
	Weight      float64 // 阶段权重（用于计算总进度）
}

// ProgressInfo 进度信息
type ProgressInfo struct {
	Stage       string        // 当前阶段
	Progress    float64       // 当前进度 (0-100)
	Message     string        // 状态消息
	ElapsedTime time.Duration // 已用时间
	ETA         time.Duration // 预计剩余时间
}

// NewProgressTracker 创建进度跟踪器
func NewProgressTracker(stages []string) *ProgressTracker {
	return &ProgressTracker{
		stages:         stages,
		startTime:      time.Now(),
		lastUpdate:     time.Now(),
		lastRender:     time.Now(),
		showSpinner:    true,
		updateInterval: 100 * time.Millisecond, // 优化更新频率
	}
}

// Start 开始进度跟踪
func (p *ProgressTracker) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isActive = true
	p.startTime = time.Now()
	p.lastUpdate = time.Now()

	if len(p.stages) > 0 {
		p.currentStage = p.stages[0]
		p.stageIndex = 0
	}
}

// Stop 停止进度跟踪
func (p *ProgressTracker) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isActive = false
	p.progress = 100.0
}

// SetStage 设置当前阶段
func (p *ProgressTracker) SetStage(stage string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentStage = stage
	p.lastUpdate = time.Now()

	// 查找阶段索引
	for i, s := range p.stages {
		if s == stage {
			p.stageIndex = i
			break
		}
	}
}

// SetProgress 设置当前进度
func (p *ProgressTracker) SetProgress(progress float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	p.progress = progress
	p.lastUpdate = time.Now()
}

// SetMessage 设置状态消息
func (p *ProgressTracker) SetMessage(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.message = message
	p.lastUpdate = time.Now()
}

// UpdateProgress 更新进度和消息（带频率限制）
func (p *ProgressTracker) UpdateProgress(stage string, progress float64, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 频率限制：避免过于频繁的更新
	now := time.Now()
	if now.Sub(p.lastUpdate) < p.updateInterval {
		return
	}

	p.currentStage = stage
	p.progress = progress
	p.message = message
	p.lastUpdate = now

	// 更新阶段索引
	for i, s := range p.stages {
		if s == stage {
			p.stageIndex = i
			break
		}
	}
}

// GetProgress 获取当前进度信息
func (p *ProgressTracker) GetProgress() ProgressInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	elapsed := time.Since(p.startTime)
	var eta time.Duration

	if p.progress > 0 && p.progress < 100 {
		totalEstimated := elapsed * time.Duration(100/p.progress)
		eta = totalEstimated - elapsed
	}

	return ProgressInfo{
		Stage:       p.currentStage,
		Progress:    p.progress,
		Message:     p.message,
		ElapsedTime: elapsed,
		ETA:         eta,
	}
}

// IsActive 检查是否处于活跃状态
func (p *ProgressTracker) IsActive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isActive
}

// RenderProgressBar 渲染进度条
func (p *ProgressTracker) RenderProgressBar(width int) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if width <= 0 {
		width = 50
	}

	filled := int(p.progress * float64(width) / 100)
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	percentage := fmt.Sprintf("%.1f%%", p.progress)

	return fmt.Sprintf("[%s] %s", bar, percentage)
}

// RenderSpinner 渲染旋转器
func (p *ProgressTracker) RenderSpinner() string {
	if !p.showSpinner {
		return ""
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinners[p.spinnerIndex%len(spinners)]
	p.spinnerIndex++

	return spinner
}

// RenderFullStatus 渲染完整状态（带渲染频率优化）
func (p *ProgressTracker) RenderFullStatus(width int) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 渲染频率限制
	now := time.Now()
	if now.Sub(p.lastRender) < p.updateInterval {
		return "" // 返回空字符串表示不需要重新渲染
	}
	p.lastRender = now

	info := p.GetProgress()

	var parts []string

	// 阶段信息
	if info.Stage != "" {
		parts = append(parts, fmt.Sprintf("阶段: %s", info.Stage))
	}

	// 进度条
	progressBar := p.RenderProgressBar(width)
	parts = append(parts, progressBar)

	// 消息
	if info.Message != "" {
		parts = append(parts, info.Message)
	}

	// 时间信息
	elapsed := info.ElapsedTime.Truncate(time.Second)
	timeInfo := fmt.Sprintf("已用时间: %v", elapsed)

	if info.ETA > 0 {
		eta := info.ETA.Truncate(time.Second)
		timeInfo += fmt.Sprintf(", 预计剩余: %v", eta)
	}

	parts = append(parts, timeInfo)

	return strings.Join(parts, "\n")
}

// ShouldRender 检查是否需要重新渲染
func (p *ProgressTracker) ShouldRender() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return time.Since(p.lastRender) >= p.updateInterval
}

// CalculateOverallProgress 计算总体进度
func (p *ProgressTracker) CalculateOverallProgress(stageWeights map[string]float64) float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.stages) == 0 {
		return p.progress
	}

	totalWeight := 0.0
	completedWeight := 0.0

	for i, stage := range p.stages {
		weight := stageWeights[stage]
		if weight == 0 {
			weight = 1.0 // 默认权重
		}

		totalWeight += weight

		if i < p.stageIndex {
			// 已完成的阶段
			completedWeight += weight
		} else if i == p.stageIndex {
			// 当前阶段
			completedWeight += weight * (p.progress / 100.0)
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return (completedWeight / totalWeight) * 100.0
}
