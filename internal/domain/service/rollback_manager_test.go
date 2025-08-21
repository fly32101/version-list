package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"version-list/internal/domain/model"
)

func TestRollbackManager_Basic(t *testing.T) {
	rm := NewRollbackManager()

	// 测试初始状态
	if rm.Count() != 0 {
		t.Errorf("初始回滚操作数量应该为0, 得到 %d", rm.Count())
	}

	if !rm.IsEnabled() {
		t.Error("回滚管理器应该默认启用")
	}

	// 注册回滚操作
	executed := false
	rm.Register(func() error {
		executed = true
		return nil
	})

	if rm.Count() != 1 {
		t.Errorf("注册后回滚操作数量应该为1, 得到 %d", rm.Count())
	}

	// 执行回滚
	err := rm.Execute()
	if err != nil {
		t.Errorf("执行回滚失败: %v", err)
	}

	if !executed {
		t.Error("回滚操作应该被执行")
	}

	if rm.GetExecutedCount() != 1 {
		t.Errorf("已执行回滚操作数量应该为1, 得到 %d", rm.GetExecutedCount())
	}
}

func TestRollbackManager_MultipleOperations(t *testing.T) {
	rm := NewRollbackManager()

	var executionOrder []int

	// 注册多个回滚操作
	for i := 0; i < 3; i++ {
		index := i
		rm.Register(func() error {
			executionOrder = append(executionOrder, index)
			return nil
		})
	}

	// 执行回滚
	err := rm.Execute()
	if err != nil {
		t.Errorf("执行回滚失败: %v", err)
	}

	// 验证执行顺序（应该是逆序）
	expectedOrder := []int{2, 1, 0}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("执行顺序长度 = %d, 期望 %d", len(executionOrder), len(expectedOrder))
	}

	for i, expected := range expectedOrder {
		if i < len(executionOrder) && executionOrder[i] != expected {
			t.Errorf("执行顺序[%d] = %d, 期望 %d", i, executionOrder[i], expected)
		}
	}
}

func TestRollbackManager_ErrorHandling(t *testing.T) {
	rm := NewRollbackManager()

	// 注册一个会失败的回滚操作
	rm.Register(func() error {
		return errors.New("rollback failed")
	})

	// 注册一个正常的回滚操作
	executed := false
	rm.Register(func() error {
		executed = true
		return nil
	})

	// 执行回滚
	err := rm.Execute()
	if err == nil {
		t.Error("期望回滚执行返回错误")
	}

	// 验证正常的操作仍然被执行
	if !executed {
		t.Error("正常的回滚操作应该被执行")
	}
}

func TestRollbackManager_FileOperations(t *testing.T) {
	rm := NewRollbackManager()
	tempDir := t.TempDir()

	// 测试文件创建回滚
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	rm.RegisterFileCreation(testFile)

	// 执行回滚
	err = rm.Execute()
	if err != nil {
		t.Errorf("执行文件回滚失败: %v", err)
	}

	// 验证文件被删除
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("文件应该被删除")
	}
}

func TestRollbackManager_DirectoryOperations(t *testing.T) {
	rm := NewRollbackManager()
	tempDir := t.TempDir()

	// 创建测试目录
	testDir := filepath.Join(tempDir, "testdir")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}

	// 在目录中创建文件
	testFile := filepath.Join(testDir, "file.txt")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	rm.RegisterDirectoryCreation(testDir)

	// 执行回滚
	err = rm.Execute()
	if err != nil {
		t.Errorf("执行目录回滚失败: %v", err)
	}

	// 验证目录被删除
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("目录应该被删除")
	}
}

func TestRollbackManager_FileMoveOperations(t *testing.T) {
	rm := NewRollbackManager()
	tempDir := t.TempDir()

	// 创建源文件
	srcFile := filepath.Join(tempDir, "src.txt")
	err := os.WriteFile(srcFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("创建源文件失败: %v", err)
	}

	// 移动文件
	destFile := filepath.Join(tempDir, "dest.txt")
	err = os.Rename(srcFile, destFile)
	if err != nil {
		t.Fatalf("移动文件失败: %v", err)
	}

	rm.RegisterFileMove(srcFile, destFile)

	// 执行回滚
	err = rm.Execute()
	if err != nil {
		t.Errorf("执行文件移动回滚失败: %v", err)
	}

	// 验证文件被移回
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("源文件应该被恢复")
	}

	if _, err := os.Stat(destFile); !os.IsNotExist(err) {
		t.Error("目标文件应该不存在")
	}
}

func TestRollbackManager_DisableEnable(t *testing.T) {
	rm := NewRollbackManager()

	executed := false
	rm.Register(func() error {
		executed = true
		return nil
	})

	// 禁用回滚管理器
	rm.Disable()
	if rm.IsEnabled() {
		t.Error("回滚管理器应该被禁用")
	}

	// 执行回滚（应该不执行）
	err := rm.Execute()
	if err != nil {
		t.Errorf("禁用状态下执行回滚失败: %v", err)
	}

	if executed {
		t.Error("禁用状态下回滚操作不应该被执行")
	}

	// 重新启用
	rm.Enable()
	if !rm.IsEnabled() {
		t.Error("回滚管理器应该被启用")
	}

	// 再次执行回滚
	err = rm.Execute()
	if err != nil {
		t.Errorf("启用状态下执行回滚失败: %v", err)
	}

	if !executed {
		t.Error("启用状态下回滚操作应该被执行")
	}
}

