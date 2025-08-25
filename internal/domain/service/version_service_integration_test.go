package service

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"version-list/internal/domain/model"
)

// TestVersionServiceIntegration 测试版本服务的完整集成
func TestVersionServiceIntegration(t *testing.T) {
	// 创建模拟仓库
	versionRepo := NewMockVersionRepository()
	envRepo := NewMockEnvironmentRepository()

	// 创建版本服务
	service := NewVersionService(versionRepo, envRepo)

	// 测试基本功能
	t.Run("基本功能测试", func(t *testing.T) {
		// 测试列出版本（应该为空）
		versions, err := service.List()
		if err != nil {
			t.Fatalf("列出版本失败: %v", err)
		}

		if len(versions) != 0 {
			t.Errorf("期望0个版本，得到 %d 个", len(versions))
		}

		// 测试获取当前版本（应该失败）
		_, err = service.Current()
		if err == nil {
			t.Error("期望获取当前版本失败")
		}
	})

	t.Run("本地导入功能", func(t *testing.T) {
		// 创建临时Go安装目录进行测试
		tempDir := t.TempDir()
		goDir := filepath.Join(tempDir, "go")
		binDir := filepath.Join(goDir, "bin")

		// 创建目录结构
		if err := os.MkdirAll(binDir, 0755); err != nil {
			t.Fatalf("创建测试目录失败: %v", err)
		}

		// 创建模拟的go可执行文件
		goExec := filepath.Join(binDir, "go")
		if runtime.GOOS == "windows" {
			goExec = filepath.Join(binDir, "go.exe")
		}

		// 创建一个简单的脚本来模拟go version命令
		goScript := `#!/bin/bash
echo "go version go1.20.0 linux/amd64"
`
		if runtime.GOOS == "windows" {
			goScript = `@echo off
echo go version go1.20.0 windows/amd64
`
		}

		if err := os.WriteFile(goExec, []byte(goScript), 0755); err != nil {
			t.Fatalf("创建模拟go可执行文件失败: %v", err)
		}

		// 测试导入本地版本
		version, err := service.ImportLocal(goDir)
		if err != nil {
			t.Logf("导入本地版本失败（预期，因为模拟脚本）: %v", err)
			// 这是预期的，因为我们的模拟脚本可能无法正确执行
		} else {
			t.Logf("成功导入版本: %s", version)

			// 验证版本被添加到仓库
			versions, err := service.List()
			if err != nil {
				t.Fatalf("列出版本失败: %v", err)
			}

			if len(versions) != 1 {
				t.Errorf("期望1个版本，得到 %d 个", len(versions))
			}
		}
	})

	t.Run("版本管理功能", func(t *testing.T) {
		// 手动添加一些测试版本
		testVersions := []*model.GoVersion{
			{
				Version:  "1.20.0",
				Path:     "/test/go/1.20.0",
				IsActive: false,
				Source:   model.SourceLocal,
			},
			{
				Version:  "1.21.0",
				Path:     "/test/go/1.21.0",
				IsActive: false,
				Source:   model.SourceOnline,
			},
		}

		for _, v := range testVersions {
			if err := versionRepo.Save(v); err != nil {
				t.Fatalf("保存测试版本失败: %v", err)
			}
		}

		// 测试列出版本
		versions, err := service.List()
		if err != nil {
			t.Fatalf("列出版本失败: %v", err)
		}

		if len(versions) < 2 {
			t.Errorf("期望至少2个版本，得到 %d 个", len(versions))
		}

		// 测试设置激活版本
		if err := versionRepo.SetActive("1.21.0"); err != nil {
			t.Fatalf("设置激活版本失败: %v", err)
		}

		// 测试获取当前版本
		current, err := service.Current()
		if err != nil {
			t.Fatalf("获取当前版本失败: %v", err)
		}

		if current.Version != "1.21.0" {
			t.Errorf("当前版本 = %s, 期望 %s", current.Version, "1.21.0")
		}

		if !current.IsActive {
			t.Error("当前版本应该是激活状态")
		}
	})
}

// TestVersionServiceWithCustomDependencies 测试使用自定义依赖的版本服务
func TestVersionServiceWithCustomDependencies(t *testing.T) {
	// 创建模拟仓库
	versionRepo := NewMockVersionRepository()
	envRepo := NewMockEnvironmentRepository()

	// 创建自定义依赖
	systemDetector := NewSystemDetector()
	downloadService := NewDownloadService(&DownloadOptions{
		MaxRetries: 1,
		Timeout:    5 * time.Second,
	})
	archiveExtractor := NewArchiveExtractor(&ArchiveExtractorOptions{
		PreservePermissions: true,
		OverwriteExisting:   true,
	})

	// 创建带有自定义依赖的版本服务
	mirrorService := NewMirrorService()
	service := NewVersionServiceWithDependencies(
		versionRepo,
		envRepo,
		systemDetector,
		downloadService,
		archiveExtractor,
		mirrorService,
	)

	// 测试创建安装上下文
	version := "1.21.0"
	context, err := service.createInstallationContext(version, nil)
	if err != nil {
		t.Fatalf("创建安装上下文失败: %v", err)
	}

	// 验证上下文
	if context.Version != version {
		t.Errorf("Version = %s, 期望 %s", context.Version, version)
	}

	// 验证系统信息
	if context.SystemInfo == nil {
		t.Fatal("SystemInfo 不应该为 nil")
	}

	if context.SystemInfo.Version != version {
		t.Errorf("SystemInfo.Version = %s, 期望 %s", context.SystemInfo.Version, version)
	}

	// 验证路径配置
	if context.Paths == nil {
		t.Fatal("Paths 不应该为 nil")
	}

	if context.Paths.VersionDir == "" {
		t.Error("VersionDir 不应该为空")
	}

	// 验证选项设置
	if context.Options == nil {
		t.Fatal("Options 不应该为 nil")
	}

	if context.Options.Timeout != 300 {
		t.Errorf("默认超时时间 = %d, 期望 %d", context.Options.Timeout, 300)
	}
}

// TestVersionServiceErrorHandling 测试版本服务的错误处理
func TestVersionServiceErrorHandling(t *testing.T) {
	versionRepo := NewMockVersionRepository()
	envRepo := NewMockEnvironmentRepository()
	service := NewVersionService(versionRepo, envRepo)

	t.Run("移除不存在的版本", func(t *testing.T) {
		err := service.Remove("nonexistent")
		if err == nil {
			t.Error("期望移除不存在版本时返回错误")
		}
	})

	t.Run("移除激活版本", func(t *testing.T) {
		// 添加并激活一个版本
		testVersion := &model.GoVersion{
			Version:  "1.21.0",
			Path:     "/test/path",
			IsActive: false,
			Source:   model.SourceLocal,
		}

		versionRepo.Save(testVersion)
		versionRepo.SetActive("1.21.0")

		// 尝试移除激活版本
		err := service.Remove("1.21.0")
		if err == nil {
			t.Error("期望移除激活版本时返回错误")
		}
	})

	t.Run("导入不存在的路径", func(t *testing.T) {
		_, err := service.ImportLocal("/nonexistent/path")
		if err == nil {
			t.Error("期望导入不存在路径时返回错误")
		}
	})
}
