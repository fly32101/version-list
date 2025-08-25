package service

import (
	"fmt"
	"runtime"
	"version-list/internal/domain/model"
)

// SystemDetector 系统信息检测服务接口
type SystemDetector interface {
	DetectOS() (string, error)                                                 // 检测操作系统
	DetectArch() (string, error)                                               // 检测CPU架构
	GetDownloadURL(version, os, arch string) string                            // 构建官方下载URL
	GetDownloadURLWithMirror(version, os, arch, mirror string) string          // 使用镜像构建下载URL
	GetExpectedFilename(version, os, arch string) string                       // 获取预期文件名
	GetSystemInfo(version string) (*model.SystemInfo, error)                   // 获取完整系统信息
	GetSystemInfoWithMirror(version, mirror string) (*model.SystemInfo, error) // 使用镜像获取系统信息
	ValidateMirrorURL(mirrorBaseURL, version, os, arch string) error           // 验证镜像URL格式
}

// SystemDetectorImpl 系统检测服务实现
type SystemDetectorImpl struct{}

// NewSystemDetector 创建系统检测服务实例
func NewSystemDetector() SystemDetector {
	return &SystemDetectorImpl{}
}

// DetectOS 检测操作系统
func (s *SystemDetectorImpl) DetectOS() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "windows", nil
	case "linux":
		return "linux", nil
	case "darwin":
		return "darwin", nil
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// DetectArch 检测CPU架构
func (s *SystemDetectorImpl) DetectArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64", nil
	case "arm64":
		return "arm64", nil
	case "386":
		return "386", nil
	default:
		return "", fmt.Errorf("不支持的CPU架构: %s", runtime.GOARCH)
	}
}

// GetDownloadURL 构建Go官方下载URL
func (s *SystemDetectorImpl) GetDownloadURL(version, os, arch string) string {
	filename := s.GetExpectedFilename(version, os, arch)
	return fmt.Sprintf("https://golang.org/dl/%s", filename)
}

// GetDownloadURLWithMirror 使用镜像构建下载URL
func (s *SystemDetectorImpl) GetDownloadURLWithMirror(version, os, arch, mirror string) string {
	filename := s.GetExpectedFilename(version, os, arch)

	// 根据镜像名称构建URL
	switch mirror {
	case "official":
		return fmt.Sprintf("https://golang.org/dl/%s", filename)
	case "goproxy-cn":
		return fmt.Sprintf("https://goproxy.cn/golang/%s", filename)
	case "aliyun":
		return fmt.Sprintf("https://mirrors.aliyun.com/golang/%s", filename)
	case "tencent":
		return fmt.Sprintf("https://mirrors.cloud.tencent.com/golang/%s", filename)
	case "huawei":
		return fmt.Sprintf("https://mirrors.huaweicloud.com/golang/%s", filename)
	default:
		// 如果是自定义镜像URL（包含http://或https://），直接拼接
		if mirror != "" && (containsString(mirror, "http://") || containsString(mirror, "https://")) {
			// 确保镜像URL以/结尾
			if mirror[len(mirror)-1] != '/' {
				mirror += "/"
			}
			return fmt.Sprintf("%s%s", mirror, filename)
		}
		// 未知镜像名称或空镜像，默认使用官方源
		return s.GetDownloadURL(version, os, arch)
	}
}

// GetExpectedFilename 获取预期的下载文件名
func (s *SystemDetectorImpl) GetExpectedFilename(version, os, arch string) string {
	switch os {
	case "windows":
		if arch == "386" {
			return fmt.Sprintf("go%s.windows-386.zip", version)
		}
		return fmt.Sprintf("go%s.windows-%s.zip", version, arch)
	case "linux":
		return fmt.Sprintf("go%s.linux-%s.tar.gz", version, arch)
	case "darwin":
		return fmt.Sprintf("go%s.darwin-%s.tar.gz", version, arch)
	default:
		return fmt.Sprintf("go%s.%s-%s.tar.gz", version, os, arch)
	}
}

// GetSystemInfo 获取完整的系统信息
func (s *SystemDetectorImpl) GetSystemInfo(version string) (*model.SystemInfo, error) {
	return s.GetSystemInfoWithMirror(version, "official")
}

// GetSystemInfoWithMirror 使用镜像获取系统信息
func (s *SystemDetectorImpl) GetSystemInfoWithMirror(version, mirror string) (*model.SystemInfo, error) {
	os, err := s.DetectOS()
	if err != nil {
		return nil, fmt.Errorf("检测操作系统失败: %v", err)
	}

	arch, err := s.DetectArch()
	if err != nil {
		return nil, fmt.Errorf("检测CPU架构失败: %v", err)
	}

	filename := s.GetExpectedFilename(version, os, arch)
	url := s.GetDownloadURLWithMirror(version, os, arch, mirror)

	return &model.SystemInfo{
		OS:       os,
		Arch:     arch,
		Version:  version,
		Filename: filename,
		URL:      url,
		Mirror:   mirror,
	}, nil
}

// ValidateMirrorURL 验证镜像URL格式
func (s *SystemDetectorImpl) ValidateMirrorURL(mirrorBaseURL, version, os, arch string) error {
	if mirrorBaseURL == "" {
		return fmt.Errorf("镜像URL不能为空")
	}

	// 构建完整URL
	filename := s.GetExpectedFilename(version, os, arch)
	fullURL := s.GetDownloadURLWithMirror(version, os, arch, mirrorBaseURL)

	// 基本URL格式验证
	if len(fullURL) == 0 {
		return fmt.Errorf("构建的下载URL为空")
	}

	// 检查是否包含预期的文件名
	if !containsString(fullURL, filename) {
		return fmt.Errorf("构建的URL不包含预期的文件名: %s", filename)
	}

	// 检查URL是否以http或https开头
	if !containsString(fullURL, "http://") && !containsString(fullURL, "https://") {
		return fmt.Errorf("无效的URL格式: %s", fullURL)
	}

	return nil
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOfSubstring(s, substr) >= 0)))
}

// indexOfSubstring 查找子字符串的位置
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
