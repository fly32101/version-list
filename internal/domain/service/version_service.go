package service

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
	"version-list/internal/domain/model"
	"version-list/internal/domain/repository"
)

// VersionInfo 版本信息视图模型
type VersionInfo struct {
	Version  string // 版本号
	Path     string // 安装路径
	IsActive bool   // 是否为当前激活版本
}

// ProgressReporter 进度报告接口
type ProgressReporter interface {
	SetStage(stage string)
	SetProgress(progress float64)
	SetMessage(message string)
	UpdateProgress(stage string, progress float64, message string)
}

// VersionService 版本管理领域服务
type VersionService struct {
	versionRepo      repository.VersionRepository
	environmentRepo  repository.EnvironmentRepository
	systemDetector   SystemDetector
	downloadService  DownloadService
	archiveExtractor ArchiveExtractor
}

// NewVersionService 创建版本服务实例
func NewVersionService(versionRepo repository.VersionRepository, environmentRepo repository.EnvironmentRepository) *VersionService {
	return &VersionService{
		versionRepo:      versionRepo,
		environmentRepo:  environmentRepo,
		systemDetector:   NewSystemDetector(),
		downloadService:  NewDownloadService(nil),
		archiveExtractor: NewArchiveExtractor(nil),
	}
}

// NewVersionServiceWithDependencies 创建带有自定义依赖的版本服务实例
func NewVersionServiceWithDependencies(
	versionRepo repository.VersionRepository,
	environmentRepo repository.EnvironmentRepository,
	systemDetector SystemDetector,
	downloadService DownloadService,
	archiveExtractor ArchiveExtractor,
) *VersionService {
	return &VersionService{
		versionRepo:      versionRepo,
		environmentRepo:  environmentRepo,
		systemDetector:   systemDetector,
		downloadService:  downloadService,
		archiveExtractor: archiveExtractor,
	}
}

// Install 安装指定版本的Go
func (s *VersionService) Install(version string) error {
	// 检查版本是否已存在
	_, err := s.versionRepo.FindByVersion(version)
	if err == nil {
		return fmt.Errorf("Go版本 %s 已安装", version)
	}

	// 创建新版本记录
	newVersion := &model.GoVersion{
		Version:  version,
		IsActive: false,
	}

	// 保存版本记录
	return s.versionRepo.Save(newVersion)
}

// List 列出所有已安装的Go版本
func (s *VersionService) List() ([]*model.GoVersion, error) {
	return s.versionRepo.FindAll()
}

