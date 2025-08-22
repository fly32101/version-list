package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"version-list/internal/domain/service"
)

func main() {
	fmt.Println("=== 简单文件验证示例 ===")

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "simple_validation")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建文件验证器
	validator := service.NewFileValidator()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, File Validation!"

	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		fmt.Printf("创建测试文件失败: %v\n", err)
		return
	}

	fmt.Printf("创建测试文件: %s\n", testFile)
	fmt.Printf("文件内容: %s\n", testContent)

	// 计算MD5校验和
	checksum, err := validator.CalculateChecksum(testFile, service.ChecksumTypeMD5)
	if err != nil {
		fmt.Printf("计算校验和失败: %v\n", err)
		return
	}

	fmt.Printf("MD5校验和: %s\n", checksum)

	// 验证校验和
	err = validator.ValidateChecksum(testFile, checksum, service.ChecksumTypeMD5)
	if err != nil {
		fmt.Printf("校验和验证失败: %v\n", err)
	} else {
		fmt.Println("校验和验证通过 ✓")
	}

	// 测试错误的校验和
	wrongChecksum := "wrongchecksum"
	err = validator.ValidateChecksum(testFile, wrongChecksum, service.ChecksumTypeMD5)
	if err != nil {
		fmt.Printf("错误校验和验证失败（预期）: %v\n", err)
	}

	// 显示支持的校验和类型
	fmt.Println("\n支持的校验和类型:")
	for i, checksumType := range validator.GetSupportedChecksumTypes() {
		fmt.Printf("  %d. %s\n", i+1, checksumType)
	}

	// 演示综合验证器
	fmt.Println("\n=== 综合验证器示例 ===")
	comprehensiveValidator := service.NewComprehensiveValidator()

	// 计算期望的校验和
	expectedChecksum := fmt.Sprintf("%x", md5.Sum([]byte(testContent)))

	// 执行综合验证
	result := comprehensiveValidator.ValidateDownloadedFile(testFile, expectedChecksum, service.ChecksumTypeMD5)

	fmt.Printf("验证结果:\n")
	fmt.Printf("  整体有效性: %t\n", result.Valid)
	fmt.Printf("  校验和验证: %t\n", result.ChecksumOK)
	fmt.Printf("  压缩包验证: %t\n", result.ArchiveOK)

	if len(result.Details) > 0 {
		fmt.Println("  验证详情:")
		for key, value := range result.Details {
			fmt.Printf("    %s: %s\n", key, value)
		}
	}

	fmt.Println("\n=== 示例完成 ===")
}
