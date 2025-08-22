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

// FindWithFilter 根据过滤器查找版本
func (r *VersionRepositoryImpl) FindWithFilter(filter *model.VersionFilter) ([]*model.GoVersion, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	return model.FilterVersions(versions, filter), nil
}

// FindWithSort 根据排序器查找版本
func (r *VersionRepositoryImpl) FindWithSort(sorter *model.VersionSorter) ([]*model.GoVersion, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	return model.SortVersions(versions, sorter), nil
}

// FindWithFilterAndSort 根据过滤器和排序器查找版本
func (r *VersionRepositoryImpl) FindWithFilterAndSort(filter *model.VersionFilter, sorter *model.VersionSorter) ([]*model.GoVersion, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	filtered := model.FilterVersions(versions, filter)
	return model.SortVersions(filtered, sorter), nil
}

// ExportVersions 导出版本数据
func (r *VersionRepositoryImpl) ExportVersions(exportPath string) (*model.VersionExport, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	stats := model.CalculateStatistics(versions)

	export := &model.VersionExport{
		ExportedAt: time.Now(),
		ExportedBy: "go-version-manager",
		Version:    "1.0.0",
		Versions:   versions,
		Statistics: stats,
	}

	// 保存到文件
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化导出数据失败: %v", err)
	}

	if err := os.WriteFile(exportPath, data, 0644); err != nil {
		return nil, fmt.Errorf("写入导出文件失败: %v", err)
	}

	return export, nil
}

// ImportVersions 导入版本数据
func (r *VersionRepositoryImpl) ImportVersions(importData *model.VersionImport) error {
	existingVersions, err := r.loadVersions()
	if err != nil {
		return err
	}

	// 创建现有版本的映射
	existingMap := make(map[string]*model.GoVersion)
	for _, v := range existingVersions {
		existingMap[v.Version] = v
	}

	var finalVersions []*model.GoVersion
	finalVersions = append(finalVersions, existingVersions...)

	// 处理导入的版本
	for _, importVersion := range importData.Versions {
		result := model.ImportResult{
			Version: importVersion.Version,
		}

		if existing, exists := existingMap[importVersion.Version]; exists {
			// 版本已存在，根据冲突模式处理
			switch importData.ConflictMode {
			case "skip":
				result.Status = "skipped"
				result.Message = "版本已存在，跳过导入"
			case "overwrite":
				// 替换现有版本
				for i, v := range finalVersions {
					if v.Version == importVersion.Version {
						finalVersions[i] = importVersion
						break
					}
				}
				result.Status = "success"
				result.Message = "版本已覆盖"
			case "merge":
				// 合并版本信息
				merged := existing.Clone()
				if importVersion.DownloadInfo != nil {
					merged.DownloadInfo = importVersion.DownloadInfo
				}
				if importVersion.ExtractInfo != nil {
					merged.ExtractInfo = importVersion.ExtractInfo
				}
				if importVersion.ValidationInfo != nil {
					merged.ValidationInfo = importVersion.ValidationInfo
				}
				// 合并标签
				for _, tag := range importVersion.Tags {
					merged.AddTag(tag)
				}

				for i, v := range finalVersions {
					if v.Version == importVersion.Version {
						finalVersions[i] = merged
						break
					}
				}
				result.Status = "success"
				result.Message = "版本已合并"
			default:
				result.Status = "conflict"
				result.Message = "版本冲突，需要指定冲突处理模式"
			}
		} else {
			// 新版本，直接添加
			finalVersions = append(finalVersions, importVersion)
			result.Status = "success"
			result.Message = "版本已导入"
		}

		importData.ImportResults = append(importData.ImportResults, result)
	}

	return r.saveVersions(finalVersions)
}

// BackupVersions 备份版本数据
func (r *VersionRepositoryImpl) BackupVersions(backupPath string) error {
	export, err := r.ExportVersions(backupPath)
	if err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}

	// 记录备份信息
	fmt.Printf("备份完成: %s\n", backupPath)
	fmt.Printf("备份时间: %s\n", export.ExportedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("版本数量: %d\n", len(export.Versions))

	return nil
}

// RestoreVersions 恢复版本数据
func (r *VersionRepositoryImpl) RestoreVersions(backupPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %v", err)
	}

	var export model.VersionExport
	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("解析备份数据失败: %v", err)
	}

	// 直接恢复所有版本（覆盖现有数据）
	if err := r.saveVersions(export.Versions); err != nil {
		return fmt.Errorf("恢复版本数据失败: %v", err)
	}

	fmt.Printf("恢复完成: %s\n", backupPath)
	fmt.Printf("恢复版本数量: %d\n", len(export.Versions))

	return nil
}

// CleanupStaleVersions 清理过时版本
func (r *VersionRepositoryImpl) CleanupStaleVersions(days int) ([]string, error) {
	versions, err := r.loadVersions()
	if err != nil {
		return nil, err
	}

	var cleanedVersions []string
	var remainingVersions []*model.GoVersion

	for _, v := range versions {
		if v.IsActive {
			// 不清理激活版本
			remainingVersions = append(remainingVersions, v)
		} else if v.IsStale(days) {
			cleanedVersions = append(cleanedVersions, v.Version)
		} else {
			remainingVersions = append(remainingVersions, v)
		}
	}

	if len(cleanedVersions) > 0 {
		if err := r.saveVersions(remainingVersions); err != nil {
			return nil, fmt.Errorf("保存清理后的版本数据失败: %v", err)
		}
	}

	return cleanedVersions, nil
}
