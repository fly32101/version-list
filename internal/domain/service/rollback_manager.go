package service

import (
	"fmt"
	"os"
	"sync"
	"version-list/internal/domain/model"
)

// RollbackOperation 回滚操作函数类型
type RollbackOperation func() error

// RollbackManager 回滚管理器
type RollbackManager struct {
	mu         sync.RWMutex
	operations []RollbackOperation
	executed   []bool
	enabled    bool
}

// NewRollbackManager 创建回滚管理器
func NewRollbackManager() *RollbackManager {
	return &RollbackManager{
		operations: make([]RollbackOperation, 0),
		executed:   make([]bool, 0),
		enabled:    true,
	}
}

// Register 注册回滚操作
func (rm *RollbackManager) Register(op RollbackOperation) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.operations = append(rm.operations, op)
	rm.executed = append(rm.executed, false)
}

// RegisterFileCreation 注册文件创建的回滚操作
func (rm *RollbackManager) RegisterFileCreation(filePath string) {
	rm.Register(func() error {
		if _, err := os.Stat(filePath); err == nil {
			return os.Remove(filePath)
		}
		return nil
	})
}

// RegisterDirectoryCreation 注册目录创建的回滚操作
func (rm *RollbackManager) RegisterDirectoryCreation(dirPath string) {
	rm.Register(func() error {
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			return os.RemoveAll(dirPath)
		}
		return nil
	})
}

// RegisterFileMove 注册文件移动的回滚操作
func (rm *RollbackManager) RegisterFileMove(srcPath, destPath string) {
	rm.Register(func() error {
		// 如果目标文件存在，将其移回源位置
		if _, err := os.Stat(destPath); err == nil {
			return os.Rename(destPath, srcPath)
		}
		return nil
	})
}

// RegisterVersionRecord 注册版本记录的回滚操作
func (rm *RollbackManager) RegisterVersionRecord(versionRepo VersionRepository, version string) {
	rm.Register(func() error {
		// 删除版本记录
		return versionRepo.Remove(version)
	})
}

// RegisterEnvironmentChange 注册环境变量更改的回滚操作
func (rm *RollbackManager) RegisterEnvironmentChange(envRepo EnvironmentRepository, originalEnv *model.Environment) {
	rm.Register(func() error {
		// 恢复原始环境变量
		return envRepo.Save(originalEnv)
	})
}

// Execute 执行所有回滚操作
func (rm *RollbackManager) Execute() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.enabled {
		return nil
	}

	var errors []error

	// 逆序执行回滚操作
	for i := len(rm.operations) - 1; i >= 0; i-- {
		if !rm.executed[i] {
			if err := rm.operations[i](); err != nil {
				errors = append(errors, fmt.Errorf("回滚操作 %d 失败: %v", i, err))
			} else {
				rm.executed[i] = true
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("回滚过程中发生 %d 个错误: %v", len(errors), errors)
	}

	return nil
}

// ExecutePartial 执行部分回滚操作
func (rm *RollbackManager) ExecutePartial(fromIndex int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.enabled {
		return nil
	}

	var errors []error

	// 从指定索引开始逆序执行
	for i := len(rm.operations) - 1; i >= fromIndex && i >= 0; i-- {
		if !rm.executed[i] {
			if err := rm.operations[i](); err != nil {
				errors = append(errors, fmt.Errorf("回滚操作 %d 失败: %v", i, err))
			} else {
				rm.executed[i] = true
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分回滚过程中发生 %d 个错误: %v", len(errors), errors)
	}

	return nil
}

// Clear 清除所有回滚操作
func (rm *RollbackManager) Clear() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.operations = rm.operations[:0]
	rm.executed = rm.executed[:0]
}

// Disable 禁用回滚管理器
func (rm *RollbackManager) Disable() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.enabled = false
}

// Enable 启用回滚管理器
func (rm *RollbackManager) Enable() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.enabled = true
}

// IsEnabled 检查回滚管理器是否启用
func (rm *RollbackManager) IsEnabled() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.enabled
}

// Count 获取注册的回滚操作数量
func (rm *RollbackManager) Count() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.operations)
}

// GetExecutedCount 获取已执行的回滚操作数量
func (rm *RollbackManager) GetExecutedCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	count := 0
	for _, executed := range rm.executed {
		if executed {
			count++
		}
	}
	return count
}

// InstallationCleaner 安装清理器
type InstallationCleaner struct {
	rollbackManager *RollbackManager
	tempFiles       []string
	tempDirs        []string
}

// NewInstallationCleaner 创建安装清理器
func NewInstallationCleaner() *InstallationCleaner {
	return &InstallationCleaner{
		rollbackManager: NewRollbackManager(),
		tempFiles:       make([]string, 0),
		tempDirs:        make([]string, 0),
	}
}

// GetRollbackManager 获取回滚管理器
func (ic *InstallationCleaner) GetRollbackManager() *RollbackManager {
	return ic.rollbackManager
}

// AddTempFile 添加临时文件
func (ic *InstallationCleaner) AddTempFile(filePath string) {
	ic.tempFiles = append(ic.tempFiles, filePath)
	ic.rollbackManager.RegisterFileCreation(filePath)
}

// AddTempDir 添加临时目录
func (ic *InstallationCleaner) AddTempDir(dirPath string) {
	ic.tempDirs = append(ic.tempDirs, dirPath)
	ic.rollbackManager.RegisterDirectoryCreation(dirPath)
}

// CleanupTempFiles 清理临时文件
func (ic *InstallationCleaner) CleanupTempFiles() error {
	var errors []error

	// 清理临时文件
	for _, filePath := range ic.tempFiles {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("删除临时文件 %s 失败: %v", filePath, err))
		}
	}

	// 清理临时目录
	for _, dirPath := range ic.tempDirs {
		if err := os.RemoveAll(dirPath); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("删除临时目录 %s 失败: %v", dirPath, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("清理临时文件时发生 %d 个错误: %v", len(errors), errors)
	}

	return nil
}

// ExecuteRollback 执行回滚
func (ic *InstallationCleaner) ExecuteRollback() error {
	return ic.rollbackManager.Execute()
}

// CleanupAndRollback 清理并回滚
func (ic *InstallationCleaner) CleanupAndRollback() error {
	// 先执行回滚
	rollbackErr := ic.ExecuteRollback()

	// 再清理临时文件
	cleanupErr := ic.CleanupTempFiles()

	// 合并错误
	if rollbackErr != nil && cleanupErr != nil {
		return fmt.Errorf("回滚失败: %v; 清理失败: %v", rollbackErr, cleanupErr)
	} else if rollbackErr != nil {
		return rollbackErr
	} else if cleanupErr != nil {
		return cleanupErr
	}

	return nil
}

// VersionRepository 接口（避免循环导入）
type VersionRepository interface {
	Remove(version string) error
}

// EnvironmentRepository 接口（避免循环导入）
type EnvironmentRepository interface {
	Save(env *model.Environment) error
}
