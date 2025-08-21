package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ArchiveType 压缩包类型
type ArchiveType int

const (
	ArchiveTypeZip ArchiveType = iota
	ArchiveTypeTarGz
	ArchiveTypeMsi
	ArchiveTypeUnknown
)

// ExtractionProgress 解压进度信息
type ExtractionProgress struct {
	ProcessedFiles int     // 已处理文件数
	TotalFiles     int     // 总文件数
	CurrentFile    string  // 当前处理的文件
	Percentage     float64 // 完成百分比
	ProcessedBytes int64   // 已处理字节数
	TotalBytes     int64   // 总字节数
}

// ExtractionProgressCallback 解压进度回调函数
type ExtractionProgressCallback func(progress ExtractionProgress)

// ArchiveExtractor 压缩包解压服务接口
type ArchiveExtractor interface {
	Extract(archivePath, destPath string, progress ExtractionProgressCallback) error
	ValidateArchive(archivePath string) error
	GetArchiveType(filename string) ArchiveType
	GetArchiveInfo(archivePath string) (*ArchiveInfo, error)
}

// ArchiveInfo 压缩包信息
type ArchiveInfo struct {
	Type      ArchiveType // 压缩包类型
	FileCount int         // 文件数量
	TotalSize int64       // 解压后总大小
	RootDir   string      // 根目录名称
}

// ArchiveExtractorImpl 压缩包解压服务实现
type ArchiveExtractorImpl struct {
	preservePermissions bool
	overwriteExisting   bool
}

// ArchiveExtractorOptions 解压选项
type ArchiveExtractorOptions struct {
	PreservePermissions bool // 保留文件权限
	OverwriteExisting   bool // 覆盖已存在的文件
}

// NewArchiveExtractor 创建压缩包解压服务实例
func NewArchiveExtractor(options *ArchiveExtractorOptions) ArchiveExtractor {
	if options == nil {
		options = &ArchiveExtractorOptions{
			PreservePermissions: true,
			OverwriteExisting:   true,
		}
	}

	return &ArchiveExtractorImpl{
		preservePermissions: options.PreservePermissions,
		overwriteExisting:   options.OverwriteExisting,
	}
}

// GetArchiveType 根据文件名判断压缩包类型
func (e *ArchiveExtractorImpl) GetArchiveType(filename string) ArchiveType {
	filename = strings.ToLower(filename)

	if strings.HasSuffix(filename, ".zip") {
		return ArchiveTypeZip
	}
	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		return ArchiveTypeTarGz
	}
	if strings.HasSuffix(filename, ".msi") {
		return ArchiveTypeMsi
	}

	return ArchiveTypeUnknown
}

// ValidateArchive 验证压缩包完整性
func (e *ArchiveExtractorImpl) ValidateArchive(archivePath string) error {
	archiveType := e.GetArchiveType(archivePath)

	switch archiveType {
	case ArchiveTypeZip:
		return e.validateZip(archivePath)
	case ArchiveTypeTarGz:
		return e.validateTarGz(archivePath)
	case ArchiveTypeMsi:
		return fmt.Errorf("MSI文件验证暂不支持")
	default:
		return fmt.Errorf("不支持的压缩包类型: %s", archivePath)
	}
}

// GetArchiveInfo 获取压缩包信息
func (e *ArchiveExtractorImpl) GetArchiveInfo(archivePath string) (*ArchiveInfo, error) {
	archiveType := e.GetArchiveType(archivePath)

	switch archiveType {
	case ArchiveTypeZip:
		return e.getZipInfo(archivePath)
	case ArchiveTypeTarGz:
		return e.getTarGzInfo(archivePath)
	default:
		return nil, fmt.Errorf("不支持的压缩包类型: %s", archivePath)
	}
}

