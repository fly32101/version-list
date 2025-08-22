package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GoVersion 表示一个Go版本
type GoVersion struct {
	Version         string          // 版本号，如 "1.21.0"
	Path            string          // 安装路径
	IsActive        bool            // 是否为当前激活版本
	Source          InstallSource   // 安装来源
	DownloadInfo    *DownloadInfo   // 下载信息（仅在线安装时有值）
	ExtractInfo     *ExtractInfo    // 解压信息
	ValidationInfo  *ValidationInfo // 验证信息
	SystemInfo      *SystemInfo     // 系统信息
	CreatedAt       time.Time       // 创建时间
	UpdatedAt       time.Time       // 更新时间
	LastUsedAt      *time.Time      // 最后使用时间
	InstallDuration time.Duration   // 安装耗时
	Tags            []string        // 标签（如 "stable", "beta", "rc"）
	Notes           string          // 备注信息
}

// Environment 表示环境变量配置
type Environment struct {
	GOROOT string // Go安装目录
	GOPATH string // 工作目录
	GOBIN  string // 二进制文件目录
}

// IsOnlineInstalled 检查是否为在线安装
func (v *GoVersion) IsOnlineInstalled() bool {
	return v.Source == SourceOnline && v.DownloadInfo != nil
}

// IsValid 检查版本是否有效
func (v *GoVersion) IsValid() bool {
	if v.ValidationInfo == nil {
		return false
	}
	return v.ValidationInfo.ChecksumValid &&
		v.ValidationInfo.ExecutableValid &&
		v.ValidationInfo.VersionValid
}

// GetDisplayName 获取显示名称
func (v *GoVersion) GetDisplayName() string {
	if len(v.Tags) > 0 {
		return v.Version + " (" + v.Tags[0] + ")"
	}
	return v.Version
}

// GetSizeInfo 获取大小信息
func (v *GoVersion) GetSizeInfo() (downloadSize, extractedSize int64) {
	if v.DownloadInfo != nil {
		downloadSize = v.DownloadInfo.Size
	}
	if v.ExtractInfo != nil {
		extractedSize = v.ExtractInfo.ExtractedSize
	}
	return
}

// MarkAsUsed 标记为已使用
func (v *GoVersion) MarkAsUsed() {
	now := time.Now()
	v.LastUsedAt = &now
	v.UpdatedAt = now
}

// AddTag 添加标签
func (v *GoVersion) AddTag(tag string) {
	for _, existingTag := range v.Tags {
		if existingTag == tag {
			return // 标签已存在
		}
	}
	v.Tags = append(v.Tags, tag)
	v.UpdatedAt = time.Now()
}

// RemoveTag 移除标签
func (v *GoVersion) RemoveTag(tag string) {
	for i, existingTag := range v.Tags {
		if existingTag == tag {
			v.Tags = append(v.Tags[:i], v.Tags[i+1:]...)
			v.UpdatedAt = time.Now()
			return
		}
	}
}

