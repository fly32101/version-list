package main

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"version-list/internal/domain/service"
)

func main() {
	fmt.Println("=== 文件完整性验证示例 ===")

	// 演示校验和计算和验证
	demonstrateChecksumValidation()

	// 演示压缩包验证
	demonstrateArchiveValidation()

	// 演示综合验证器
	demonstrateComprehensiveValidation()

	// 演示错误处理
	demonstrateValidationErrors()

	fmt.Println("\n=== 示例完成 ===")
}

func demonstrateChecksumValidation() {
	fmt.Println("\n--- 校验和验证示例 ---")

	validator := service.NewFileValidator()

	// 创建临时目录和测试文件
	tempDir, err := os.MkdirTemp("", "validation_demo")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test_file.txt")
	testContent := "这是一个测试文件，用于演示校验和验证功能。"

	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		fmt.Printf("创建测试文件失败: %v\n", err)
		return
	}

	fmt.Printf("创建测试文件: %s\n", testFile)
	fmt.Printf("文件内容: %s\n", testContent)

	// 演示不同类型的校验和计算
	checksumTypes := []string{
		service.ChecksumTypeMD5,
		service.ChecksumTypeSHA1,
		service.ChecksumTypeSHA256,
		service.ChecksumTypeSHA512,
	}

	fmt.Println("\n计算不同类型的校验和:")
	checksums := make(map[string]string)

	for _, checksumType := range checksumTypes {
		checksum, err := validator.CalculateChecksum(testFile, checksumType)
		if err != nil {
			fmt.Printf("计算 %s 校验和失败: %v\n", checksumType, err)
			continue
		}

		checksums[checksumType] = checksum
		fmt.Printf("  %s: %s\n", checksumType, checksum)
	}

	// 演示校验和验证
	fmt.Println("\n验证校验和:")
	for checksumType, checksum := range checksums {
		err := validator.ValidateChecksum(testFile, checksum, checksumType)
		if err != nil {
			fmt.Printf("  %s 验证失败: %v\n", checksumType, err)
		} else {
			fmt.Printf("  %s 验证通过 ✓\n", checksumType)
		}
	}

	// 演示错误的校验和
	fmt.Println("\n测试错误的校验和:")
	wrongChecksum := "0123456789abcdef"
	err = validator.ValidateChecksum(testFile, wrongChecksum, service.ChecksumTypeMD5)
	if err != nil {
		fmt.Printf("  错误校验和验证失败（预期）: %v\n", err)
	}
}

func demonstrateArchiveValidation() {
	fmt.Println("\n--- 压缩包验证示例 ---")

	validator := service.NewFileValidator()
	tempDir, err := os.MkdirTemp("", "archive_demo")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建模拟的压缩包文件
	testCases := []struct {
		filename string
		content  []byte
		valid    bool
		desc     string
	}{
		{
			filename: "valid.zip",
			content:  []byte("PK\x03\x04"), // ZIP文件头
			valid:    false,                // 实际上不是完整的ZIP，但有正确的头部
			desc:     "ZIP文件头",
		},
		{
			filename: "valid.gz",
			content:  []byte{0x1f, 0x8b, 0x08}, // gzip魔数
			valid:    true,
			desc:     "gzip文件头",
		},
		{
			filename: "invalid.zip",
			content:  []byte("not a zip file"),
			valid:    false,
			desc:     "无效ZIP文件",
		},
		{
			filename: "empty.tar.gz",
			content:  []byte{},
			valid:    false,
			desc:     "空文件",
		},
	}

	fmt.Println("测试不同类型的压缩包:")
	for _, tc := range testCases {
		filePath := filepath.Join(tempDir, tc.filename)
		err := os.WriteFile(filePath, tc.content, 0644)
		if err != nil {
			fmt.Printf("创建文件 %s 失败: %v\n", tc.filename, err)
			continue
		}

		err = validator.ValidateArchiveIntegrity(filePath)
		if err != nil {
			fmt.Printf("  %s (%s): 验证失败 - %v\n", tc.filename, tc.desc, err)
		} else {
			fmt.Printf("  %s (%s): 验证通过 ✓\n", tc.filename, tc.desc)
		}
	}
}

