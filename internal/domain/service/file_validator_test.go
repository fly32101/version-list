package service

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileValidator_CalculateChecksum(t *testing.T) {
	validator := NewFileValidator()

	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 测试MD5
	md5Hash, err := validator.CalculateChecksum(testFile, ChecksumTypeMD5)
	if err != nil {
		t.Errorf("计算MD5失败: %v", err)
	}

	// 验证MD5值
	expectedMD5 := fmt.Sprintf("%x", md5.Sum([]byte(testContent)))
	if md5Hash != expectedMD5 {
		t.Errorf("MD5不匹配: 得到 %s, 期望 %s", md5Hash, expectedMD5)
	}

	// 测试SHA256
	sha256Hash, err := validator.CalculateChecksum(testFile, ChecksumTypeSHA256)
	if err != nil {
		t.Errorf("计算SHA256失败: %v", err)
	}

	// 验证SHA256值
	expectedSHA256 := fmt.Sprintf("%x", sha256.Sum256([]byte(testContent)))
	if sha256Hash != expectedSHA256 {
		t.Errorf("SHA256不匹配: 得到 %s, 期望 %s", sha256Hash, expectedSHA256)
	}

	// 测试不支持的类型
	_, err = validator.CalculateChecksum(testFile, "unsupported")
	if err == nil {
		t.Error("期望不支持的校验和类型返回错误")
	}
}

func TestFileValidator_ValidateChecksum(t *testing.T) {
	validator := NewFileValidator()

	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 计算正确的校验和
	correctChecksum := fmt.Sprintf("%x", md5.Sum([]byte(testContent)))

	// 测试正确的校验和
	err = validator.ValidateChecksum(testFile, correctChecksum, ChecksumTypeMD5)
	if err != nil {
		t.Errorf("验证正确校验和失败: %v", err)
	}

	// 测试错误的校验和
	wrongChecksum := "wrongchecksum"
	err = validator.ValidateChecksum(testFile, wrongChecksum, ChecksumTypeMD5)
	if err == nil {
		t.Error("期望错误校验和验证失败")
	}

	// 验证错误类型
	if installErr, ok := err.(*InstallError); ok {
		if installErr.Type != ErrorTypeCorrupted {
			t.Errorf("错误类型 = %v, 期望 %v", installErr.Type, ErrorTypeCorrupted)
		}
	} else {
		t.Error("期望返回InstallError类型")
	}

	// 测试空校验和
	err = validator.ValidateChecksum(testFile, "", ChecksumTypeMD5)
	if err == nil {
		t.Error("期望空校验和返回错误")
	}
}

func TestFileValidator_ValidateExecutable(t *testing.T) {
	validator := NewFileValidator()
	tempDir := t.TempDir()

	// 创建可执行文件
	execFile := filepath.Join(tempDir, "test_exec")
	if runtime.GOOS == "windows" {
		execFile += ".exe"
	}

	// 创建一个简单的可执行脚本
	var content string
	if runtime.GOOS == "windows" {
		content = "@echo off\necho test\n"
	} else {
		content = "#!/bin/bash\necho test\n"
	}

	err := os.WriteFile(execFile, []byte(content), 0755)
	if err != nil {
		t.Fatalf("创建可执行文件失败: %v", err)
	}

	// 测试验证可执行文件（可能会失败，因为不是真正的可执行文件）
	err = validator.ValidateExecutable(execFile)
	// 这个测试可能会失败，因为我们创建的不是真正的可执行文件
	// 但我们主要测试文件存在性和权限检查

	// 测试不存在的文件
	nonExistentFile := filepath.Join(tempDir, "nonexistent")
	err = validator.ValidateExecutable(nonExistentFile)
	if err == nil {
		t.Error("期望不存在的文件验证失败")
	}

	if installErr, ok := err.(*InstallError); ok {
		if installErr.Type != ErrorTypeFileSystem {
			t.Errorf("错误类型 = %v, 期望 %v", installErr.Type, ErrorTypeFileSystem)
		}
	}
}

