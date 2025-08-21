package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"version-list/internal/domain/service"
)

func main() {
	fmt.Println("=== 错误处理和回滚机制示例 ===")

	// 演示错误分类
	demonstrateErrorClassification()

	// 演示回滚管理器
	demonstrateRollbackManager()

	// 演示安装清理器
	demonstrateInstallationCleaner()

	// 演示错误恢复策略
	demonstrateErrorRecoveryStrategy()

	fmt.Println("\n=== 示例完成 ===")
}

func demonstrateErrorClassification() {
	fmt.Println("\n--- 错误分类示例 ---")

	classifier := service.NewErrorClassifier()
	reporter := service.NewErrorReporter(true)

	// 测试不同类型的错误
	testErrors := []error{
		errors.New("network connection failed"),
		errors.New("file not found"),
		errors.New("permission denied"),
		errors.New("checksum mismatch"),
		errors.New("zip extraction failed"),
		errors.New("connection timeout"),
		errors.New("disk space full"),
		errors.New("version already exists"),
	}

	for _, err := range testErrors {
		classifiedErr := classifier.ClassifyError(err)

		fmt.Printf("\n原始错误: %s\n", err.Error())
		fmt.Printf("错误类型: %s\n", classifiedErr.Type.String())
		fmt.Printf("可重试: %t\n", classifiedErr.IsRetryable())
		fmt.Printf("建议: %s\n", classifiedErr.GetSuggestion())

		// 生成错误报告
		report := reporter.ReportError(classifiedErr)
		fmt.Printf("错误报告:\n%s", report)
	}
}

func demonstrateRollbackManager() {
	fmt.Println("\n--- 回滚管理器示例 ---")

	// 创建临时目录进行演示
	tempDir, err := os.MkdirTemp("", "rollback_demo")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	rm := service.NewRollbackManager()

	// 模拟安装过程中的操作
	fmt.Println("模拟安装过程...")

	// 1. 创建安装目录
	installDir := filepath.Join(tempDir, "go-1.21.0")
	err = os.MkdirAll(installDir, 0755)
	if err != nil {
		fmt.Printf("创建安装目录失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 创建安装目录: %s\n", installDir)
	rm.RegisterDirectoryCreation(installDir)

	// 2. 创建配置文件
	configFile := filepath.Join(installDir, "config.json")
	err = os.WriteFile(configFile, []byte(`{"version": "1.21.0"}`), 0644)
	if err != nil {
		fmt.Printf("创建配置文件失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 创建配置文件: %s\n", configFile)
	rm.RegisterFileCreation(configFile)

	// 3. 创建二进制文件
	binDir := filepath.Join(installDir, "bin")
	err = os.MkdirAll(binDir, 0755)
	if err != nil {
		fmt.Printf("创建bin目录失败: %v\n", err)
		return
	}

	goExec := filepath.Join(binDir, "go")
	err = os.WriteFile(goExec, []byte("#!/bin/bash\necho 'go version go1.21.0'\n"), 0755)
	if err != nil {
		fmt.Printf("创建go可执行文件失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 创建go可执行文件: %s\n", goExec)
	rm.RegisterFileCreation(goExec)

	fmt.Printf("\n注册了 %d 个回滚操作\n", rm.Count())

	// 模拟安装失败，需要回滚
	fmt.Println("\n模拟安装失败，执行回滚...")
	err = rm.Execute()
	if err != nil {
		fmt.Printf("回滚失败: %v\n", err)
	} else {
		fmt.Println("✓ 回滚成功")
	}

	// 验证文件和目录被清理
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		fmt.Println("✓ 安装目录已被清理")
	} else {
		fmt.Println("❌ 安装目录未被清理")
	}

	fmt.Printf("执行了 %d 个回滚操作\n", rm.GetExecutedCount())
}

func demonstrateInstallationCleaner() {
	fmt.Println("\n--- 安装清理器示例 ---")

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "cleaner_demo")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	cleaner := service.NewInstallationCleaner()

	// 模拟下载过程
	fmt.Println("模拟下载和解压过程...")

	// 创建下载文件
	downloadFile := filepath.Join(tempDir, "go1.21.0.zip")
	err = os.WriteFile(downloadFile, []byte("fake zip content"), 0644)
	if err != nil {
		fmt.Printf("创建下载文件失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 下载文件: %s\n", downloadFile)
	cleaner.AddTempFile(downloadFile)

	// 创建解压目录
	extractDir := filepath.Join(tempDir, "extracted")
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		fmt.Printf("创建解压目录失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 解压目录: %s\n", extractDir)
	cleaner.AddTempDir(extractDir)

	// 添加一些回滚操作
	operationExecuted := false
	cleaner.GetRollbackManager().Register(func() error {
		operationExecuted = true
		fmt.Println("✓ 执行自定义回滚操作")
		return nil
	})

	// 模拟安装失败，执行清理和回滚
	fmt.Println("\n模拟安装失败，执行清理和回滚...")
	err = cleaner.CleanupAndRollback()
	if err != nil {
		fmt.Printf("清理和回滚失败: %v\n", err)
	} else {
		fmt.Println("✓ 清理和回滚成功")
	}

	// 验证清理结果
	if _, err := os.Stat(downloadFile); os.IsNotExist(err) {
		fmt.Println("✓ 下载文件已被清理")
	}

	if _, err := os.Stat(extractDir); os.IsNotExist(err) {
		fmt.Println("✓ 解压目录已被清理")
	}

	if operationExecuted {
		fmt.Println("✓ 自定义回滚操作已执行")
	}
}

func demonstrateErrorRecoveryStrategy() {
	fmt.Println("\n--- 错误恢复策略示例 ---")

	strategy := service.NewErrorRecoveryStrategy(3, 2, 2.0)

	// 创建不同类型的错误
	retryableErr := service.NewInstallError(service.ErrorTypeNetwork, "connection failed", nil)
	nonRetryableErr := service.NewInstallError(service.ErrorTypeValidation, "checksum mismatch", nil)

	fmt.Println("测试可重试错误:")
	for attempt := 0; attempt < 5; attempt++ {
		shouldRetry := strategy.ShouldRetry(retryableErr, attempt)
		delay := strategy.GetRetryDelay(attempt)

		fmt.Printf("  尝试 %d: 应该重试=%t, 延迟=%d秒\n", attempt+1, shouldRetry, delay)

		if !shouldRetry {
			break
		}
	}

	fmt.Println("\n测试不可重试错误:")
	shouldRetry := strategy.ShouldRetry(nonRetryableErr, 0)
	fmt.Printf("  第一次尝试: 应该重试=%t\n", shouldRetry)

	// 演示错误上下文
	fmt.Println("\n错误上下文示例:")
	contextErr := service.NewInstallError(service.ErrorTypeExtraction, "解压失败", errors.New("archive corrupted"))
	contextErr.WithContext("version", "1.21.0")
	contextErr.WithContext("file", "go1.21.0.zip")
	contextErr.WithContext("size", 163553657)

	reporter := service.NewErrorReporter(true)
	report := reporter.ReportError(contextErr)
	fmt.Printf("详细错误报告:\n%s", report)
}
