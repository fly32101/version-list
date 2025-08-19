package application

import (
	"version-list/internal/domain/service"
	"version-list/internal/infrastructure/persistence"
)

// VersionAppService 版本管理应用服务
type VersionAppService struct {
	versionService *service.VersionService
}

// NewVersionAppService 创建版本管理应用服务实例
func NewVersionAppService() (*VersionAppService, error) {
	// 初始化基础设施层
	versionRepo, err := persistence.NewVersionRepositoryImpl()
	if err != nil {
		return nil, err
	}

	environmentRepo, err := persistence.NewEnvironmentRepositoryImpl()
	if err != nil {
		return nil, err
	}

	// 初始化领域服务
	versionService := service.NewVersionService(versionRepo, environmentRepo)

	return &VersionAppService{
		versionService: versionService,
	}, nil
}

// Install 安装指定版本的Go
func (s *VersionAppService) Install(version string) error {
	return s.versionService.Install(version)
}

// List 列出所有已安装的Go版本
func (s *VersionAppService) List() ([]*service.VersionInfo, error) {
	versions, err := s.versionService.List()
	if err != nil {
		return nil, err
	}

	// 转换为视图模型
	var result []*service.VersionInfo
	for _, v := range versions {
		result = append(result, &service.VersionInfo{
			Version:  v.Version,
			Path:     v.Path,
			IsActive: v.IsActive,
		})
	}

	return result, nil
}

// Use 切换到指定版本的Go
func (s *VersionAppService) Use(version string) error {
	return s.versionService.Use(version)
}

// Current 获取当前使用的Go版本
func (s *VersionAppService) Current() (*service.VersionInfo, error) {
	version, err := s.versionService.Current()
	if err != nil {
		return nil, err
	}

	return &service.VersionInfo{
		Version:  version.Version,
		Path:     version.Path,
		IsActive: version.IsActive,
	}, nil
}

// Remove 移除指定版本的Go
func (s *VersionAppService) Remove(version string) error {
	return s.versionService.Remove(version)
}

// ImportLocal 导入本地已安装的Go版本
func (s *VersionAppService) ImportLocal(path string) (string, error) {
	return s.versionService.ImportLocal(path)
}