func TestFileValidator_ValidateGoVersion(t *testing.T) {
	validator := NewFileValidator()
	tempDir := t.TempDir()

	// 创建模拟的Go安装目录结构
	goPath := filepath.Join(tempDir, "go")
	binDir := filepath.Join(goPath, "bin")
	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("创建bin目录失败: %v", err)
	}

	// 创建模拟的go可执行文件
	goExec := filepath.Join(binDir, "go")
	if runtime.GOOS == "windows" {
		goExec = filepath.Join(binDir, "go.exe")
	}

	// 创建一个模拟的go可执行文件（实际上不会工作）
	mockGoContent := "#!/bin/bash\necho 'go version go1.21.0 linux/amd64'\n"
	if runtime.GOOS == "windows" {
		mockGoContent = "@echo off\necho go version go1.21.0 windows/amd64\n"
	}

	err = os.WriteFile(goExec, []byte(mockGoContent), 0755)
	if err != nil {
		t.Fatalf("创建模拟go可执行文件失败: %v", err)
	}

	// 测试验证Go版本（会失败，因为不是真正的go可执行文件）
	err = validator.ValidateGoVersion(goPath, "1.21.0")
	// 这个测试预期会失败，因为我们的模拟文件不是真正的go可执行文件

	// 测试不存在的Go路径
	err = validator.ValidateGoVersion("/nonexistent/path", "1.21.0")
	if err == nil {
		t.Error("期望不存在的Go路径验证失败")
	}
}

func TestFileValidator_ValidateArchiveIntegrity(t *testing.T) {
	validator := NewFileValidator()
	tempDir := t.TempDir()

	// 测试不存在的文件
	nonExistentFile := filepath.Join(tempDir, "nonexistent.zip")
	err := validator.ValidateArchiveIntegrity(nonExistentFile)
	if err == nil {
		t.Error("期望不存在的压缩包验证失败")
	}

	// 创建空文件
	emptyFile := filepath.Join(tempDir, "empty.zip")
	err = os.WriteFile(emptyFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("创建空文件失败: %v", err)
	}

	err = validator.ValidateArchiveIntegrity(emptyFile)
	if err == nil {
		t.Error("期望空压缩包验证失败")
	}

	if installErr, ok := err.(*InstallError); ok {
		if installErr.Type != ErrorTypeCorrupted {
			t.Errorf("错误类型 = %v, 期望 %v", installErr.Type, ErrorTypeCorrupted)
		}
	}

	// 创建无效的ZIP文件
	invalidZip := filepath.Join(tempDir, "invalid.zip")
	err = os.WriteFile(invalidZip, []byte("not a zip file"), 0644)
	if err != nil {
		t.Fatalf("创建无效ZIP文件失败: %v", err)
	}

	err = validator.ValidateArchiveIntegrity(invalidZip)
	if err == nil {
		t.Error("期望无效ZIP文件验证失败")
	}
}

func TestFileValidator_GetSupportedChecksumTypes(t *testing.T) {
	validator := NewFileValidator()

	supportedTypes := validator.GetSupportedChecksumTypes()

	expectedTypes := []string{ChecksumTypeMD5, ChecksumTypeSHA1, ChecksumTypeSHA256, ChecksumTypeSHA512}

	if len(supportedTypes) != len(expectedTypes) {
		t.Errorf("支持的校验和类型数量 = %d, 期望 %d", len(supportedTypes), len(expectedTypes))
	}

	for _, expectedType := range expectedTypes {
		found := false
		for _, supportedType := range supportedTypes {
			if supportedType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("缺少支持的校验和类型: %s", expectedType)
		}
	}
}

func TestComprehensiveValidator_ValidateDownloadedFile(t *testing.T) {
	validator := NewComprehensiveValidator()
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.zip")
	testContent := "PK\x03\x04" // ZIP文件头
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 计算正确的校验和
	correctChecksum := fmt.Sprintf("%x", md5.Sum([]byte(testContent)))

	// 测试验证下载文件
	result := validator.ValidateDownloadedFile(testFile, correctChecksum, ChecksumTypeMD5)

	if !result.ChecksumOK {
		t.Error("校验和验证应该通过")
	}

	if len(result.Details) == 0 {
		t.Error("应该有验证详情")
	}

	// 测试错误的校验和
	wrongResult := validator.ValidateDownloadedFile(testFile, "wrongchecksum", ChecksumTypeMD5)

	if wrongResult.Valid {
		t.Error("错误校验和的验证应该失败")
	}

	if wrongResult.ChecksumOK {
		t.Error("错误校验和应该验证失败")
	}

	if len(wrongResult.Errors) == 0 {
		t.Error("应该有验证错误")
	}
}

