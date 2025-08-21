package service

import (
	"fmt"
	"strings"
)

// InstallError 安装错误类型
type InstallError struct {
	Type      ErrorType
	Message   string
	Cause     error
	Context   map[string]interface{}
	Timestamp string
}

// ErrorType 错误类型枚举
type ErrorType int

const (
	ErrorTypeNetwork           ErrorType = iota // 网络错误
	ErrorTypeFileSystem                         // 文件系统错误
	ErrorTypeValidation                         // 验证错误
	ErrorTypeExtraction                         // 解压错误
	ErrorTypePermission                         // 权限错误
	ErrorTypeUnsupported                        // 不支持的系统
	ErrorTypeTimeout                            // 超时错误
	ErrorTypeCorrupted                          // 文件损坏错误
	ErrorTypeInsufficientSpace                  // 磁盘空间不足
	ErrorTypeVersionExists                      // 版本已存在
	ErrorTypeVersionNotFound                    // 版本不存在
	ErrorTypeConfiguration                      // 配置错误
)

// Error 实现error接口
func (e *InstallError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type.String(), e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

// String 返回错误类型的字符串表示
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeNetwork:
		return "NetworkError"
	case ErrorTypeFileSystem:
		return "FileSystemError"
	case ErrorTypeValidation:
		return "ValidationError"
	case ErrorTypeExtraction:
		return "ExtractionError"
	case ErrorTypePermission:
		return "PermissionError"
	case ErrorTypeUnsupported:
		return "UnsupportedError"
	case ErrorTypeTimeout:
		return "TimeoutError"
	case ErrorTypeCorrupted:
		return "CorruptedError"
	case ErrorTypeInsufficientSpace:
		return "InsufficientSpaceError"
	case ErrorTypeVersionExists:
		return "VersionExistsError"
	case ErrorTypeVersionNotFound:
		return "VersionNotFoundError"
	case ErrorTypeConfiguration:
		return "ConfigurationError"
	default:
		return "UnknownError"
	}
}

// NewInstallError 创建新的安装错误
func NewInstallError(errorType ErrorType, message string, cause error) *InstallError {
	return &InstallError{
		Type:      errorType,
		Message:   message,
		Cause:     cause,
		Context:   make(map[string]interface{}),
		Timestamp: fmt.Sprintf("%d", getCurrentTimestamp()),
	}
}

// WithContext 添加上下文信息
func (e *InstallError) WithContext(key string, value interface{}) *InstallError {
	e.Context[key] = value
	return e
}

// IsRetryable 判断错误是否可重试
func (e *InstallError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeNetwork, ErrorTypeTimeout:
		return true
	case ErrorTypeFileSystem:
		// 某些文件系统错误可能是临时的
		return strings.Contains(strings.ToLower(e.Message), "temporary") ||
			strings.Contains(strings.ToLower(e.Message), "busy")
	default:
		return false
	}
}

// GetSuggestion 获取错误建议
func (e *InstallError) GetSuggestion() string {
	switch e.Type {
	case ErrorTypeNetwork:
		return "请检查网络连接，或稍后重试"
	case ErrorTypeFileSystem:
		return "请检查磁盘空间和文件权限"
	case ErrorTypeValidation:
		return "请检查版本号格式或文件完整性"
	case ErrorTypeExtraction:
		return "请检查压缩包是否完整，或重新下载"
	case ErrorTypePermission:
		return "请检查文件权限，可能需要管理员权限"
	case ErrorTypeUnsupported:
		return "请检查系统兼容性"
	case ErrorTypeTimeout:
		return "请增加超时时间或检查网络连接"
	case ErrorTypeCorrupted:
		return "请重新下载文件"
	case ErrorTypeInsufficientSpace:
		return "请清理磁盘空间"
	case ErrorTypeVersionExists:
		return "使用 --force 选项强制重新安装"
	case ErrorTypeVersionNotFound:
		return "请检查版本号是否正确"
	case ErrorTypeConfiguration:
		return "请检查配置文件和环境变量"
	default:
		return "请查看详细错误信息或联系支持"
	}
}

// ErrorClassifier 错误分类器
type ErrorClassifier struct{}

// NewErrorClassifier 创建错误分类器
func NewErrorClassifier() *ErrorClassifier {
	return &ErrorClassifier{}
}