// HasTag 检查是否有指定标签
func (v *GoVersion) HasTag(tag string) bool {
	for _, existingTag := range v.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// GetAge 获取版本的年龄
func (v *GoVersion) GetAge() time.Duration {
	return time.Since(v.CreatedAt)
}

// GetUsageFrequency 获取使用频率（基于最后使用时间）
func (v *GoVersion) GetUsageFrequency() string {
	if v.LastUsedAt == nil {
		return "never"
	}

	daysSinceLastUse := int(time.Since(*v.LastUsedAt).Hours() / 24)

	switch {
	case daysSinceLastUse == 0:
		return "today"
	case daysSinceLastUse <= 7:
		return "this_week"
	case daysSinceLastUse <= 30:
		return "this_month"
	case daysSinceLastUse <= 90:
		return "this_quarter"
	default:
		return "rarely"
	}
}

// IsStale 检查版本是否过时（超过指定天数未使用）
func (v *GoVersion) IsStale(days int) bool {
	if v.LastUsedAt == nil {
		return true
	}
	return time.Since(*v.LastUsedAt) > time.Duration(days)*24*time.Hour
}

// GetInstallationEfficiency 获取安装效率（下载大小vs解压大小的比率）
func (v *GoVersion) GetInstallationEfficiency() float64 {
	if v.DownloadInfo == nil || v.ExtractInfo == nil {
		return 0
	}
	if v.DownloadInfo.Size == 0 {
		return 0
	}
	return float64(v.ExtractInfo.ExtractedSize) / float64(v.DownloadInfo.Size)
}

// GetDownloadSpeed 获取下载速度的人性化表示
func (v *GoVersion) GetDownloadSpeed() string {
	if v.DownloadInfo == nil || v.DownloadInfo.Speed == 0 {
		return "unknown"
	}
	return formatBytes(int64(v.DownloadInfo.Speed)) + "/s"
}

// GetInstallationDate 获取安装日期的字符串表示
func (v *GoVersion) GetInstallationDate() string {
	return v.CreatedAt.Format("2006-01-02 15:04:05")
}

// GetLastUsedDate 获取最后使用日期的字符串表示
func (v *GoVersion) GetLastUsedDate() string {
	if v.LastUsedAt == nil {
		return "never"
	}
	return v.LastUsedAt.Format("2006-01-02 15:04:05")
}

// CompareVersion 比较两个版本号
func (v *GoVersion) CompareVersion(other *GoVersion) *VersionComparison {
	return CompareVersionStrings(v.Version, other.Version)
}

// Matches 检查版本是否匹配过滤器
func (v *GoVersion) Matches(filter *VersionFilter) bool {
	if filter == nil {
		return true
	}

	// 检查安装来源
	if filter.Source != nil && v.Source != *filter.Source {
		return false
	}

	// 检查激活状态
	if filter.IsActive != nil && v.IsActive != *filter.IsActive {
		return false
	}

	// 检查标签
	if len(filter.Tags) > 0 {
		hasAnyTag := false
		for _, tag := range filter.Tags {
			if v.HasTag(tag) {
				hasAnyTag = true
				break
			}
		}
		if !hasAnyTag {
			return false
		}
	}

	// 检查大小
	_, extractedSize := v.GetSizeInfo()
	if filter.MinSize != nil && extractedSize < *filter.MinSize {
		return false
	}
	if filter.MaxSize != nil && extractedSize > *filter.MaxSize {
		return false
	}

	// 检查创建时间
	if filter.CreatedAfter != nil && v.CreatedAt.Before(*filter.CreatedAfter) {
		return false
	}
	if filter.CreatedBefore != nil && v.CreatedAt.After(*filter.CreatedBefore) {
		return false
	}

	// 检查使用时间
	if filter.UsedAfter != nil {
		if v.LastUsedAt == nil || v.LastUsedAt.Before(*filter.UsedAfter) {
			return false
		}
	}

	// 检查版本号模式
	if filter.Pattern != "" {
		matched, err := matchPattern(v.Version, filter.Pattern)
		if err != nil || !matched {
			return false
		}
	}

	return true
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// matchPattern 简单的模式匹配
func matchPattern(version, pattern string) (bool, error) {
	// 简单的通配符匹配实现
	if pattern == "*" {
		return true, nil
	}

	// 检查前缀匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(version, prefix), nil
	}

	// 检查后缀匹配
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(version, suffix), nil
	}

	// 精确匹配
	return version == pattern, nil
}

// CompareVersionStrings 比较两个版本字符串
func CompareVersionStrings(v1, v2 string) *VersionComparison {
	result := &VersionComparison{
		Version1: v1,
		Version2: v2,
	}

	// 简单的版本比较实现
	// 实际项目中可能需要更复杂的语义版本比较
	if v1 == v2 {
		result.Result = 0
		return result
	}

	// 按点分割版本号
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		var err error

		if i < len(parts1) {
			p1, err = strconv.Atoi(parts1[i])
			if err != nil {
				result.Error = fmt.Sprintf("无效的版本号格式: %s", v1)
				return result
			}
		}

		if i < len(parts2) {
			p2, err = strconv.Atoi(parts2[i])
			if err != nil {
				result.Error = fmt.Sprintf("无效的版本号格式: %s", v2)
				return result
			}
		}

		if p1 < p2 {
			result.Result = -1
			return result
		} else if p1 > p2 {
			result.Result = 1
			return result
		}
	}

	result.Result = 0
	return result
}

// FilterVersions 过滤版本列表
func FilterVersions(versions []*GoVersion, filter *VersionFilter) []*GoVersion {
	if filter == nil {
		return versions
	}

	var filtered []*GoVersion
	for _, version := range versions {
		if version.Matches(filter) {
			filtered = append(filtered, version)
		}
	}

	return filtered
}