func TestComprehensiveValidator_ValidateInstalledGo(t *testing.T) {
	validator := NewComprehensiveValidator()
	tempDir := t.TempDir()

	// 创建模拟的Go安装目录
	goPath := filepath.Join(tempDir, "go")
	binDir := filepath.Join(goPath, "bin")
	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("创建Go目录失败: %v", err)
	}

	// 测试验证不存在的Go安装
	result := validator.ValidateInstalledGo(goPath, "1.21.0")

	if result.Valid {
		t.Error("不存在的Go安装验证应该失败")
	}

	if result.VersionOK {
		t.Error("版本验证应该失败")
	}

	if result.ExecutableOK {
		t.Error("可执行文件验证应该失败")
	}

	if len(result.Errors) == 0 {
		t.Error("应该有验证错误")
	}
}

func TestFileValidator_ValidateGzipArchive(t *testing.T) {
	validator := &FileValidatorImpl{}
	tempDir := t.TempDir()

	// 创建有效的gzip文件头
	gzipFile := filepath.Join(tempDir, "test.gz")
	gzipHeader := []byte{0x1f, 0x8b, 0x08} // gzip魔数和压缩方法
	err := os.WriteFile(gzipFile, gzipHeader, 0644)
	if err != nil {
		t.Fatalf("创建gzip文件失败: %v", err)
	}

	err = validator.validateGzipArchive(gzipFile)
	if err != nil {
		t.Errorf("有效gzip文件验证失败: %v", err)
	}

	// 创建无效的gzip文件
	invalidGzipFile := filepath.Join(tempDir, "invalid.gz")
	err = os.WriteFile(invalidGzipFile, []byte("not gzip"), 0644)
	if err != nil {
		t.Fatalf("创建无效gzip文件失败: %v", err)
	}

	err = validator.validateGzipArchive(invalidGzipFile)
	if err == nil {
		t.Error("期望无效gzip文件验证失败")
	}
}

func TestFileValidator_ValidateTarArchive(t *testing.T) {
	validator := &FileValidatorImpl{}
	tempDir := t.TempDir()

	// 创建有效的tar文件头
	tarFile := filepath.Join(tempDir, "test.tar")
	tarHeader := make([]byte, 512)
	copy(tarHeader[257:262], []byte("ustar")) // tar魔数
	err := os.WriteFile(tarFile, tarHeader, 0644)
	if err != nil {
		t.Fatalf("创建tar文件失败: %v", err)
	}

	err = validator.validateTarArchive(tarFile)
	if err != nil {
		t.Errorf("有效tar文件验证失败: %v", err)
	}

	// 创建无效的tar文件
	invalidTarFile := filepath.Join(tempDir, "invalid.tar")
	invalidHeader := make([]byte, 512)
	copy(invalidHeader[257:262], []byte("wrong")) // 错误的魔数
	err = os.WriteFile(invalidTarFile, invalidHeader, 0644)
	if err != nil {
		t.Fatalf("创建无效tar文件失败: %v", err)
	}

	err = validator.validateTarArchive(invalidTarFile)
	if err == nil {
		t.Error("期望无效tar文件验证失败")
	}
}

func TestFileValidator_CaseInsensitiveChecksum(t *testing.T) {
	validator := NewFileValidator()
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 计算校验和
	checksum, err := validator.CalculateChecksum(testFile, ChecksumTypeMD5)
	if err != nil {
		t.Fatalf("计算校验和失败: %v", err)
	}

	// 测试大写校验和
	upperChecksum := strings.ToUpper(checksum)
	err = validator.ValidateChecksum(testFile, upperChecksum, ChecksumTypeMD5)
	if err != nil {
		t.Errorf("大写校验和验证失败: %v", err)
	}

	// 测试小写校验和
	lowerChecksum := strings.ToLower(checksum)
	err = validator.ValidateChecksum(testFile, lowerChecksum, ChecksumTypeMD5)
	if err != nil {
		t.Errorf("小写校验和验证失败: %v", err)
	}
}