func demonstrateComprehensiveValidation() {
	fmt.Println("\n--- 综合验证器示例 ---")

	validator := service.NewComprehensiveValidator()
	tempDir, err := os.MkdirTemp("", "comprehensive_demo")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "go1.21.0.zip")
	testContent := "PK\x03\x04\x14\x00\x00\x00\x08\x00" // 更完整的ZIP头

	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		fmt.Printf("创建测试文件失败: %v\n", err)
		return
	}

	// 计算校验和
	expectedChecksum := fmt.Sprintf("%x", md5.Sum([]byte(testContent)))

	fmt.Printf("验证下载文件: %s\n", testFile)
	fmt.Printf("期望校验和: %s\n", expectedChecksum)

	// 执行综合验证
	result := validator.ValidateDownloadedFile(testFile, expectedChecksum, service.ChecksumTypeMD5)

	fmt.Println("\n验证结果:")
	fmt.Printf("  整体有效性: %t\n", result.Valid)
	fmt.Printf("  校验和验证: %t\n", result.ChecksumOK)
	fmt.Printf("  压缩包验证: %t\n", result.ArchiveOK)

	if len(result.Details) > 0 {
		fmt.Println("  验证详情:")
		for key, value := range result.Details {
			fmt.Printf("    %s: %s\n", key, value)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Println("  验证错误:")
		for i, err := range result.Errors {
			fmt.Printf("    %d. %s\n", i+1, err.Error())
		}
	}

	// 测试错误情况
	fmt.Println("\n测试错误校验和:")
	wrongResult := validator.ValidateDownloadedFile(testFile, "wrongchecksum", service.ChecksumTypeMD5)

	fmt.Printf("  整体有效性: %t\n", wrongResult.Valid)
	fmt.Printf("  错误数量: %d\n", len(wrongResult.Errors))

	if len(wrongResult.Errors) > 0 {
		fmt.Printf("  第一个错误: %s\n", wrongResult.Errors[0].Error())
	}
}

func demonstrateValidationErrors() {
	fmt.Println("\n--- 验证错误处理示例 ---")

	validator := service.NewFileValidator()
	tempDir, err := os.MkdirTemp("", "error_demo")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 测试不存在的文件
	fmt.Println("1. 测试不存在的文件:")
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	_, err = validator.CalculateChecksum(nonExistentFile, service.ChecksumTypeMD5)
	if err != nil {
		fmt.Printf("   错误: %v\n", err)
	}

	// 测试不支持的校验和类型
	fmt.Println("\n2. 测试不支持的校验和类型:")
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err == nil {
		_, err = validator.CalculateChecksum(testFile, "unsupported")
		if err != nil {
			fmt.Printf("   错误: %v\n", err)
		}
	}

	// 测试空校验和
	fmt.Println("\n3. 测试空校验和:")
	err = validator.ValidateChecksum(testFile, "", service.ChecksumTypeMD5)
	if err != nil {
		fmt.Printf("   错误: %v\n", err)

		// 检查错误类型
		if installErr, ok := err.(*service.InstallError); ok {
			fmt.Printf("   错误类型: %s\n", installErr.Type.String())
			fmt.Printf("   建议: %s\n", installErr.GetSuggestion())
		}
	}

	// 测试校验和不匹配
	fmt.Println("\n4. 测试校验和不匹配:")
	err = validator.ValidateChecksum(testFile, "wrongchecksum", service.ChecksumTypeMD5)
	if err != nil {
		fmt.Printf("   错误: %v\n", err)

		if installErr, ok := err.(*service.InstallError); ok {
			fmt.Printf("   错误类型: %s\n", installErr.Type.String())
			fmt.Printf("   可重试: %t\n", installErr.IsRetryable())

			// 显示上下文信息
			if len(installErr.Context) > 0 {
				fmt.Println("   上下文信息:")
				for key, value := range installErr.Context {
					fmt.Printf("     %s: %v\n", key, value)
				}
			}
		}
	}

	// 演示支持的校验和类型
	fmt.Println("\n5. 支持的校验和类型:")
	supportedTypes := validator.GetSupportedChecksumTypes()
	for i, checksumType := range supportedTypes {
		fmt.Printf("   %d. %s\n", i+1, checksumType)
	}
}

// 演示实际的校验和计算
func demonstrateRealChecksums() {
	fmt.Println("\n--- 实际校验和计算示例 ---")

	// 创建一些示例内容
	contents := []string{
		"Hello, World!",
		"Go version manager",
		"File integrity validation",
	}

	for i, content := range contents {
		fmt.Printf("\n内容 %d: %s\n", i+1, content)

		// 计算不同类型的校验和
		md5Sum := fmt.Sprintf("%x", md5.Sum([]byte(content)))
		sha256Sum := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))

		fmt.Printf("  MD5:    %s\n", md5Sum)
		fmt.Printf("  SHA256: %s\n", sha256Sum)
	}
}