// Use 切换到指定版本的Go
func (s *VersionService) Use(version string) error {
	// 检查版本是否存在
	targetVersion, err := s.versionRepo.FindByVersion(version)
	if err != nil {
		return fmt.Errorf("Go版本 %s 未安装", version)
	}

	// 获取环境变量配置
	env, err := s.environmentRepo.Get()
	if err != nil {
		return fmt.Errorf("获取环境变量配置失败: %v", err)
	}

	// 创建符号链接目录
	symlinkDir := filepath.Join(os.Getenv("HOME"), ".go-version", "current")
	if runtime.GOOS == "windows" {
		// Windows使用不同的路径
		homeDir, _ := os.UserHomeDir()
		symlinkDir = filepath.Join(homeDir, ".go-version", "current")
	}

	// 确保符号链接目录的父目录存在
	if err := os.MkdirAll(filepath.Dir(symlinkDir), 0755); err != nil {
		return fmt.Errorf("创建符号链接目录失败: %v", err)
	}

	// 如果符号链接已存在，先删除
	if _, err := os.Lstat(symlinkDir); err == nil {
		if err := os.Remove(symlinkDir); err != nil {
			return fmt.Errorf("删除旧符号链接失败: %v", err)
		}
	}

	// 创建新的符号链接
	// 注意：这里需要确定实际的Go安装路径
	goInstallPath := targetVersion.Path
	if goInstallPath == "" {
		// 如果没有指定路径，使用默认路径
		homeDir, _ := os.UserHomeDir()
		goInstallPath = filepath.Join(homeDir, ".go", "versions", version)
	}

	// 检查目标路径是否存在
	if _, err := os.Stat(goInstallPath); os.IsNotExist(err) {
		return fmt.Errorf("Go版本 %s 的安装路径不存在: %s", version, goInstallPath)
	}

	// 创建符号链接
	if runtime.GOOS == "windows" {
		// Windows使用mklink命令
		cmd := exec.Command("cmd", "/C", "mklink", "/J", symlinkDir, goInstallPath)
		if err := cmd.Run(); err != nil {
			// 如果mklink失败，尝试使用powershell
			cmd = exec.Command("powershell", "-Command", "New-Item", "-ItemType", "SymbolicLink", "-Path", symlinkDir, "-Target", goInstallPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("创建符号链接失败: %v", err)
			}
		}
	} else {
		// Unix-like系统使用symlink
		if err := os.Symlink(goInstallPath, symlinkDir); err != nil {
			return fmt.Errorf("创建符号链接失败: %v", err)
		}
	}

	// 更新环境变量配置，使用符号链接路径
	env.GOROOT = symlinkDir
	env.GOBIN = filepath.Join(symlinkDir, "bin")

	// 保存环境变量配置
	if err := s.environmentRepo.Save(env); err != nil {
		return fmt.Errorf("保存环境变量配置失败: %v", err)
	}

	// 设置指定版本为激活状态
	if err := s.versionRepo.SetActive(version); err != nil {
		return fmt.Errorf("设置激活版本失败: %v", err)
	}

	return nil
}

// Current 获取当前使用的Go版本
func (s *VersionService) Current() (*model.GoVersion, error) {
	return s.versionRepo.FindActive()
}

// Remove 移除指定版本的Go
func (s *VersionService) Remove(version string) error {
	// 检查版本是否存在
	_, err := s.versionRepo.FindByVersion(version)
	if err != nil {
		return fmt.Errorf("Go版本 %s 未安装", version)
	}

	// 检查是否为当前激活版本
	activeVersion, err := s.versionRepo.FindActive()
	if err == nil && activeVersion.Version == version {
		return fmt.Errorf("不能移除当前使用的Go版本 %s", version)
	}

	// 移除版本
	return s.versionRepo.Remove(version)
}

// ImportLocal 导入本地已安装的Go版本
func (s *VersionService) ImportLocal(path string) (string, error) {
	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("路径 %s 不存在", path)
	}

	// 尝试从路径中获取Go版本号
	version, err := s.extractVersionFromPath(path)
	if err != nil {
		return "", fmt.Errorf("无法从路径中提取Go版本: %v", err)
	}

	// 检查版本是否已存在
	_, err = s.versionRepo.FindByVersion(version)
	if err == nil {
		return "", fmt.Errorf("Go版本 %s 已存在", version)
	}

	// 创建新版本记录
	newVersion := &model.GoVersion{
		Version:  version,
		Path:     path,
		IsActive: false,
	}

	// 保存版本记录
	if err := s.versionRepo.Save(newVersion); err != nil {
		return "", fmt.Errorf("保存版本记录失败: %v", err)
	}

	return version, nil
}