// Extract 解压压缩包
func (e *ArchiveExtractorImpl) Extract(archivePath, destPath string, progress ExtractionProgressCallback) error {
	// 验证压缩包
	if err := e.ValidateArchive(archivePath); err != nil {
		return fmt.Errorf("压缩包验证失败: %v", err)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	archiveType := e.GetArchiveType(archivePath)

	switch archiveType {
	case ArchiveTypeZip:
		return e.extractZip(archivePath, destPath, progress)
	case ArchiveTypeTarGz:
		return e.extractTarGz(archivePath, destPath, progress)
	default:
		return fmt.Errorf("不支持的压缩包类型: %s", archivePath)
	}
}

// validateZip 验证ZIP文件
func (e *ArchiveExtractorImpl) validateZip(archivePath string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("打开ZIP文件失败: %v", err)
	}
	defer reader.Close()

	// 简单验证：尝试读取文件列表
	if len(reader.File) == 0 {
		return fmt.Errorf("ZIP文件为空")
	}

	return nil
}

// validateTarGz 验证tar.gz文件
func (e *ArchiveExtractorImpl) validateTarGz(archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("打开tar.gz文件失败: %v", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("创建gzip读取器失败: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// 尝试读取第一个文件头
	_, err = tarReader.Next()
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取tar文件头失败: %v", err)
	}

	return nil
}

// getZipInfo 获取ZIP文件信息
func (e *ArchiveExtractorImpl) getZipInfo(archivePath string) (*ArchiveInfo, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("打开ZIP文件失败: %v", err)
	}
	defer reader.Close()

	info := &ArchiveInfo{
		Type:      ArchiveTypeZip,
		FileCount: len(reader.File),
	}

	var rootDirs = make(map[string]bool)

	for _, file := range reader.File {
		info.TotalSize += int64(file.UncompressedSize64)

		// 找出根目录
		parts := strings.Split(file.Name, "/")
		if len(parts) > 0 && parts[0] != "" {
			rootDirs[parts[0]] = true
		}
	}

	// 如果只有一个根目录，记录它
	if len(rootDirs) == 1 {
		for rootDir := range rootDirs {
			info.RootDir = rootDir
		}
	}

	return info, nil
}

// getTarGzInfo 获取tar.gz文件信息
func (e *ArchiveExtractorImpl) getTarGzInfo(archivePath string) (*ArchiveInfo, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("打开tar.gz文件失败: %v", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("创建gzip读取器失败: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	info := &ArchiveInfo{
		Type: ArchiveTypeTarGz,
	}

	var rootDirs = make(map[string]bool)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取tar文件头失败: %v", err)
		}

		info.FileCount++
		info.TotalSize += header.Size

		// 找出根目录
		parts := strings.Split(header.Name, "/")
		if len(parts) > 0 && parts[0] != "" {
			rootDirs[parts[0]] = true
		}
	}

	// 如果只有一个根目录，记录它
	if len(rootDirs) == 1 {
		for rootDir := range rootDirs {
			info.RootDir = rootDir
		}
	}

	return info, nil
}

// extractZip 解压ZIP文件
func (e *ArchiveExtractorImpl) extractZip(archivePath, destPath string, progress ExtractionProgressCallback) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("打开ZIP文件失败: %v", err)
	}
	defer reader.Close()

	totalFiles := len(reader.File)
	var processedFiles int
	var processedBytes int64
	var totalBytes int64

	// 计算总字节数
	for _, file := range reader.File {
		totalBytes += int64(file.UncompressedSize64)
	}

	for _, file := range reader.File {
		if err := e.extractZipFile(file, destPath); err != nil {
			return fmt.Errorf("解压文件 %s 失败: %v", file.Name, err)
		}

		processedFiles++
		processedBytes += int64(file.UncompressedSize64)

		if progress != nil {
			progress(ExtractionProgress{
				ProcessedFiles: processedFiles,
				TotalFiles:     totalFiles,
				CurrentFile:    file.Name,
				Percentage:     float64(processedFiles) / float64(totalFiles) * 100,
				ProcessedBytes: processedBytes,
				TotalBytes:     totalBytes,
			})
		}
	}

	return nil
}

