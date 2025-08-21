package service

import (
	"fmt"
	"runtime"
	"version-list/internal/domain/model"
)

// SystemDetector 系统信息检测服务接口
type SystemDetector interface {
	DetectOS() (string, error)                               // 检测操作系统
	DetectArch() (string, error)                             // 检测CPU架构
	GetDownloadURL(version, os, arch string) string          // 构建下载URL
	GetExpectedFilename(version, os, arch string) string     // 获取预期文件名
	GetSystemInfo(version string) (*model.SystemInfo, error) // 获取完整系统信息
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
	os, err := s.DetectOS()
	if err != nil {
		return nil, fmt.Errorf("检测操作系统失败: %v", err)
	}

	arch, err := s.DetectArch()
	if err != nil {
		return nil, fmt.Errorf("检测CPU架构失败: %v", err)
	}

	filename := s.GetExpectedFilename(version, os, arch)
	url := s.GetDownloadURL(version, os, arch)

	return &model.SystemInfo{
		OS:       os,
		Arch:     arch,
		Version:  version,
		Filename: filename,
		URL:      url,
	}, nil
}
