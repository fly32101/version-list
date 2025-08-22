package service

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FileValidator 文件验证器接口
type FileValidator interface {
	ValidateChecksum(filePath, expectedChecksum, checksumType string) error
	CalculateChecksum(filePath, checksumType string) (string, error)
	ValidateExecutable(executablePath string) error
	ValidateGoVersion(goPath, expectedVersion string) error
	ValidateArchiveIntegrity(archivePath string) error
	GetSupportedChecksumTypes() []string
}

// FileValidatorImpl 文件验证器实现
type FileValidatorImpl struct {
	supportedTypes []string
}

// ChecksumType 校验和类型常量
const (
	ChecksumTypeMD5    = "md5"
	ChecksumTypeSHA1   = "sha1"
	ChecksumTypeSHA256 = "sha256"
	ChecksumTypeSHA512 = "sha512"
)

// NewFileValidator 创建文件验证器
func NewFileValidator() FileValidator {
	return &FileValidatorImpl{
		supportedTypes: []string{
			ChecksumTypeMD5,
			ChecksumTypeSHA1,
			ChecksumTypeSHA256,
			ChecksumTypeSHA512,
		},
	}
}

// GetSupportedChecksumTypes 获取支持的校验和类型
func (v *FileValidatorImpl) GetSupportedChecksumTypes() []string {
	return v.supportedTypes
}

// ValidateChecksum 验证文件校验和
func (v *FileValidatorImpl) ValidateChecksum(filePath, expectedChecksum, checksumType string) error {
	if expectedChecksum == "" {
		return NewInstallError(ErrorTypeValidation, "期望的校验和不能为空", nil)
	}

	actualChecksum, err := v.CalculateChecksum(filePath, checksumType)
	if err != nil {
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("计算文件校验和失败: %v", err), err)
	}

	// 比较校验和（不区分大小写）
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return NewInstallError(ErrorTypeCorrupted,
			fmt.Sprintf("文件校验和不匹配: 期望 %s, 实际 %s", expectedChecksum, actualChecksum), nil).
			WithContext("file", filePath).
			WithContext("checksum_type", checksumType).
			WithContext("expected", expectedChecksum).
			WithContext("actual", actualChecksum)
	}

	return nil
}

// CalculateChecksum 计算文件校验和
func (v *FileValidatorImpl) CalculateChecksum(filePath, checksumType string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	var hasher hash.Hash
	switch strings.ToLower(checksumType) {
	case ChecksumTypeMD5:
		hasher = md5.New()
	case ChecksumTypeSHA1:
		hasher = sha1.New()
	case ChecksumTypeSHA256:
		hasher = sha256.New()
	case ChecksumTypeSHA512:
		hasher = sha512.New()
	default:
		return "", fmt.Errorf("不支持的校验和类型: %s", checksumType)
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("读取文件内容失败: %v", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// ValidateExecutable 验证可执行文件
func (v *FileValidatorImpl) ValidateExecutable(executablePath string) error {
	// 检查文件是否存在
	fileInfo, err := os.Stat(executablePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewInstallError(ErrorTypeFileSystem,
				fmt.Sprintf("可执行文件不存在: %s", executablePath), err)
		}
		return NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("访问可执行文件失败: %v", err), err)
	}

	// 检查是否为普通文件
	if !fileInfo.Mode().IsRegular() {
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("路径不是普通文件: %s", executablePath), nil)
	}

	// 检查文件权限（Unix系统）
	if runtime.GOOS != "windows" {
		if fileInfo.Mode()&0111 == 0 {
			return NewInstallError(ErrorTypePermission,
				fmt.Sprintf("文件没有执行权限: %s", executablePath), nil)
		}
	}

	// 尝试执行文件验证其可执行性
	cmd := exec.Command(executablePath, "version")
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		// 某些情况下可执行文件可能需要特定参数或环境
		// 这里我们只检查基本的可执行性
		if exitErr, ok := err.(*exec.ExitError); ok {
			// 如果程序运行了但返回非零退出码，说明文件是可执行的
			if exitErr.ExitCode() != 0 {
				return nil // 文件可执行，只是参数不对
			}
		}
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("文件不可执行或执行失败: %v", err), err)
	}

	return nil
}