// extractZipFile 解压单个ZIP文件
func (e *ArchiveExtractorImpl) extractZipFile(file *zip.File, destPath string) error {
	// 构建目标路径
	targetPath := filepath.Join(destPath, file.Name)

	// 安全检查：防止路径遍历攻击
	if !strings.HasPrefix(targetPath, filepath.Clean(destPath)+string(os.PathSeparator)) {
		return fmt.Errorf("不安全的文件路径: %s", file.Name)
	}

	// 如果是目录
	if file.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, file.FileInfo().Mode())
	}

	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("创建父目录失败: %v", err)
	}

	// 检查是否覆盖已存在的文件
	if !e.overwriteExisting {
		if _, err := os.Stat(targetPath); err == nil {
			return nil // 跳过已存在的文件
		}
	}

	// 打开ZIP文件中的文件
	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开ZIP中的文件失败: %v", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	destFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer destFile.Close()

	// 复制文件内容
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	// 设置文件权限
	if e.preservePermissions {
		if err := os.Chmod(targetPath, file.FileInfo().Mode()); err != nil {
			// 在Windows上权限设置可能失败，不作为致命错误
			if runtime.GOOS != "windows" {
				return fmt.Errorf("设置文件权限失败: %v", err)
			}
		}
	}

	return nil
}

// extractTarGz 解压tar.gz文件
func (e *ArchiveExtractorImpl) extractTarGz(archivePath, destPath string, progress ExtractionProgressCallback) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("打开tar.gz文件失败: %v", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("创建gzip读取器失败: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// 先获取文件信息以计算进度
	info, err := e.getTarGzInfo(archivePath)
	if err != nil {
		return fmt.Errorf("获取tar.gz信息失败: %v", err)
	}

	var processedFiles int
	var processedBytes int64

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取tar文件头失败: %v", err)
		}

		if err := e.extractTarFile(tarReader, header, destPath); err != nil {
			return fmt.Errorf("解压文件 %s 失败: %v", header.Name, err)
		}

		processedFiles++
		processedBytes += header.Size

		if progress != nil {
			progress(ExtractionProgress{
				ProcessedFiles: processedFiles,
				TotalFiles:     info.FileCount,
				CurrentFile:    header.Name,
				Percentage:     float64(processedFiles) / float64(info.FileCount) * 100,
				ProcessedBytes: processedBytes,
				TotalBytes:     info.TotalSize,
			})
		}
	}

	return nil
}

// extractTarFile 解压单个tar文件
func (e *ArchiveExtractorImpl) extractTarFile(tarReader *tar.Reader, header *tar.Header, destPath string) error {
	// 构建目标路径
	targetPath := filepath.Join(destPath, header.Name)

	// 安全检查：防止路径遍历攻击
	if !strings.HasPrefix(targetPath, filepath.Clean(destPath)+string(os.PathSeparator)) {
		return fmt.Errorf("不安全的文件路径: %s", header.Name)
	}

	switch header.Typeflag {
	case tar.TypeDir:
		// 目录
		return os.MkdirAll(targetPath, os.FileMode(header.Mode))

	case tar.TypeReg:
		// 普通文件
		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("创建父目录失败: %v", err)
		}

		// 检查是否覆盖已存在的文件
		if !e.overwriteExisting {
			if _, err := os.Stat(targetPath); err == nil {
				return nil // 跳过已存在的文件
			}
		}

		// 创建目标文件
		destFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("创建目标文件失败: %v", err)
		}
		defer destFile.Close()

		// 复制文件内容
		if _, err := io.Copy(destFile, tarReader); err != nil {
			return fmt.Errorf("复制文件内容失败: %v", err)
		}

		// 设置文件权限
		if e.preservePermissions {
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				// 在Windows上权限设置可能失败，不作为致命错误
				if runtime.GOOS != "windows" {
					return fmt.Errorf("设置文件权限失败: %v", err)
				}
			}
		}

	case tar.TypeSymlink:
		// 符号链接（在Windows上可能不支持）
		if runtime.GOOS != "windows" {
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("创建父目录失败: %v", err)
			}

			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return fmt.Errorf("创建符号链接失败: %v", err)
			}
		}

	default:
		// 其他类型的文件，暂时跳过
		return nil
	}

	return nil
}
