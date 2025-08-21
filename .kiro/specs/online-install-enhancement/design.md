# Design Document

## Overview

本设计文档描述了如何增强现有的install命令，使其能够从Go官方网站在线下载、解压和配置指定版本的Go。设计遵循现有的领域驱动设计（DDD）架构，通过添加新的领域服务和基础设施组件来实现在线安装功能。

核心设计理念：
- 保持现有架构的一致性
- 添加可复用的下载和解压组件
- 提供清晰的进度反馈
- 确保操作的原子性和错误恢复

## Architecture

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Interface Layer                          │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │   CLI Command   │  │  Progress UI    │                  │
│  │   (install)     │  │   Component     │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Application Layer                          │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │           VersionAppService                             │ │
│  │  + InstallOnline(version, options)                     │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │ VersionService  │  │DownloadService  │                  │
│  │ (enhanced)      │  │    (new)        │                  │
│  └─────────────────┘  └─────────────────┘                  │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │ SystemDetector  │  │ArchiveExtractor │                  │
│  │    (new)        │  │     (new)       │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │   HTTP Client   │  │  File System    │                  │
│  │   (download)    │  │   Operations    │                  │
│  └─────────────────┘  └─────────────────┘                  │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │ Progress Tracker│  │ Archive Handler │                  │
│  │     (new)       │  │     (new)       │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

### 数据流

1. **用户输入** → CLI命令解析版本号和选项
2. **系统检测** → 自动检测OS和架构
3. **URL构建** → 根据检测结果构建下载URL
4. **下载管理** → 带进度的文件下载
5. **文件验证** → 校验下载文件完整性
6. **解压处理** → 解压到目标目录
7. **版本注册** → 将新版本添加到管理系统

## Components and Interfaces

### 1. Enhanced CLI Command

```go
// 增强的install命令，支持在线安装
var installCmd = &cobra.Command{
    Use:   "install [version] [flags]",
    Short: "在线安装指定版本的Go",
    Long:  `从Go官网在线下载并安装指定版本的Go`,
    Args:  cobra.ExactArgs(1),
    Run:   runInstallCommand,
}

// 命令行选项
type InstallOptions struct {
    Version     string // Go版本号
    InstallPath string // 自定义安装路径
    Force       bool   // 强制重新安装
    Timeout     int    // 下载超时时间（秒）
}
```

### 2. System Detection Service

```go
// SystemDetector 系统信息检测服务
type SystemDetector interface {
    DetectOS() (string, error)           // 检测操作系统
    DetectArch() (string, error)         // 检测CPU架构
    GetDownloadURL(version, os, arch string) string // 构建下载URL
    GetExpectedFilename(version, os, arch string) string // 获取预期文件名
}

// SystemInfo 系统信息
type SystemInfo struct {
    OS       string // windows, linux, darwin
    Arch     string // amd64, arm64, 386
    Version  string // Go版本号
    Filename string // 下载文件名
    URL      string // 下载URL
}
```

### 3. Download Service

```go
// DownloadService 下载服务接口
type DownloadService interface {
    Download(url, destPath string, progress ProgressCallback) error
    VerifyChecksum(filePath, expectedChecksum string) error
    GetFileSize(url string) (int64, error)
}

// ProgressCallback 进度回调函数
type ProgressCallback func(downloaded, total int64, speed float64)

// DownloadProgress 下载进度信息
type DownloadProgress struct {
    Downloaded int64   // 已下载字节数
    Total      int64   // 总字节数
    Speed      float64 // 下载速度 (bytes/sec)
    Percentage float64 // 完成百分比
    ETA        int64   // 预计剩余时间（秒）
}
```

### 4. Archive Extractor Service

```go
// ArchiveExtractor 压缩包解压服务
type ArchiveExtractor interface {
    Extract(archivePath, destPath string, progress ProgressCallback) error
    ValidateArchive(archivePath string) error
    GetArchiveType(filename string) ArchiveType
}

// ArchiveType 压缩包类型
type ArchiveType int

const (
    ArchiveTypeZip ArchiveType = iota
    ArchiveTypeTarGz
    ArchiveTypeMsi
)

// ExtractionProgress 解压进度信息
type ExtractionProgress struct {
    ProcessedFiles int     // 已处理文件数
    TotalFiles     int     // 总文件数
    CurrentFile    string  // 当前处理的文件
    Percentage     float64 // 完成百分比
}
```

### 5. Enhanced Version Service