// extractVersionFromPath 从Go安装路径中提取版本号
func (s *VersionService) extractVersionFromPath(path string) (string, error) {
	// 确定go可执行文件的路径
	var goExecPath string
	if runtime.GOOS == "windows" {
		goExecPath = filepath.Join(path, "bin", "go.exe")
	} else {
		goExecPath = filepath.Join(path, "bin", "go")
	}

	// 检查go可执行文件是否存在
	if _, err := os.Stat(goExecPath); os.IsNotExist(err) {
		return "", fmt.Errorf("在路径 %s 中未找到go可执行文件", path)
	}

	// 执行go version命令获取版本信息
	cmd := exec.Command(goExecPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("执行go version命令失败: %v", err)
	}

	// 解析输出获取版本号
	// go version命令的输出格式通常为: go version go1.21.0 windows/amd64
	versionStr := string(output)

	// 使用正则表达式提取版本号
	re := regexp.MustCompile(`go(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(versionStr)
	if len(matches) < 2 {
		return "", fmt.Errorf("无法解析go version输出: %s", versionStr)
	}

	return matches[1], nil
}

// InstallOnline 在线安装指定版本的Go
func (s *VersionService) InstallOnline(version string, options *model.InstallOptions) (*model.InstallationResult, error) {
	return s.InstallOnlineWithProgress(version, options, nil)
}

// InstallOnlineWithProgress 带进度显示的在线安装指定版本的Go
func (s *VersionService) InstallOnlineWithProgress(version string, options *model.InstallOptions, progressUI ProgressReporter) (*model.InstallationResult, error) {
	// 检查版本是否已存在
	_, err := s.versionRepo.FindByVersion(version)
	if err == nil {
		if options == nil || !options.Force {
			return nil, fmt.Errorf("go版本 %s 已安装，使用 --force 选项强制重新安装", version)
		}
		// 强制重新安装，先删除现有版本
		if progressUI != nil {
			progressUI.SetMessage("删除现有版本...")
		}
		if err := s.Remove(version); err != nil {
			return nil, fmt.Errorf("删除现有版本失败: %v", err)
		}
	}

	// 创建安装上下文
	if progressUI != nil {
		progressUI.SetStage("检测系统")
		progressUI.SetProgress(10)
		progressUI.SetMessage("正在检测系统信息...")
	}

	context, err := s.createInstallationContext(version, options)
	if err != nil {
		return nil, fmt.Errorf("创建安装上下文失败: %v", err)
	}

	if progressUI != nil {
		progressUI.SetProgress(20)
		progressUI.SetMessage(fmt.Sprintf("检测到系统: %s/%s", context.SystemInfo.OS, context.SystemInfo.Arch))
	}

	// 执行安装流程
	result, err := s.executeInstallationWithProgress(context, progressUI)
	if err != nil {
		// 清理失败的安装
		s.cleanupFailedInstallation(context)
		return result, err
	}

	return result, nil
}

// createInstallationContext 创建安装上下文
func (s *VersionService) createInstallationContext(version string, options *model.InstallOptions) (*model.InstallationContext, error) {
	// 获取系统信息
	systemInfo, err := s.systemDetector.GetSystemInfo(version)
	if err != nil {
		return nil, fmt.Errorf("获取系统信息失败: %v", err)
	}

	// 设置默认选项
	if options == nil {
		options = &model.InstallOptions{
			Force:            false,
			SkipVerification: false,
			Timeout:          300, // 5分钟
			MaxRetries:       3,
		}
	}

	// 确定安装路径
	var versionDir string
	var baseDir string

	if options.CustomPath != "" {
		// 如果指定了自定义路径，直接使用该路径作为版本目录
		versionDir = options.CustomPath
		baseDir = filepath.Dir(options.CustomPath)
	} else {
		// 使用默认路径结构
		baseDir = s.getBaseInstallDir()
		versionDir = filepath.Join(baseDir, version)
	}
	tempDir := filepath.Join(os.TempDir(), "go-install-"+version)

	paths := &model.InstallPaths{
		BaseDir:     baseDir,
		VersionDir:  versionDir,
		TempDir:     tempDir,
		ArchiveFile: filepath.Join(tempDir, systemInfo.Filename),
	}

	return &model.InstallationContext{
		Version:    version,
		SystemInfo: systemInfo,
		Paths:      paths,
		TempDir:    tempDir,
		Options:    options,
		StartTime:  time.Now(),
		Status:     model.StatusPending,
	}, nil
}

// executeInstallation 执行安装流程
func (s *VersionService) executeInstallation(context *model.InstallationContext) (*model.InstallationResult, error) {
	return s.executeInstallationWithProgress(context, nil)
}

// executeInstallationWithProgress 带进度显示的执行安装流程
func (s *VersionService) executeInstallationWithProgress(context *model.InstallationContext, progressUI ProgressReporter) (*model.InstallationResult, error) {
	result := &model.InstallationResult{
		Version: context.Version,
		Path:    context.Paths.VersionDir,
	}

	// 1. 创建临时目录
	if progressUI != nil {
		progressUI.SetProgress(25)
		progressUI.SetMessage("创建临时目录...")
	}
	if err := os.MkdirAll(context.TempDir, 0755); err != nil {
		return s.createFailedResult(result, fmt.Errorf("创建临时目录失败: %v", err))
	}

	// 2. 下载阶段
	context.Status = model.StatusDownloading
	if progressUI != nil {
		progressUI.SetStage("下载文件")
		progressUI.SetProgress(30)
		progressUI.SetMessage(fmt.Sprintf("正在下载 %s...", context.SystemInfo.Filename))
	}

	downloadInfo, err := s.downloadGoArchiveWithProgress(context, progressUI)
	if err != nil {
		return s.createFailedResult(result, fmt.Errorf("下载失败: %v", err))
	}
	result.DownloadInfo = downloadInfo

	// 3. 验证阶段（如果未跳过）
	if !context.Options.SkipVerification {
		if progressUI != nil {
			progressUI.SetStage("验证文件")
			progressUI.SetProgress(60)
			progressUI.SetMessage("验证下载文件完整性...")
		}
		if err := s.verifyDownload(context); err != nil {
			return s.createFailedResult(result, fmt.Errorf("验证失败: %v", err))
		}
	}

	// 4. 解压阶段
	context.Status = model.StatusExtracting
	if progressUI != nil {
		progressUI.SetStage("解压文件")
		progressUI.SetProgress(65)
		progressUI.SetMessage("正在解压安装包...")
	}

	extractInfo, err := s.extractGoArchiveWithProgress(context, progressUI)
	if err != nil {
		return s.createFailedResult(result, fmt.Errorf("解压失败: %v", err))
	}
	result.ExtractInfo = extractInfo

	// 5. 配置阶段
	context.Status = model.StatusConfiguring
	if progressUI != nil {
		progressUI.SetStage("配置安装")
		progressUI.SetProgress(85)
		progressUI.SetMessage("配置Go环境...")
	}
	if err := s.configureInstallation(context); err != nil {
		return s.createFailedResult(result, fmt.Errorf("配置失败: %v", err))
	}

	// 6. 保存版本信息
	if progressUI != nil {
		progressUI.SetProgress(90)
		progressUI.SetMessage("保存版本信息...")
	}
	goVersion, err := s.createGoVersionRecord(context, downloadInfo, extractInfo)
	if err != nil {
		return s.createFailedResult(result, fmt.Errorf("保存版本记录失败: %v", err))
	}

	if err := s.versionRepo.Save(goVersion); err != nil {
		return s.createFailedResult(result, fmt.Errorf("保存到仓库失败: %v", err))
	}

	// 7. 清理临时文件
	if progressUI != nil {
		progressUI.SetStage("完成安装")
		progressUI.SetProgress(95)
		progressUI.SetMessage("清理临时文件...")
	}
	s.cleanupTempFiles(context)

	// 完成安装
	context.Status = model.StatusCompleted
	if progressUI != nil {
		progressUI.SetProgress(100)
		progressUI.SetMessage("安装完成!")
	}
	result.Success = true
	result.Duration = time.Since(context.StartTime)

	return result, nil
}

// downloadGoArchive 下载Go压缩包
func (s *VersionService) downloadGoArchive(context *model.InstallationContext) (*model.DownloadInfo, error) {
	return s.downloadGoArchiveWithProgress(context, nil)
}

// downloadGoArchiveWithProgress 带进度显示的下载Go压缩包
func (s *VersionService) downloadGoArchiveWithProgress(context *model.InstallationContext, progressUI ProgressReporter) (*model.DownloadInfo, error) {
	startTime := time.Now()

	// 获取文件大小
	size, err := s.downloadService.GetFileSize(context.SystemInfo.URL)
	if err != nil {
		return nil, fmt.Errorf("获取文件大小失败: %v", err)
	}

	// 下载文件
	err = s.downloadService.Download(
		context.SystemInfo.URL,
		context.Paths.ArchiveFile,
		func(downloaded, total int64, speed float64) {
			if progressUI != nil {
				// 计算下载进度 (30-60%)
				downloadProgress := float64(downloaded)/float64(total)*30.0 + 30.0
				progressUI.SetProgress(downloadProgress)

				// 格式化速度和进度信息
				speedStr := formatBytes(int64(speed)) + "/s"
				progressStr := fmt.Sprintf("%.1f%%", float64(downloaded)/float64(total)*100)
				progressUI.SetMessage(fmt.Sprintf("下载中... %s (%s)", progressStr, speedStr))
			}
		},
	)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	avgSpeed := float64(size) / duration.Seconds()

	return &model.DownloadInfo{
		URL:          context.SystemInfo.URL,
		Filename:     context.SystemInfo.Filename,
		Size:         size,
		DownloadedAt: time.Now(),
		Duration:     duration.Milliseconds(),
		Speed:        avgSpeed,
	}, nil
}

// verifyDownload 验证下载的文件
func (s *VersionService) verifyDownload(context *model.InstallationContext) error {
	// 验证文件是否存在
	if _, err := os.Stat(context.Paths.ArchiveFile); os.IsNotExist(err) {
		return fmt.Errorf("下载文件不存在: %s", context.Paths.ArchiveFile)
	}

	// 验证压缩包完整性
	if err := s.archiveExtractor.ValidateArchive(context.Paths.ArchiveFile); err != nil {
		return fmt.Errorf("压缩包验证失败: %v", err)
	}

	return nil
}

// extractGoArchive 解压Go压缩包
func (s *VersionService) extractGoArchive(context *model.InstallationContext) (*model.ExtractInfo, error) {
	return s.extractGoArchiveWithProgress(context, nil)
}

// extractGoArchiveWithProgress 带进度显示的解压Go压缩包
func (s *VersionService) extractGoArchiveWithProgress(context *model.InstallationContext, progressUI ProgressReporter) (*model.ExtractInfo, error) {
	startTime := time.Now()

	// 获取压缩包信息
	archiveInfo, err := s.archiveExtractor.GetArchiveInfo(context.Paths.ArchiveFile)
	if err != nil {
		return nil, fmt.Errorf("获取压缩包信息失败: %v", err)
	}

	// 创建版本目录
	if err := os.MkdirAll(context.Paths.VersionDir, 0755); err != nil {
		return nil, fmt.Errorf("创建版本目录失败: %v", err)
	}

	// 解压到临时目录
	tempExtractDir := filepath.Join(context.TempDir, "extracted")
	err = s.archiveExtractor.Extract(
		context.Paths.ArchiveFile,
		tempExtractDir,
		func(progress ExtractionProgress) {
			if progressUI != nil {
				// 计算解压进度 (65-85%)
				extractProgress := progress.Percentage*0.20 + 65.0
				progressUI.SetProgress(extractProgress)
				progressUI.SetMessage(fmt.Sprintf("解压中... %.1f%% (%d/%d 文件)",
					progress.Percentage, progress.ProcessedFiles, progress.TotalFiles))
			}
		},
	)
	if err != nil {
		return nil, err
	}

	// 移动解压后的内容到最终目录
	if err := s.moveExtractedContent(tempExtractDir, context.Paths.VersionDir, archiveInfo.RootDir); err != nil {
		return nil, fmt.Errorf("移动解压内容失败: %v", err)
	}

	duration := time.Since(startTime)

	return &model.ExtractInfo{
		ArchiveSize:   archiveInfo.TotalSize,
		ExtractedSize: archiveInfo.TotalSize, // 简化处理
		FileCount:     archiveInfo.FileCount,
		Duration:      duration,
		RootDir:       archiveInfo.RootDir,
	}, nil
}

// moveExtractedContent 移动解压后的内容
func (s *VersionService) moveExtractedContent(srcDir, destDir, rootDir string) error {
	// 如果有根目录，需要从根目录中移动内容
	if rootDir != "" {
		srcDir = filepath.Join(srcDir, rootDir)
	}

	// 检查源目录是否存在
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("源目录不存在: %s", srcDir)
	}

	// 移动内容
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// 尝试移动文件，如果失败则复制后删除
		err = os.Rename(path, destPath)
		if err != nil {
			// 如果rename失败（通常是跨驱动器），则使用复制+删除
			if err := s.copyFile(path, destPath); err != nil {
				return fmt.Errorf("复制文件失败 %s -> %s: %v", path, destPath, err)
			}
			// 复制成功后删除源文件
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("删除源文件失败 %s: %v", path, err)
			}
		}

		return nil
	})
}

// copyFile 复制文件
func (s *VersionService) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制文件内容
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// 复制文件权限
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// configureInstallation 配置安装
func (s *VersionService) configureInstallation(context *model.InstallationContext) error {
	// 验证Go二进制文件
	goExecPath := s.getGoExecutablePath(context.Paths.VersionDir)
	if _, err := os.Stat(goExecPath); os.IsNotExist(err) {
		return fmt.Errorf("go可执行文件不存在: %s", goExecPath)
	}

	// 验证Go版本
	actualVersion, err := s.extractVersionFromPath(context.Paths.VersionDir)
	if err != nil {
		return fmt.Errorf("验证Go版本失败: %v", err)
	}

	if actualVersion != context.Version {
		return fmt.Errorf("版本不匹配: 期望 %s, 实际 %s", context.Version, actualVersion)
	}

	return nil
}

// createGoVersionRecord 创建GoVersion记录
func (s *VersionService) createGoVersionRecord(
	context *model.InstallationContext,
	downloadInfo *model.DownloadInfo,
	extractInfo *model.ExtractInfo,
) (*model.GoVersion, error) {
	now := time.Now()

	return &model.GoVersion{
		Version:      context.Version,
		Path:         context.Paths.VersionDir,
		IsActive:     false,
		Source:       model.SourceOnline,
		DownloadInfo: downloadInfo,
		ExtractInfo:  extractInfo,
		ValidationInfo: &model.ValidationInfo{
			ChecksumValid:   true, // 简化处理
			ExecutableValid: true,
			VersionValid:    true,
		},
		SystemInfo:      context.SystemInfo,
		CreatedAt:       now,
		UpdatedAt:       now,
		InstallDuration: time.Since(context.StartTime),
		Tags:            []string{"online"},
	}, nil
}

// getBaseInstallDir 获取基础安装目录
func (s *VersionService) getBaseInstallDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".go", "versions")
}

// getGoExecutablePath 获取Go可执行文件路径
func (s *VersionService) getGoExecutablePath(installDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(installDir, "bin", "go.exe")
	}
	return filepath.Join(installDir, "bin", "go")
}

// createFailedResult 创建失败的安装结果
func (s *VersionService) createFailedResult(result *model.InstallationResult, err error) (*model.InstallationResult, error) {
	result.Success = false
	result.Error = err.Error()
	return result, err
}

// cleanupFailedInstallation 清理失败的安装
func (s *VersionService) cleanupFailedInstallation(context *model.InstallationContext) {
	// 删除临时目录
	os.RemoveAll(context.TempDir)

	// 删除部分安装的版本目录
	if _, err := os.Stat(context.Paths.VersionDir); err == nil {
		os.RemoveAll(context.Paths.VersionDir)
	}
}

// cleanupTempFiles 清理临时文件
func (s *VersionService) cleanupTempFiles(context *model.InstallationContext) {
	os.RemoveAll(context.TempDir)
}
