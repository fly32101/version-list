package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"version-list/internal/domain/model"
)

const (
	versionsFile = "versions.json"
	configDir    = ".go-version"
)

// VersionRepositoryImpl 版本仓库实现
type VersionRepositoryImpl struct {
	configPath string
}

// NewVersionRepositoryImpl 创建版本仓库实现实例
func NewVersionRepositoryImpl() (*VersionRepositoryImpl, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %v", err)
	}

	configPath := filepath.Join(homeDir, configDir)
	// 确保配置目录存在
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return nil, fmt.Errorf("创建配置目录失败: %v", err)
	}

	return &VersionRepositoryImpl{
		configPath: configPath,
	}, nil
}

// getVersionsFilePath 获取版本文件路径
func (r *VersionRepositoryImpl) getVersionsFilePath() string {
	return filepath.Join(r.configPath, versionsFile)
}

// loadVersions 加载所有版本数据
func (r *VersionRepositoryImpl) loadVersions() ([]*model.GoVersion, error) {
	filePath := r.getVersionsFilePath()

	// 如果文件不存在，返回空列表
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []*model.GoVersion{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取版本文件失败: %v", err)
	}

	var versions []*model.GoVersion
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, fmt.Errorf("解析版本数据失败: %v", err)
	}

	return versions, nil
}

// saveVersions 保存所有版本数据
func (r *VersionRepositoryImpl) saveVersions(versions []*model.GoVersion) error {
	filePath := r.getVersionsFilePath()

	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化版本数据失败: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入版本文件失败: %v", err)
	}

	return nil
}

// Save 保存一个Go版本
func (r *VersionRepositoryImpl) Save(version *model.GoVersion) error {
	versions, err := r.loadVersions()
	if err != nil {
		return err
	}

	// 设置时间戳
	now := time.Now()
	version.CreatedAt = now
	version.UpdatedAt = now

	versions = append(versions, version)
	return r.saveVersions(versions)
}

// FindByVersion 根据版本号查找Go版本
func (r *VersionRepositoryImpl) FindByVersion(version string) (*model.GoVersion, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	for _, v := range versions {
		if v.Version == version {
			return v, nil
		}
	}

	return nil, fmt.Errorf("未找到版本 %s", version)
}

// FindAll 获取所有Go版本
func (r *VersionRepositoryImpl) FindAll() ([]*model.GoVersion, error) {
	return r.loadVersions()
}

// FindActive 获取当前激活的Go版本
func (r *VersionRepositoryImpl) FindActive() (*model.GoVersion, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	for _, v := range versions {
		if v.IsActive {
			return v, nil
		}
	}

	return nil, fmt.Errorf("未找到激活的Go版本")
}

// SetActive 设置指定版本为激活状态
func (r *VersionRepositoryImpl) SetActive(version string) error {
	versions, err := r.loadVersions()
	if err != nil {
		return err
	}

	found := false
	now := time.Now()
	for _, v := range versions {
		if v.Version == version {
			v.IsActive = true
			v.UpdatedAt = now
			found = true
		} else {
			v.IsActive = false
		}
	}

	if !found {
		return fmt.Errorf("未找到版本 %s", version)
	}

	return r.saveVersions(versions)
}

// Remove 删除指定版本
func (r *VersionRepositoryImpl) Remove(version string) error {
	versions, err := r.loadVersions()
	if err != nil {
		return err
	}

	var newVersions []*model.GoVersion
	found := false
	for _, v := range versions {
		if v.Version == version {
			found = true
		} else {
			newVersions = append(newVersions, v)
		}
	}

	if !found {
		return fmt.Errorf("未找到版本 %s", version)
	}

	return r.saveVersions(newVersions)
}