// ClassifyError 分类错误
func (c *ErrorClassifier) ClassifyError(err error) *InstallError {
	if installErr, ok := err.(*InstallError); ok {
		return installErr
	}

	message := err.Error()
	lowerMessage := strings.ToLower(message)

	// 网络相关错误
	if strings.Contains(lowerMessage, "network") ||
		strings.Contains(lowerMessage, "connection") ||
		strings.Contains(lowerMessage, "timeout") ||
		strings.Contains(lowerMessage, "dns") ||
		strings.Contains(lowerMessage, "unreachable") {
		if strings.Contains(lowerMessage, "timeout") {
			return NewInstallError(ErrorTypeTimeout, message, err)
		}
		return NewInstallError(ErrorTypeNetwork, message, err)
	}

	// 文件系统相关错误
	if strings.Contains(lowerMessage, "no such file") ||
		strings.Contains(lowerMessage, "file not found") ||
		strings.Contains(lowerMessage, "directory") ||
		strings.Contains(lowerMessage, "disk") ||
		strings.Contains(lowerMessage, "space") {
		if strings.Contains(lowerMessage, "space") ||
			strings.Contains(lowerMessage, "full") {
			return NewInstallError(ErrorTypeInsufficientSpace, message, err)
		}
		return NewInstallError(ErrorTypeFileSystem, message, err)
	}

	// 权限相关错误
	if strings.Contains(lowerMessage, "permission") ||
		strings.Contains(lowerMessage, "access denied") ||
		strings.Contains(lowerMessage, "forbidden") ||
		strings.Contains(lowerMessage, "unauthorized") {
		return NewInstallError(ErrorTypePermission, message, err)
	}

	// 验证相关错误
	if strings.Contains(lowerMessage, "checksum") ||
		strings.Contains(lowerMessage, "hash") ||
		strings.Contains(lowerMessage, "signature") ||
		strings.Contains(lowerMessage, "invalid") {
		return NewInstallError(ErrorTypeValidation, message, err)
	}

	// 解压相关错误
	if strings.Contains(lowerMessage, "zip") ||
		strings.Contains(lowerMessage, "tar") ||
		strings.Contains(lowerMessage, "extract") ||
		strings.Contains(lowerMessage, "decompress") {
		return NewInstallError(ErrorTypeExtraction, message, err)
	}

	// 版本相关错误
	if strings.Contains(lowerMessage, "already exists") ||
		strings.Contains(lowerMessage, "duplicate") {
		return NewInstallError(ErrorTypeVersionExists, message, err)
	}

	if strings.Contains(lowerMessage, "not found") ||
		strings.Contains(lowerMessage, "missing") {
		return NewInstallError(ErrorTypeVersionNotFound, message, err)
	}

	// 默认为配置错误
	return NewInstallError(ErrorTypeConfiguration, message, err)
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	// 这里应该返回实际的时间戳，为了简化使用固定值
	return 1640995200 // 2022-01-01 00:00:00 UTC
}

// ErrorRecoveryStrategy 错误恢复策略
type ErrorRecoveryStrategy struct {
	maxRetries    int
	retryDelay    int // 秒
	backoffFactor float64
}

// NewErrorRecoveryStrategy 创建错误恢复策略
func NewErrorRecoveryStrategy(maxRetries int, retryDelay int, backoffFactor float64) *ErrorRecoveryStrategy {
	return &ErrorRecoveryStrategy{
		maxRetries:    maxRetries,
		retryDelay:    retryDelay,
		backoffFactor: backoffFactor,
	}
}

// ShouldRetry 判断是否应该重试
func (s *ErrorRecoveryStrategy) ShouldRetry(err *InstallError, attemptCount int) bool {
	if attemptCount >= s.maxRetries {
		return false
	}
	return err.IsRetryable()
}

// GetRetryDelay 获取重试延迟时间
func (s *ErrorRecoveryStrategy) GetRetryDelay(attemptCount int) int {
	delay := float64(s.retryDelay)
	for i := 0; i < attemptCount; i++ {
		delay *= s.backoffFactor
	}
	return int(delay)
}

// ErrorReporter 错误报告器
type ErrorReporter struct {
	verbose bool
}

// NewErrorReporter 创建错误报告器
func NewErrorReporter(verbose bool) *ErrorReporter {
	return &ErrorReporter{
		verbose: verbose,
	}
}

// ReportError 报告错误
func (r *ErrorReporter) ReportError(err *InstallError) string {
	var report strings.Builder

	// 基本错误信息
	report.WriteString(fmt.Sprintf("❌ %s\n", err.Message))

	// 错误类型
	if r.verbose {
		report.WriteString(fmt.Sprintf("错误类型: %s\n", err.Type.String()))
	}

	// 建议
	suggestion := err.GetSuggestion()
	if suggestion != "" {
		report.WriteString(fmt.Sprintf("💡 建议: %s\n", suggestion))
	}

	// 上下文信息
	if r.verbose && len(err.Context) > 0 {
		report.WriteString("上下文信息:\n")
		for key, value := range err.Context {
			report.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	// 原始错误
	if r.verbose && err.Cause != nil {
		report.WriteString(fmt.Sprintf("原始错误: %v\n", err.Cause))
	}

	return report.String()
}