func TestRollbackManager_PartialExecution(t *testing.T) {
	rm := NewRollbackManager()

	var executionOrder []int

	// 注册多个回滚操作
	for i := 0; i < 5; i++ {
		index := i
		rm.Register(func() error {
			executionOrder = append(executionOrder, index)
			return nil
		})
	}

	// 执行部分回滚（从索引2开始）
	err := rm.ExecutePartial(2)
	if err != nil {
		t.Errorf("执行部分回滚失败: %v", err)
	}

	// 验证只执行了索引2-4的操作（逆序）
	expectedOrder := []int{4, 3, 2}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("部分执行顺序长度 = %d, 期望 %d", len(executionOrder), len(expectedOrder))
	}

	for i, expected := range expectedOrder {
		if i < len(executionOrder) && executionOrder[i] != expected {
			t.Errorf("部分执行顺序[%d] = %d, 期望 %d", i, executionOrder[i], expected)
		}
	}
}

func TestRollbackManager_Clear(t *testing.T) {
	rm := NewRollbackManager()

	// 注册一些回滚操作
	rm.Register(func() error { return nil })
	rm.Register(func() error { return nil })

	if rm.Count() != 2 {
		t.Errorf("清理前回滚操作数量应该为2, 得到 %d", rm.Count())
	}

	// 清理
	rm.Clear()

	if rm.Count() != 0 {
		t.Errorf("清理后回滚操作数量应该为0, 得到 %d", rm.Count())
	}
}

func TestInstallationCleaner_Basic(t *testing.T) {
	cleaner := NewInstallationCleaner()

	if cleaner.GetRollbackManager() == nil {
		t.Error("安装清理器应该有回滚管理器")
	}

	tempDir := t.TempDir()

	// 添加临时文件
	tempFile := filepath.Join(tempDir, "temp.txt")
	err := os.WriteFile(tempFile, []byte("temp content"), 0644)
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}

	cleaner.AddTempFile(tempFile)

	// 添加临时目录
	tempSubDir := filepath.Join(tempDir, "tempdir")
	err = os.MkdirAll(tempSubDir, 0755)
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	cleaner.AddTempDir(tempSubDir)

	// 清理临时文件
	err = cleaner.CleanupTempFiles()
	if err != nil {
		t.Errorf("清理临时文件失败: %v", err)
	}

	// 验证文件和目录被删除
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("临时文件应该被删除")
	}

	if _, err := os.Stat(tempSubDir); !os.IsNotExist(err) {
		t.Error("临时目录应该被删除")
	}
}

func TestInstallationCleaner_CleanupAndRollback(t *testing.T) {
	cleaner := NewInstallationCleaner()
	tempDir := t.TempDir()

	// 创建临时文件
	tempFile := filepath.Join(tempDir, "temp.txt")
	err := os.WriteFile(tempFile, []byte("temp content"), 0644)
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}

	cleaner.AddTempFile(tempFile)

	// 添加回滚操作
	rollbackExecuted := false
	cleaner.GetRollbackManager().Register(func() error {
		rollbackExecuted = true
		return nil
	})

	// 执行清理和回滚
	err = cleaner.CleanupAndRollback()
	if err != nil {
		t.Errorf("清理和回滚失败: %v", err)
	}

	// 验证回滚被执行
	if !rollbackExecuted {
		t.Error("回滚操作应该被执行")
	}

	// 验证临时文件被清理
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("临时文件应该被删除")
	}
}

// MockVersionRepositoryForRollback Mock实现用于回滚测试
type MockVersionRepositoryForRollback struct {
	removedVersions []string
}

func (m *MockVersionRepositoryForRollback) Remove(version string) error {
	m.removedVersions = append(m.removedVersions, version)
	return nil
}

type MockEnvironmentRepositoryForRollback struct {
	savedEnv *model.Environment
}

func (m *MockEnvironmentRepositoryForRollback) Save(env *model.Environment) error {
	m.savedEnv = env
	return nil
}

func TestRollbackManager_VersionRecord(t *testing.T) {
	rm := NewRollbackManager()
	mockRepo := &MockVersionRepositoryForRollback{}

	// 注册版本记录回滚
	rm.RegisterVersionRecord(mockRepo, "1.21.0")

	// 执行回滚
	err := rm.Execute()
	if err != nil {
		t.Errorf("执行版本记录回滚失败: %v", err)
	}

	// 验证版本被删除
	if len(mockRepo.removedVersions) != 1 {
		t.Errorf("删除版本数量 = %d, 期望 1", len(mockRepo.removedVersions))
	}

	if mockRepo.removedVersions[0] != "1.21.0" {
		t.Errorf("删除版本 = %s, 期望 1.21.0", mockRepo.removedVersions[0])
	}
}

func TestRollbackManager_EnvironmentChange(t *testing.T) {
	rm := NewRollbackManager()
	mockRepo := &MockEnvironmentRepositoryForRollback{}

	originalEnv := &model.Environment{
		GOROOT: "/original/go",
		GOPATH: "/original/gopath",
		GOBIN:  "/original/go/bin",
	}

	// 注册环境变量回滚
	rm.RegisterEnvironmentChange(mockRepo, originalEnv)

	// 执行回滚
	err := rm.Execute()
	if err != nil {
		t.Errorf("执行环境变量回滚失败: %v", err)
	}

	// 验证环境变量被恢复
	if mockRepo.savedEnv == nil {
		t.Error("环境变量应该被保存")
	}

	if mockRepo.savedEnv.GOROOT != originalEnv.GOROOT {
		t.Errorf("GOROOT = %s, 期望 %s", mockRepo.savedEnv.GOROOT, originalEnv.GOROOT)
	}
}
