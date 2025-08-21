package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestInstallError_Creation(t *testing.T) {
	cause := errors.New("original error")
	err := NewInstallError(ErrorTypeNetwork, "network connection failed", cause)

	if err.Type != ErrorTypeNetwork {
		t.Errorf("错误类型 = %v, 期望 %v", err.Type, ErrorTypeNetwork)
	}

	if err.Message != "network connection failed" {
		t.Errorf("错误消息 = %s, 期望 %s", err.Message, "network connection failed")
	}

	if err.Cause != cause {
		t.Errorf("错误原因不匹配")
	}

	if len(err.Context) != 0 {
		t.Errorf("上下文应该为空")
	}
}

func TestInstallError_WithContext(t *testing.T) {
	err := NewInstallError(ErrorTypeValidation, "validation failed", nil)
	err.WithContext("version", "1.21.0").WithContext("file", "go1.21.0.zip")

	if len(err.Context) != 2 {
		t.Errorf("上下文数量 = %d, 期望 %d", len(err.Context), 2)
	}

	if err.Context["version"] != "1.21.0" {
		t.Errorf("版本上下文 = %v, 期望 %s", err.Context["version"], "1.21.0")
	}

	if err.Context["file"] != "go1.21.0.zip" {
		t.Errorf("文件上下文 = %v, 期望 %s", err.Context["file"], "go1.21.0.zip")
	}
}

func TestInstallError_IsRetryable(t *testing.T) {
	testCases := []struct {
		errorType ErrorType
		message   string
		retryable bool
	}{
		{ErrorTypeNetwork, "connection failed", true},
		{ErrorTypeTimeout, "request timeout", true},
		{ErrorTypeFileSystem, "temporary file busy", true},
		{ErrorTypeFileSystem, "disk full", false},
		{ErrorTypeValidation, "checksum mismatch", false},
		{ErrorTypePermission, "access denied", false},
		{ErrorTypeUnsupported, "unsupported platform", false},
	}

	for _, tc := range testCases {
		t.Run(tc.errorType.String(), func(t *testing.T) {
			err := NewInstallError(tc.errorType, tc.message, nil)
			result := err.IsRetryable()

			if result != tc.retryable {
				t.Errorf("IsRetryable() = %v, 期望 %v (错误类型: %s)",
					result, tc.retryable, tc.errorType.String())
			}
		})
	}
}

func TestInstallError_GetSuggestion(t *testing.T) {
	testCases := []struct {
		errorType ErrorType
		contains  string
	}{
		{ErrorTypeNetwork, "网络连接"},
		{ErrorTypeFileSystem, "磁盘空间"},
		{ErrorTypeValidation, "版本号格式"},
		{ErrorTypePermission, "权限"},
		{ErrorTypeTimeout, "超时时间"},
		{ErrorTypeVersionExists, "--force"},
	}

	for _, tc := range testCases {
		t.Run(tc.errorType.String(), func(t *testing.T) {
			err := NewInstallError(tc.errorType, "test message", nil)
			suggestion := err.GetSuggestion()

			if !strings.Contains(suggestion, tc.contains) {
				t.Errorf("建议 '%s' 应该包含 '%s'", suggestion, tc.contains)
			}
		})
	}
}

func TestErrorType_String(t *testing.T) {
	testCases := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeNetwork, "NetworkError"},
		{ErrorTypeFileSystem, "FileSystemError"},
		{ErrorTypeValidation, "ValidationError"},
		{ErrorTypeExtraction, "ExtractionError"},
		{ErrorTypePermission, "PermissionError"},
		{ErrorTypeUnsupported, "UnsupportedError"},
		{ErrorTypeTimeout, "TimeoutError"},
		{ErrorTypeCorrupted, "CorruptedError"},
		{ErrorTypeInsufficientSpace, "InsufficientSpaceError"},
		{ErrorTypeVersionExists, "VersionExistsError"},
		{ErrorTypeVersionNotFound, "VersionNotFoundError"},
		{ErrorTypeConfiguration, "ConfigurationError"},
		{ErrorType(999), "UnknownError"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.errorType.String()
			if result != tc.expected {
				t.Errorf("String() = %s, 期望 %s", result, tc.expected)
			}
		})
	}
}