// SortVersions 排序版本列表
func SortVersions(versions []*GoVersion, sorter *VersionSorter) []*GoVersion {
	if sorter == nil {
		return versions
	}

	// 创建副本以避免修改原始切片
	sorted := make([]*GoVersion, len(versions))
	copy(sorted, versions)

	// 根据排序字段和方向进行排序
	switch sorter.Field {
	case "version":
		if sorter.Direction == "desc" {
			// 降序排列
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					comp := CompareVersionStrings(sorted[i].Version, sorted[j].Version)
					if comp.Result < 0 {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			// 升序排列
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					comp := CompareVersionStrings(sorted[i].Version, sorted[j].Version)
					if comp.Result > 0 {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	case "created_at":
		if sorter.Direction == "desc" {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].CreatedAt.Before(sorted[j].CreatedAt) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].CreatedAt.After(sorted[j].CreatedAt) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	case "size":
		if sorter.Direction == "desc" {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					_, size1 := sorted[i].GetSizeInfo()
					_, size2 := sorted[j].GetSizeInfo()
					if size1 < size2 {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					_, size1 := sorted[i].GetSizeInfo()
					_, size2 := sorted[j].GetSizeInfo()
					if size1 > size2 {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	}

	return sorted
}

// CalculateStatistics 计算版本统计信息
func CalculateStatistics(versions []*GoVersion) *VersionStatistics {
	stats := &VersionStatistics{
		TotalVersions:   len(versions),
		TagStatistics:   make(map[string]int),
		VersionsByMonth: make(map[string]int),
		UsageFrequency:  make(map[string]int),
	}

	if len(versions) == 0 {
		return stats
	}

	var totalInstallTime time.Duration
	var installCount int

	for _, v := range versions {
		// 统计安装来源
		if v.Source == SourceOnline {
			stats.OnlineVersions++
		} else {
			stats.LocalVersions++
		}

		// 记录激活版本
		if v.IsActive {
			stats.ActiveVersion = v.Version
		}

		// 计算磁盘使用量
		_, extractedSize := v.GetSizeInfo()
		stats.TotalDiskUsage += extractedSize

		// 计算平均安装时间
		if v.InstallDuration > 0 {
			totalInstallTime += v.InstallDuration
			installCount++
		}

		// 统计标签
		for _, tag := range v.Tags {
			stats.TagStatistics[tag]++
		}

		// 按月份统计
		monthKey := v.CreatedAt.Format("2006-01")
		stats.VersionsByMonth[monthKey]++

		// 使用频率统计
		frequency := v.GetUsageFrequency()
		stats.UsageFrequency[frequency]++

		// 找最近使用的版本
		if v.LastUsedAt != nil {
			if stats.MostRecentlyUsed == "" {
				stats.MostRecentlyUsed = v.Version
			} else {
				// 找到最近使用的版本
				for _, other := range versions {
					if other.Version == stats.MostRecentlyUsed && other.LastUsedAt != nil {
						if v.LastUsedAt.After(*other.LastUsedAt) {
							stats.MostRecentlyUsed = v.Version
						}
						break
					}
				}
			}
		}

		// 找最老和最新的版本
		if stats.OldestVersion == "" || v.CreatedAt.Before(findVersionCreatedAt(versions, stats.OldestVersion)) {
			stats.OldestVersion = v.Version
		}
		if stats.NewestVersion == "" || v.CreatedAt.After(findVersionCreatedAt(versions, stats.NewestVersion)) {
			stats.NewestVersion = v.Version
		}
	}

	if installCount > 0 {
		stats.AverageInstallTime = totalInstallTime / time.Duration(installCount)
	}

	return stats
}

// findVersionCreatedAt 辅助函数：根据版本名查找创建时间
func findVersionCreatedAt(versions []*GoVersion, versionName string) time.Time {
	for _, v := range versions {
		if v.Version == versionName {
			return v.CreatedAt
		}
	}
	return time.Time{}
}

// VersionStatistics 版本统计信息
type VersionStatistics struct {
	TotalVersions      int            // 总版本数
	OnlineVersions     int            // 在线安装版本数
	LocalVersions      int            // 本地导入版本数
	ActiveVersion      string         // 当前激活版本
	TotalDiskUsage     int64          // 总磁盘使用量
	MostRecentlyUsed   string         // 最近使用的版本
	OldestVersion      string         // 最老的版本
	NewestVersion      string         // 最新的版本
	AverageInstallTime time.Duration  // 平均安装时间
	TagStatistics      map[string]int // 标签统计
	VersionsByMonth    map[string]int // 按月份统计的版本数
	UsageFrequency     map[string]int // 使用频率统计
}

// VersionComparison 版本比较结果
type VersionComparison struct {
	Version1 string // 版本1
	Version2 string // 版本2
	Result   int    // 比较结果: -1(小于), 0(等于), 1(大于)
	Error    string // 错误信息
}

// VersionFilter 版本过滤器
type VersionFilter struct {
	Source        *InstallSource // 按安装来源过滤
	Tags          []string       // 按标签过滤
	IsActive      *bool          // 按激活状态过滤
	MinSize       *int64         // 最小大小过滤
	MaxSize       *int64         // 最大大小过滤
	CreatedAfter  *time.Time     // 创建时间之后
	CreatedBefore *time.Time     // 创建时间之前
	UsedAfter     *time.Time     // 使用时间之后
	Pattern       string         // 版本号模式匹配
}

// VersionSorter 版本排序器
type VersionSorter struct {
	Field     string // 排序字段: version, created_at, updated_at, size, usage
	Direction string // 排序方向: asc, desc
}

// VersionExport 版本导出数据
type VersionExport struct {
	ExportedAt time.Time          `json:"exported_at"`
	ExportedBy string             `json:"exported_by"`
	Version    string             `json:"export_version"`
	Versions   []*GoVersion       `json:"versions"`
	Statistics *VersionStatistics `json:"statistics"`
}

// VersionImport 版本导入数据
type VersionImport struct {
	ImportedAt    time.Time      `json:"imported_at"`
	ImportedBy    string         `json:"imported_by"`
	SourceFile    string         `json:"source_file"`
	Versions      []*GoVersion   `json:"versions"`
	ConflictMode  string         // skip, overwrite, merge
	ImportResults []ImportResult `json:"import_results"`
}

// ImportResult 导入结果
type ImportResult struct {
	Version string `json:"version"`
	Status  string `json:"status"` // success, skipped, failed, conflict
	Message string `json:"message"`
}
