package model

import (
	"time"
)

// SystemInfo 系统信息
type SystemInfo struct {
	OS       string // 操作系统: windows, linux, darwin
	Arch     string // CPU架构: amd64, arm64, 386
	Version  string // Go版本号
	Filename string // 下载文件名
	URL      string // 下载URL
	Mirror   string // 使用的镜像源名称
}

// InstallSource 安装来源
type InstallSource int

const (
	SourceLocal  InstallSource = iota // 本地导入
	SourceOnline                      // 在线下载
)

// DownloadInfo 下载信息
type DownloadInfo struct {
	URL          string    // 下载URL
	Filename     string    // 文件名
	Size         int64     // 文件大小
	Checksum     string    // 校验和
	ChecksumType string    // 校验和类型 (sha256, md5等)
	DownloadedAt time.Time // 下载时间
	Duration     int64     // 下载耗时（毫秒）
	Speed        float64   // 平均下载速度（字节/秒）
}

// InstallationContext 安装上下文
type InstallationContext struct {
	Version    string          // 目标版本
	SystemInfo *SystemInfo     // 系统信息
	Paths      *InstallPaths   // 安装路径
	TempDir    string          // 临时目录
	Options    *InstallOptions // 安装选项
	StartTime  time.Time       // 安装开始时间
	Status     InstallStatus   // 安装状态
}

// InstallPaths 安装路径配置
type InstallPaths struct {
	BaseDir     string // 基础目录
	VersionDir  string // 版本目录
	TempDir     string // 临时目录
	ArchiveFile string // 压缩包文件路径
}

// InstallOptions 安装选项
type InstallOptions struct {
	Force            bool   // 强制重新安装
	CustomPath       string // 自定义安装路径
	SkipVerification bool   // 跳过文件验证
	Timeout          int    // 超时时间（秒）
	MaxRetries       int    // 最大重试次数
	Mirror           string // 指定镜像源名称
	AutoMirror       bool   // 自动选择最快镜像
}

// InstallStatus 安装状态
type InstallStatus int

const (
	StatusPending     InstallStatus = iota // 等待中
	StatusDownloading                      // 下载中
	StatusExtracting                       // 解压中
	StatusConfiguring                      // 配置中
	StatusCompleted                        // 已完成
	StatusFailed                           // 失败
	StatusCancelled                        // 已取消
)

// String 返回安装状态的字符串表示
func (s InstallStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusDownloading:
		return "downloading"
	case StatusExtracting:
		return "extracting"
	case StatusConfiguring:
		return "configuring"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// InstallationResult 安装结果
type InstallationResult struct {
	Success      bool          // 是否成功
	Version      string        // 安装的版本
	Path         string        // 安装路径
	Duration     time.Duration // 安装耗时
	Error        string        // 错误信息（如果失败）
	DownloadInfo *DownloadInfo // 下载信息
	ExtractInfo  *ExtractInfo  // 解压信息
}

// ExtractInfo 解压信息
type ExtractInfo struct {
	ArchiveSize   int64         // 压缩包大小
	ExtractedSize int64         // 解压后大小
	FileCount     int           // 文件数量
	Duration      time.Duration // 解压耗时
	RootDir       string        // 根目录名称
}

// ValidationInfo 验证信息
type ValidationInfo struct {
	ChecksumValid   bool   // 校验和是否有效
	ExecutableValid bool   // 可执行文件是否有效
	VersionValid    bool   // 版本信息是否正确
	Error           string // 验证错误信息
}