func TestErrorClassifier_ClassifyError(t *testing.T) {
	classifier := NewErrorClassifier()

	testCases := []struct {
		errorMessage string
		expectedType ErrorType
	}{
		{"network connection failed", ErrorTypeNetwork},
		{"connection timeout", ErrorTypeTimeout},
		{"file not found", ErrorTypeFileSystem},
		{"disk space full", ErrorTypeInsufficientSpace},
		{"permission denied", ErrorTypePermission},
		{"checksum mismatch", ErrorTypeValidation},
		{"zip extraction failed", ErrorTypeExtraction},
		{"version already exists", ErrorTypeVersionExists},
		{"version not found", ErrorTypeVersionNotFound},
		{"unknown error", ErrorTypeConfiguration},
	}

	for _, tc := range testCases {
		t.Run(tc.errorMessage, func(t *testing.T) {
			originalErr := errors.New(tc.errorMessage)
			classifiedErr := classifier.ClassifyError(originalErr)

			if classifiedErr.Type != tc.expectedType {
				t.Errorf("分类错误类型 = %v, 期望 %v (消息: %s)",
					classifiedErr.Type, tc.expectedType, tc.errorMessage)
			}

			if classifiedErr.Cause != originalErr {
				t.Error("原始错误应该被保留")
			}
		})
	}
}

func TestErrorClassifier_ClassifyInstallError(t *testing.T) {
	classifier := NewErrorClassifier()

	// 测试已经是InstallError的情况
	originalErr := NewInstallError(ErrorTypeNetwork, "test error", nil)
	classifiedErr := classifier.ClassifyError(originalErr)

	if classifiedErr != originalErr {
		t.Error("InstallError应该直接返回，不进行重新分类")
	}
}

func TestErrorRecoveryStrategy_ShouldRetry(t *testing.T) {
	strategy := NewErrorRecoveryStrategy(3, 2, 1.5)

	// 测试可重试错误
	retryableErr := NewInstallError(ErrorTypeNetwork, "connection failed", nil)

	// 第一次尝试
	if !strategy.ShouldRetry(retryableErr, 0) {
		t.Error("第一次尝试应该允许重试")
	}

	// 第二次尝试
	if !strategy.ShouldRetry(retryableErr, 1) {
		t.Error("第二次尝试应该允许重试")
	}

	// 第三次尝试
	if !strategy.ShouldRetry(retryableErr, 2) {
		t.Error("第三次尝试应该允许重试")
	}

	// 第四次尝试（超过最大重试次数）
	if strategy.ShouldRetry(retryableErr, 3) {
		t.Error("超过最大重试次数时不应该重试")
	}

	// 测试不可重试错误
	nonRetryableErr := NewInstallError(ErrorTypeValidation, "validation failed", nil)
	if strategy.ShouldRetry(nonRetryableErr, 0) {
		t.Error("不可重试错误不应该重试")
	}
}

func TestErrorRecoveryStrategy_GetRetryDelay(t *testing.T) {
	strategy := NewErrorRecoveryStrategy(3, 2, 2.0)

	testCases := []struct {
		attempt     int
		expectedMin int
		expectedMax int
	}{
		{0, 2, 2},   // 2 * 2^0 = 2
		{1, 4, 4},   // 2 * 2^1 = 4
		{2, 8, 8},   // 2 * 2^2 = 8
		{3, 16, 16}, // 2 * 2^3 = 16
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("attempt_%d", tc.attempt), func(t *testing.T) {
			delay := strategy.GetRetryDelay(tc.attempt)

			if delay < tc.expectedMin || delay > tc.expectedMax {
				t.Errorf("重试延迟 = %d, 期望在 %d-%d 之间",
					delay, tc.expectedMin, tc.expectedMax)
			}
		})
	}
}

func TestErrorReporter_ReportError(t *testing.T) {
	// 测试详细模式
	verboseReporter := NewErrorReporter(true)
	err := NewInstallError(ErrorTypeNetwork, "connection failed", errors.New("original error"))
	err.WithContext("version", "1.21.0")

	report := verboseReporter.ReportError(err)

	// 验证报告包含关键信息
	if !strings.Contains(report, "connection failed") {
		t.Error("报告应该包含错误消息")
	}

	if !strings.Contains(report, "NetworkError") {
		t.Error("详细模式应该包含错误类型")
	}

	if !strings.Contains(report, "建议") {
		t.Error("报告应该包含建议")
	}

	if !strings.Contains(report, "上下文信息") {
		t.Error("详细模式应该包含上下文信息")
	}

	if !strings.Contains(report, "original error") {
		t.Error("详细模式应该包含原始错误")
	}

	// 测试简洁模式
	simpleReporter := NewErrorReporter(false)
	simpleReport := simpleReporter.ReportError(err)

	if strings.Contains(simpleReport, "NetworkError") {
		t.Error("简洁模式不应该包含错误类型")
	}

	if strings.Contains(simpleReport, "上下文信息") {
		t.Error("简洁模式不应该包含上下文信息")
	}
}