```go
// 增强的VersionService，添加在线安装功能
type VersionService struct {
    versionRepo     repository.VersionRepository
    environmentRepo repository.EnvironmentRepository
    downloadService DownloadService
    extractor       ArchiveExtractor
    systemDetector  SystemDetector
}

// InstallOnline 在线安装Go版本
func (s *VersionService) InstallOnline(version string, options *InstallOptions) error
```

## Data Models

### 1. Enhanced GoVersion Model

```go
// 增强的GoVersion模型
type GoVersion struct {
    Version      string            // 版本号
    Path         string            // 安装路径
    IsActive     bool              // 是否激活
    Source       InstallSource     // 安装来源
    DownloadInfo *DownloadInfo     // 下载信息
    CreatedAt    time.Time         // 创建时间
    UpdatedAt    time.Time         // 更新时间
}

// InstallSource 安装来源
type InstallSource int

const (
    SourceLocal InstallSource = iota  // 本地导入
    SourceOnline                      // 在线下载
)

// DownloadInfo 下载信息
type DownloadInfo struct {
    URL          string    // 下载URL
    Filename     string    // 文件名
    Size         int64     // 文件大小
    Checksum     string    // 校验和
    DownloadedAt time.Time // 下载时间
}
```

### 2. Installation Context

```go
// InstallationContext 安装上下文
type InstallationContext struct {
    Version     string        // 目标版本
    SystemInfo  *SystemInfo   // 系统信息
    Paths       *InstallPaths // 安装路径
    Options     *InstallOptions // 安装选项
    TempDir     string        // 临时目录
}

// InstallPaths 安装路径配置
type InstallPaths struct {
    BaseDir     string // 基础目录
    VersionDir  string // 版本目录
    TempDir     string // 临时目录
    ArchiveFile string // 压缩包文件路径
}
```

## Error Handling

### 错误类型定义

```go
// 自定义错误类型
type InstallError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType int

const (
    ErrorTypeNetwork ErrorType = iota    // 网络错误
    ErrorTypeFileSystem                  // 文件系统错误
    ErrorTypeValidation                  // 验证错误
    ErrorTypeExtraction                  // 解压错误
    ErrorTypePermission                  // 权限错误
    ErrorTypeUnsupported                 // 不支持的系统
)
```

### 错误处理策略

1. **网络错误**：自动重试机制（最多3次），支持断点续传
2. **文件系统错误**：自动清理临时文件，提供详细错误信息
3. **验证错误**：重新下载或提示用户手动处理
4. **解压错误**：清理部分解压的文件，回滚操作
5. **权限错误**：提供权限修复建议
6. **不支持的系统**：列出支持的系统列表

### 回滚机制

```go
// RollbackManager 回滚管理器
type RollbackManager struct {
    operations []RollbackOperation
}

type RollbackOperation func() error

// 在安装过程中注册回滚操作
func (rm *RollbackManager) Register(op RollbackOperation)
func (rm *RollbackManager) Execute() error
```

## Testing Strategy

### 1. 单元测试

- **SystemDetector**: 测试各种操作系统和架构的检测
- **DownloadService**: 模拟网络请求，测试下载逻辑
- **ArchiveExtractor**: 测试各种压缩格式的解压
- **VersionService**: 测试完整的安装流程

### 2. 集成测试

- **端到端安装流程**: 使用测试版本进行完整安装测试
- **错误场景测试**: 模拟网络中断、磁盘空间不足等场景
- **并发安装测试**: 测试同时安装多个版本的情况

### 3. 性能测试

- **下载性能**: 测试大文件下载的性能和稳定性
- **解压性能**: 测试解压大型压缩包的性能
- **内存使用**: 监控安装过程中的内存使用情况

### 4. 兼容性测试

- **多平台测试**: Windows、Linux、macOS
- **多架构测试**: amd64、arm64、386
- **Go版本兼容性**: 测试不同Go版本的安装

### 测试工具和框架

```go
// 使用testify进行单元测试
func TestSystemDetector_DetectOS(t *testing.T) {
    detector := NewSystemDetector()
    os, err := detector.DetectOS()
    assert.NoError(t, err)
    assert.Contains(t, []string{"windows", "linux", "darwin"}, os)
}

// 使用httptest模拟HTTP服务器
func TestDownloadService_Download(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("test content"))
    }))
    defer server.Close()
    
    // 测试下载逻辑
}
```

### Mock和Stub策略

- **网络请求Mock**: 使用httpmock模拟Go官网API
- **文件系统Stub**: 使用afero进行文件系统操作的测试
- **进度回调Mock**: 验证进度报告的正确性