// ValidateGoVersion 验证Go版本
func (v *FileValidatorImpl) ValidateGoVersion(goPath, expectedVersion string) error {
	// 构建go可执行文件路径
	goExecPath := filepath.Join(goPath, "bin", "go")
	if runtime.GOOS == "windows" {
		goExecPath = filepath.Join(goPath, "bin", "go.exe")
	}

	// 先验证可执行文件
	if err := v.ValidateExecutable(goExecPath); err != nil {
		return err
	}

	// 执行go version命令
	cmd := exec.Command(goExecPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("执行go version命令失败: %v", err), err)
	}

	versionOutput := string(output)

	// 解析版本信息
	// go version输出格式: go version go1.21.0 windows/amd64
	if !strings.Contains(versionOutput, "go version") {
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("无效的go version输出: %s", versionOutput), nil)
	}

	// 检查版本号
	expectedVersionStr := "go" + expectedVersion
	if !strings.Contains(versionOutput, expectedVersionStr) {
		return NewInstallError(ErrorTypeValidation,
			fmt.Sprintf("Go版本不匹配: 期望 %s, 输出 %s", expectedVersion, versionOutput), nil).
			WithContext("expected_version", expectedVersion).
			WithContext("actual_output", versionOutput).
			WithContext("go_path", goPath)
	}

	return nil
}

// ValidateArchiveIntegrity 验证压缩包完整性
func (v *FileValidatorImpl) ValidateArchiveIntegrity(archivePath string) error {
	// 检查文件是否存在
	fileInfo, err := os.Stat(archivePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewInstallError(ErrorTypeFileSystem,
				fmt.Sprintf("压缩包文件不存在: %s", archivePath), err)
		}
		return NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("访问压缩包文件失败: %v", err), err)
	}

	// 检查文件大小
	if fileInfo.Size() == 0 {
		return NewInstallError(ErrorTypeCorrupted,
			fmt.Sprintf("压缩包文件为空: %s", archivePath), nil)
	}

	// 根据文件扩展名验证压缩包格式
	ext := strings.ToLower(filepath.Ext(archivePath))
	switch ext {
	case ".zip":
		return v.validateZipArchive(archivePath)
	case ".gz":
		if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
			return v.validateTarGzArchive(archivePath)
		}
		return v.validateGzipArchive(archivePath)
	case ".tar":
		return v.validateTarArchive(archivePath)
	default:
		// 对于未知格式，只进行基本检查
		return v.validateGenericArchive(archivePath)
	}
}

// validateZipArchive 验证ZIP压缩包
func (v *FileValidatorImpl) validateZipArchive(archivePath string) error {
	// 尝试打开ZIP文件
	extractor := NewArchiveExtractor(nil)
	return extractor.ValidateArchive(archivePath)
}

// validateTarGzArchive 验证tar.gz压缩包
func (v *FileValidatorImpl) validateTarGzArchive(archivePath string) error {
	// 尝试打开tar.gz文件
	extractor := NewArchiveExtractor(nil)
	return extractor.ValidateArchive(archivePath)
}

// validateGzipArchive 验证gzip压缩包
func (v *FileValidatorImpl) validateGzipArchive(archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("打开gzip文件失败: %v", err), err)
	}
	defer file.Close()

	// 读取gzip文件头
	header := make([]byte, 3)
	_, err = file.Read(header)
	if err != nil {
		return NewInstallError(ErrorTypeCorrupted,
			fmt.Sprintf("读取gzip文件头失败: %v", err), err)
	}

	// 检查gzip魔数
	if header[0] != 0x1f || header[1] != 0x8b {
		return NewInstallError(ErrorTypeCorrupted,
			fmt.Sprintf("无效的gzip文件头: %s", archivePath), nil)
	}

	return nil
}

// validateTarArchive 验证tar压缩包
func (v *FileValidatorImpl) validateTarArchive(archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("打开tar文件失败: %v", err), err)
	}
	defer file.Close()

	// 读取tar文件头（512字节）
	header := make([]byte, 512)
	_, err = file.Read(header)
	if err != nil {
		return NewInstallError(ErrorTypeCorrupted,
			fmt.Sprintf("读取tar文件头失败: %v", err), err)
	}

	// 检查tar文件头的魔数（位置257-262）
	magic := string(header[257:262])
	if magic != "ustar" {
		return NewInstallError(ErrorTypeCorrupted,
			fmt.Sprintf("无效的tar文件头: %s", archivePath), nil)
	}

	return nil
}

