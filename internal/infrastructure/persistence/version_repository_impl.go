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

// Update 更新Go版本信息
func (r *VersionRepositoryImpl) Update(version *model.GoVersion) error {
	versions, err := r.loadVersions()
	if err != nil {
		return err
	}

	found := false
	now := time.Now()
	for i, v := range versions {
		if v.Version == version.Version {
			version.UpdatedAt = now
			versions[i] = version
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到版本 %s", version.Version)
	}

	return r.saveVersions(versions)
}

// FindBySource 根据安装来源查找版本
func (r *VersionRepositoryImpl) FindBySource(source model.InstallSource) ([]*model.GoVersion, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	var result []*model.GoVersion
	for _, v := range versions {
		if v.Source == source {
			result = append(result, v)
		}
	}

	return result, nil
}

// FindByTag 根据标签查找版本
func (r *VersionRepositoryImpl) FindByTag(tag string) ([]*model.GoVersion, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	var result []*model.GoVersion
	for _, v := range versions {
		if v.HasTag(tag) {
			result = append(result, v)
		}
	}

	return result, nil
}

// UpdateLastUsed 更新最后使用时间
func (r *VersionRepositoryImpl) UpdateLastUsed(version string) error {
	versions, err := r.loadVersions()
	if err != nil {
		return err
	}

	found := false
	for _, v := range versions {
		if v.Version == version {
			v.MarkAsUsed()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到版本 %s", version)
	}

	return r.saveVersions(versions)
}

// GetStatistics 获取统计信息
func (r *VersionRepositoryImpl) GetStatistics() (*model.VersionStatistics, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	stats := &model.VersionStatistics{
		TotalVersions: len(versions),
	}

	var activeVersion string
	var mostRecentlyUsed string
	var oldestVersion string
	var newestVersion string
	var totalInstallTime time.Duration
	var installCount int

	for _, v := range versions {
		// 统计安装来源
		if v.Source == model.SourceOnline {
			stats.OnlineVersions++
		} else {
			stats.LocalVersions++
		}

		// 记录激活版本
		if v.IsActive {
			activeVersion = v.Version
		}

		// 计算磁盘使用量
		if v.ExtractInfo != nil {
			stats.TotalDiskUsage += v.ExtractInfo.ExtractedSize
		}

		// 计算平均安装时间
		if v.InstallDuration > 0 {
			totalInstallTime += v.InstallDuration
			installCount++
		}

		// 找最近使用的版本
		if v.LastUsedAt != nil {
			if mostRecentlyUsed == "" {
				mostRecentlyUsed = v.Version
			}
		}

		// 找最老和最新的版本（基于创建时间）
		if oldestVersion == "" || v.CreatedAt.Before(findVersionByName(versions, oldestVersion).CreatedAt) {
			oldestVersion = v.Version
		}
		if newestVersion == "" || v.CreatedAt.After(findVersionByName(versions, newestVersion).CreatedAt) {
			newestVersion = v.Version
		}
	}

	stats.ActiveVersion = activeVersion
	stats.MostRecentlyUsed = mostRecentlyUsed
	stats.OldestVersion = oldestVersion
	stats.NewestVersion = newestVersion

	if installCount > 0 {
		stats.AverageInstallTime = totalInstallTime / time.Duration(installCount)
	}

	return stats, nil
}

// findVersionByName 辅助函数：根据版本名查找版本
func findVersionByName(versions []*model.GoVersion, name string) *model.GoVersion {
	for _, v := range versions {
		if v.Version == name {
			return v
		}
	}
	return nil
}
