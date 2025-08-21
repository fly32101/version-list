package model

import (
	"time"
)

// GoVersion 表示一个Go版本
type GoVersion struct {
	Version         string          // 版本号，如 "1.21.0"
	Path            string          // 安装路径
	IsActive        bool            // 是否为当前激活版本
	Source          InstallSource   // 安装来源
	DownloadInfo    *DownloadInfo   // 下载信息（仅在线安装时有值）
	ExtractInfo     *ExtractInfo    // 解压信息
	ValidationInfo  *ValidationInfo // 验证信息
	SystemInfo      *SystemInfo     // 系统信息
	CreatedAt       time.Time       // 创建时间
	UpdatedAt       time.Time       // 更新时间
	LastUsedAt      *time.Time      // 最后使用时间
	InstallDuration time.Duration   // 安装耗时
	Tags            []string        // 标签（如 "stable", "beta", "rc"）
	Notes           string          // 备注信息
}

// Environment 表示环境变量配置
type Environment struct {
	GOROOT string // Go安装目录
	GOPATH string // 工作目录
	GOBIN  string // 二进制文件目录
}

// IsOnlineInstalled 检查是否为在线安装
func (v *GoVersion) IsOnlineInstalled() bool {
	return v.Source == SourceOnline && v.DownloadInfo != nil
}

// IsValid 检查版本是否有效
func (v *GoVersion) IsValid() bool {
	if v.ValidationInfo == nil {
		return false
	}
	return v.ValidationInfo.ChecksumValid &&
		v.ValidationInfo.ExecutableValid &&
		v.ValidationInfo.VersionValid
}

// GetDisplayName 获取显示名称
func (v *GoVersion) GetDisplayName() string {
	if len(v.Tags) > 0 {
		return v.Version + " (" + v.Tags[0] + ")"
	}
	return v.Version
}

// GetSizeInfo 获取大小信息
func (v *GoVersion) GetSizeInfo() (downloadSize, extractedSize int64) {
	if v.DownloadInfo != nil {
		downloadSize = v.DownloadInfo.Size
	}
	if v.ExtractInfo != nil {
		extractedSize = v.ExtractInfo.ExtractedSize
	}
	return
}

// MarkAsUsed 标记为已使用
func (v *GoVersion) MarkAsUsed() {
	now := time.Now()
	v.LastUsedAt = &now
	v.UpdatedAt = now
}

// AddTag 添加标签
func (v *GoVersion) AddTag(tag string) {
	for _, existingTag := range v.Tags {
		if existingTag == tag {
			return // 标签已存在
		}
	}
	v.Tags = append(v.Tags, tag)
	v.UpdatedAt = time.Now()
}

// RemoveTag 移除标签
func (v *GoVersion) RemoveTag(tag string) {
	for i, existingTag := range v.Tags {
		if existingTag == tag {
			v.Tags = append(v.Tags[:i], v.Tags[i+1:]...)
			v.UpdatedAt = time.Now()
			return
		}
	}
}

// HasTag 检查是否有指定标签
func (v *GoVersion) HasTag(tag string) bool {
	for _, existingTag := range v.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// VersionStatistics 版本统计信息
type VersionStatistics struct {
	TotalVersions      int           // 总版本数
	OnlineVersions     int           // 在线安装版本数
	LocalVersions      int           // 本地导入版本数
	ActiveVersion      string        // 当前激活版本
	TotalDiskUsage     int64         // 总磁盘使用量
	MostRecentlyUsed   string        // 最近使用的版本
	OldestVersion      string        // 最老的版本
	NewestVersion      string        // 最新的版本
	AverageInstallTime time.Duration // 平均安装时间
}