// validateGenericArchive 验证通用压缩包
func (v *FileValidatorImpl) validateGenericArchive(archivePath string) error {
	// 对于未知格式，只检查文件是否可读
	file, err := os.Open(archivePath)
	if err != nil {
		return NewInstallError(ErrorTypeFileSystem,
			fmt.Sprintf("无法打开压缩包文件: %v", err), err)
	}
	defer file.Close()

	// 尝试读取前几个字节
	buffer := make([]byte, 1024)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return NewInstallError(ErrorTypeCorrupted,
			fmt.Sprintf("读取压缩包文件失败: %v", err), err)
	}

	return nil
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid        bool              // 是否有效
	ChecksumOK   bool              // 校验和是否正确
	ExecutableOK bool              // 可执行文件是否有效
	VersionOK    bool              // 版本是否正确
	ArchiveOK    bool              // 压缩包是否完整
	Errors       []*InstallError   // 验证错误列表
	Details      map[string]string // 验证详情
}

// ComprehensiveValidator 综合验证器
type ComprehensiveValidator struct {
	fileValidator FileValidator
}

// NewComprehensiveValidator 创建综合验证器
func NewComprehensiveValidator() *ComprehensiveValidator {
	return &ComprehensiveValidator{
		fileValidator: NewFileValidator(),
	}
}

// ValidateDownloadedFile 验证下载的文件
func (cv *ComprehensiveValidator) ValidateDownloadedFile(filePath, expectedChecksum, checksumType string) *ValidationResult {
	result := &ValidationResult{
		Valid:   true,
		Details: make(map[string]string),
		Errors:  make([]*InstallError, 0),
	}

	// 验证校验和
	if expectedChecksum != "" {
		err := cv.fileValidator.ValidateChecksum(filePath, expectedChecksum, checksumType)
		if err != nil {
			result.Valid = false
			result.ChecksumOK = false
			if installErr, ok := err.(*InstallError); ok {
				result.Errors = append(result.Errors, installErr)
			} else {
				result.Errors = append(result.Errors, NewInstallError(ErrorTypeValidation, err.Error(), err))
			}
			result.Details["checksum_error"] = err.Error()
		} else {
			result.ChecksumOK = true
			result.Details["checksum_status"] = "验证通过"
		}
	}

	// 验证压缩包完整性
	err := cv.fileValidator.ValidateArchiveIntegrity(filePath)
	if err != nil {
		result.Valid = false
		result.ArchiveOK = false
		if installErr, ok := err.(*InstallError); ok {
			result.Errors = append(result.Errors, installErr)
		} else {
			result.Errors = append(result.Errors, NewInstallError(ErrorTypeValidation, err.Error(), err))
		}
		result.Details["archive_error"] = err.Error()
	} else {
		result.ArchiveOK = true
		result.Details["archive_status"] = "压缩包完整"
	}

	return result
}

// ValidateInstalledGo 验证已安装的Go
func (cv *ComprehensiveValidator) ValidateInstalledGo(goPath, expectedVersion string) *ValidationResult {
	result := &ValidationResult{
		Valid:   true,
		Details: make(map[string]string),
		Errors:  make([]*InstallError, 0),
	}

	// 验证Go版本
	err := cv.fileValidator.ValidateGoVersion(goPath, expectedVersion)
	if err != nil {
		result.Valid = false
		result.VersionOK = false
		result.ExecutableOK = false
		if installErr, ok := err.(*InstallError); ok {
			result.Errors = append(result.Errors, installErr)
		} else {
			result.Errors = append(result.Errors, NewInstallError(ErrorTypeValidation, err.Error(), err))
		}
		result.Details["version_error"] = err.Error()
	} else {
		result.VersionOK = true
		result.ExecutableOK = true
		result.Details["version_status"] = "版本验证通过"
	}

	return result
}
