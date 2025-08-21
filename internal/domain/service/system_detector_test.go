package service

import (
	"runtime"
	"strings"
	"testing"
)

func TestSystemDetectorImpl_DetectOS(t *testing.T) {
	detector := NewSystemDetector()

	os, err := detector.DetectOS()
	if err != nil {
		t.Fatalf("DetectOS() 返回错误: %v", err)
	}

	// 验证返回的操作系统是支持的类型之一
	supportedOS := []string{"windows", "linux", "darwin"}
	found := false
	for _, supported := range supportedOS {
		if os == supported {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("DetectOS() 返回不支持的操作系统: %s", os)
	}

	// 验证与runtime.GOOS一致
	if os != runtime.GOOS {
		t.Errorf("DetectOS() 返回 %s, 期望 %s", os, runtime.GOOS)
	}
}

func TestSystemDetectorImpl_DetectArch(t *testing.T) {
	detector := NewSystemDetector()

	arch, err := detector.DetectArch()
	if err != nil {
		t.Fatalf("DetectArch() 返回错误: %v", err)
	}

	// 验证返回的架构是支持的类型之一
	supportedArch := []string{"amd64", "arm64", "386"}
	found := false
	for _, supported := range supportedArch {
		if arch == supported {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("DetectArch() 返回不支持的架构: %s", arch)
	}

	// 验证与runtime.GOARCH一致
	if arch != runtime.GOARCH {
		t.Errorf("DetectArch() 返回 %s, 期望 %s", arch, runtime.GOARCH)
	}
}

func TestSystemDetectorImpl_GetExpectedFilename(t *testing.T) {
	detector := &SystemDetectorImpl{}

	testCases := []struct {
		version  string
		os       string
		arch     string
		expected string
	}{
		{"1.21.0", "windows", "amd64", "go1.21.0.windows-amd64.zip"},
		{"1.21.0", "windows", "386", "go1.21.0.windows-386.zip"},
		{"1.21.0", "linux", "amd64", "go1.21.0.linux-amd64.tar.gz"},
		{"1.21.0", "linux", "arm64", "go1.21.0.linux-arm64.tar.gz"},
		{"1.21.0", "darwin", "amd64", "go1.21.0.darwin-amd64.tar.gz"},
		{"1.21.0", "darwin", "arm64", "go1.21.0.darwin-arm64.tar.gz"},
		{"1.20.5", "linux", "386", "go1.20.5.linux-386.tar.gz"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := detector.GetExpectedFilename(tc.version, tc.os, tc.arch)
			if result != tc.expected {
				t.Errorf("GetExpectedFilename(%s, %s, %s) = %s, 期望 %s",
					tc.version, tc.os, tc.arch, result, tc.expected)
			}
		})
	}
}

func TestSystemDetectorImpl_GetDownloadURL(t *testing.T) {
	detector := &SystemDetectorImpl{}

	testCases := []struct {
		version string
		os      string
		arch    string
	}{
		{"1.21.0", "windows", "amd64"},
		{"1.21.0", "linux", "amd64"},
		{"1.21.0", "darwin", "arm64"},
	}

	for _, tc := range testCases {
		t.Run(tc.version+"_"+tc.os+"_"+tc.arch, func(t *testing.T) {
			url := detector.GetDownloadURL(tc.version, tc.os, tc.arch)

			// 验证URL格式
			if !strings.HasPrefix(url, "https://golang.org/dl/") {
				t.Errorf("GetDownloadURL() 返回的URL格式不正确: %s", url)
			}

			// 验证URL包含正确的文件名
			expectedFilename := detector.GetExpectedFilename(tc.version, tc.os, tc.arch)
			if !strings.HasSuffix(url, expectedFilename) {
				t.Errorf("GetDownloadURL() 返回的URL不包含正确的文件名: %s", url)
			}
		})
	}
}

func TestSystemDetectorImpl_GetSystemInfo(t *testing.T) {
	detector := NewSystemDetector()

	version := "1.21.0"
	systemInfo, err := detector.GetSystemInfo(version)
	if err != nil {
		t.Fatalf("GetSystemInfo() 返回错误: %v", err)
	}

	// 验证版本号
	if systemInfo.Version != version {
		t.Errorf("SystemInfo.Version = %s, 期望 %s", systemInfo.Version, version)
	}

	// 验证操作系统
	if systemInfo.OS == "" {
		t.Error("SystemInfo.OS 不能为空")
	}

	// 验证架构
	if systemInfo.Arch == "" {
		t.Error("SystemInfo.Arch 不能为空")
	}

	// 验证文件名
	if systemInfo.Filename == "" {
		t.Error("SystemInfo.Filename 不能为空")
	}

	// 验证URL
	if systemInfo.URL == "" {
		t.Error("SystemInfo.URL 不能为空")
	}

	if !strings.HasPrefix(systemInfo.URL, "https://golang.org/dl/") {
		t.Errorf("SystemInfo.URL 格式不正确: %s", systemInfo.URL)
	}

	// 验证文件名和URL的一致性
	if !strings.HasSuffix(systemInfo.URL, systemInfo.Filename) {
		t.Errorf("URL和文件名不匹配: URL=%s, Filename=%s", systemInfo.URL, systemInfo.Filename)
	}
}

func TestSystemDetectorImpl_GetSystemInfo_InvalidVersion(t *testing.T) {
	detector := NewSystemDetector()

	// 测试空版本号
	_, err := detector.GetSystemInfo("")
	if err != nil {
		// 空版本号应该能正常处理，只是生成的URL可能不正确
		// 这里我们主要测试不会崩溃
	}

	// 测试特殊字符版本号
	systemInfo, err := detector.GetSystemInfo("1.21.0-rc1")
	if err != nil {
		t.Fatalf("GetSystemInfo() 对于版本 '1.21.0-rc1' 返回错误: %v", err)
	}

	if systemInfo.Version != "1.21.0-rc1" {
		t.Errorf("SystemInfo.Version = %s, 期望 %s", systemInfo.Version, "1.21.0-rc1")
	}
}
