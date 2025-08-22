package repository

import (
	"version-list/internal/domain/model"
)

// VersionRepository 定义版本仓库接口
type VersionRepository interface {
	// Save 保存一个Go版本
	Save(version *model.GoVersion) error
	// Update 更新Go版本信息
	Update(version *model.GoVersion) error
	// FindByVersion 根据版本号查找Go版本
	FindByVersion(version string) (*model.GoVersion, error)
	// FindAll 获取所有Go版本
	FindAll() ([]*model.GoVersion, error)
	// FindActive 获取当前激活的Go版本
	FindActive() (*model.GoVersion, error)
	// FindBySource 根据安装来源查找版本
	FindBySource(source model.InstallSource) ([]*model.GoVersion, error)
	// FindByTag 根据标签查找版本
	FindByTag(tag string) ([]*model.GoVersion, error)
	// SetActive 设置指定版本为激活状态
	SetActive(version string) error
	// Remove 删除指定版本
	Remove(version string) error
	// UpdateLastUsed 更新最后使用时间
	UpdateLastUsed(version string) error
	// GetStatistics 获取统计信息
	GetStatistics() (*model.VersionStatistics, error)
	// FindWithFilter 根据过滤器查找版本
	FindWithFilter(filter *model.VersionFilter) ([]*model.GoVersion, error)
	// FindWithSort 根据排序器查找版本
	FindWithSort(sorter *model.VersionSorter) ([]*model.GoVersion, error)
	// FindWithFilterAndSort 根据过滤器和排序器查找版本
	FindWithFilterAndSort(filter *model.VersionFilter, sorter *model.VersionSorter) ([]*model.GoVersion, error)
	// ExportVersions 导出版本数据
	ExportVersions(exportPath string) (*model.VersionExport, error)
	// ImportVersions 导入版本数据
	ImportVersions(importData *model.VersionImport) error
	// BackupVersions 备份版本数据
	BackupVersions(backupPath string) error
	// RestoreVersions 恢复版本数据
	RestoreVersions(backupPath string) error
	// CleanupStaleVersions 清理过时版本
	CleanupStaleVersions(days int) ([]string, error)
}

// EnvironmentRepository 定义环境变量仓库接口
type EnvironmentRepository interface {
	// Get 获取当前环境变量配置
	Get() (*model.Environment, error)
	// Save 保存环境变量配置
	Save(env *model.Environment) error
	// UpdatePath 更新系统PATH环境变量
	UpdatePath(goPath string) error
}
