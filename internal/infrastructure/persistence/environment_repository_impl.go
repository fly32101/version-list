package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"version-list/internal/domain/model"
)

const (
	envFile = "environment.json"
)

// EnvironmentRepositoryImpl 环境变量仓库实现
type EnvironmentRepositoryImpl struct {
	configPath string
}

// NewEnvironmentRepositoryImpl 创建环境变量仓库实现实例
func NewEnvironmentRepositoryImpl() (*EnvironmentRepositoryImpl, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %v", err)
	}

	configPath := filepath.Join(homeDir, ".go-version")
	// 确保配置目录存在
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return nil, fmt.Errorf("创建配置目录失败: %v", err)
	}

	return &EnvironmentRepositoryImpl{
		configPath: configPath,
	}, nil
}

// getEnvFilePath 获取环境变量文件路径
func (r *EnvironmentRepositoryImpl) getEnvFilePath() string {
	return filepath.Join(r.configPath, envFile)
}

// loadEnvironment 加载环境变量配置
func (r *EnvironmentRepositoryImpl) loadEnvironment() (*model.Environment, error) {
	filePath := r.getEnvFilePath()

	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return r.getDefaultEnvironment(), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取环境变量配置文件失败: %v", err)
	}

	var env model.Environment
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("解析环境变量配置失败: %v", err)
	}

	return &env, nil
}

// saveEnvironment 保存环境变量配置
func (r *EnvironmentRepositoryImpl) saveEnvironment(env *model.Environment) error {
	filePath := r.getEnvFilePath()

	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化环境变量配置失败: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入环境变量配置文件失败: %v", err)
	}

	return nil
}

// getDefaultEnvironment 获取默认环境变量配置
func (r *EnvironmentRepositoryImpl) getDefaultEnvironment() *model.Environment {
	homeDir, _ := os.UserHomeDir()
	return &model.Environment{
		GOROOT: "",
		GOPATH: filepath.Join(homeDir, "go"),
		GOBIN:  filepath.Join(homeDir, "go", "bin"),
	}
}

// Get 获取当前环境变量配置
func (r *EnvironmentRepositoryImpl) Get() (*model.Environment, error) {
	return r.loadEnvironment()
}

// Save 保存环境变量配置
func (r *EnvironmentRepositoryImpl) Save(env *model.Environment) error {
	return r.saveEnvironment(env)
}

// UpdatePath 更新系统PATH环境变量（已废弃，使用符号链接方式）
func (r *EnvironmentRepositoryImpl) UpdatePath(goBin string) error {
	// 符号链接方式不再需要手动更新PATH
	return nil
}
