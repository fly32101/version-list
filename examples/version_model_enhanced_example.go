package main

import (
	"fmt"
	"log"
	"time"

	"version-list/internal/domain/model"
)

func main() {
	fmt.Println("=== Go Version Manager - 增强数据模型示例 ===\n")

	// 1. 创建测试版本数据
	fmt.Println("1. 创建测试版本数据:")
	versions := createTestVersions()
	fmt.Printf("   创建了 %d 个测试版本\n", len(versions))
	for _, v := range versions {
		fmt.Printf("   - %s\n", v.Summary())
	}
	fmt.Println()

	// 2. 版本比较
	fmt.Println("2. 版本比较:")
	comp1 := model.CompareVersionStrings("1.21.0", "1.20.0")
	fmt.Printf("   1.21.0 vs 1.20.0: %s\n", formatComparisonResult(comp1))

	comp2 := model.CompareVersionStrings("1.19.5", "1.21.0")
	fmt.Printf("   1.19.5 vs 1.21.0: %s\n", formatComparisonResult(comp2))

	comp3 := model.CompareVersionStrings("1.21.0", "1.21.0")
	fmt.Printf("   1.21.0 vs 1.21.0: %s\n", formatComparisonResult(comp3))
	fmt.Println()

	// 3. 版本过滤
	fmt.Println("3. 版本过滤:")

	// 按安装来源过滤
	onlineSource := model.SourceOnline
	filter1 := &model.VersionFilter{
		Source: &onlineSource,
	}
	onlineVersions := model.FilterVersions(versions, filter1)
	fmt.Printf("   在线安装版本: %d 个\n", len(onlineVersions))
	for _, v := range onlineVersions {
		fmt.Printf("     - %s\n", v.Version)
	}

	// 按标签过滤
	filter2 := &model.VersionFilter{
		Tags: []string{"stable"},
	}
	stableVersions := model.FilterVersions(versions, filter2)
	fmt.Printf("   稳定版本: %d 个\n", len(stableVersions))
	for _, v := range stableVersions {
		fmt.Printf("     - %s [%s]\n", v.Version, joinTags(v.Tags))
	}

	// 按大小过滤
	minSize := int64(400 * 1024 * 1024) // 400MB
	filter3 := &model.VersionFilter{
		MinSize: &minSize,
	}
	largeVersions := model.FilterVersions(versions, filter3)
	fmt.Printf("   大于400MB的版本: %d 个\n", len(largeVersions))
	for _, v := range largeVersions {
		_, size := v.GetSizeInfo()
		fmt.Printf("     - %s (%s)\n", v.Version, formatBytes(size))
	}
	fmt.Println()

	// 4. 版本排序
	fmt.Println("4. 版本排序:")

	// 按版本号排序
	sorter1 := &model.VersionSorter{
		Field:     "version",
		Direction: "desc",
	}
	sortedByVersion := model.SortVersions(versions, sorter1)
	fmt.Printf("   按版本号降序:\n")
	for _, v := range sortedByVersion {
		fmt.Printf("     - %s\n", v.Version)
	}

	// 按创建时间排序
	sorter2 := &model.VersionSorter{
		Field:     "created_at",
		Direction: "desc",
	}
	sortedByTime := model.SortVersions(versions, sorter2)
	fmt.Printf("   按创建时间降序:\n")
	for _, v := range sortedByTime {
		fmt.Printf("     - %s (%s)\n", v.Version, v.GetInstallationDate())
	}
	fmt.Println()

	// 5. 版本统计
	fmt.Println("5. 版本统计:")
	stats := model.CalculateStatistics(versions)
	fmt.Printf("   总版本数: %d\n", stats.TotalVersions)
	fmt.Printf("   在线安装: %d\n", stats.OnlineVersions)
	fmt.Printf("   本地导入: %d\n", stats.LocalVersions)
	fmt.Printf("   当前激活: %s\n", stats.ActiveVersion)
	fmt.Printf("   总磁盘使用: %s\n", formatBytes(stats.TotalDiskUsage))
	fmt.Printf("   平均安装时间: %s\n", stats.AverageInstallTime)

	fmt.Printf("   标签统计:\n")
	for tag, count := range stats.TagStatistics {
		fmt.Printf("     - %s: %d\n", tag, count)
	}

	fmt.Printf("   使用频率统计:\n")
	for freq, count := range stats.UsageFrequency {
		fmt.Printf("     - %s: %d\n", freq, count)
	}
	fmt.Println()

	// 6. 版本详细信息
	fmt.Println("6. 版本详细信息:")
	for _, v := range versions {
		if v.IsActive {
			fmt.Printf("   当前激活版本: %s\n", v.Version)
			fmt.Printf("     年龄: %s\n", formatDuration(v.GetAge()))
			fmt.Printf("     使用频率: %s\n", v.GetUsageFrequency())
			fmt.Printf("     安装效率: %.2f\n", v.GetInstallationEfficiency())
			fmt.Printf("     下载速度: %s\n", v.GetDownloadSpeed())
			fmt.Printf("     最后使用: %s\n", v.GetLastUsedDate())
			fmt.Printf("     是否过时(30天): %t\n", v.IsStale(30))
			break
		}
	}
	fmt.Println()

	// 7. 版本序列化
	fmt.Println("7. 版本序列化:")
	activeVersion := findActiveVersion(versions)
	if activeVersion != nil {
		jsonStr, err := activeVersion.ToJSON()
		if err != nil {
			log.Printf("序列化失败: %v", err)
		} else {
			fmt.Printf("   JSON长度: %d 字符\n", len(jsonStr))

			// 反序列化测试
			restored, err := model.GoVersionFromJSON(jsonStr)
			if err != nil {
				log.Printf("反序列化失败: %v", err)
			} else {
				fmt.Printf("   反序列化成功: %s\n", restored.Version)
			}
		}

		// 转换为Map
		versionMap := activeVersion.ToMap()
		fmt.Printf("   Map字段数: %d\n", len(versionMap))

		// 克隆测试
		cloned := activeVersion.Clone()
		fmt.Printf("   克隆成功: %s\n", cloned.Version)

		// 验证测试
		if err := activeVersion.Validate(); err != nil {
			fmt.Printf("   验证失败: %v\n", err)
		} else {
			fmt.Printf("   验证通过\n")
		}
	}
	fmt.Println()

	// 8. 版本导出数据结构
	fmt.Println("8. 版本导出数据结构:")
	export := &model.VersionExport{
		ExportedAt: time.Now(),
		ExportedBy: "example-user",
		Version:    "1.0.0",
		Versions:   versions,
		Statistics: stats,
	}
	fmt.Printf("   导出时间: %s\n", export.ExportedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   导出用户: %s\n", export.ExportedBy)
	fmt.Printf("   导出版本数: %d\n", len(export.Versions))
	fmt.Println()

	fmt.Println("=== 增强数据模型示例完成 ===")
}

// createTestVersions 创建测试版本数据
func createTestVersions() []*model.GoVersion {
	now := time.Now()

	return []*model.GoVersion{
		{
			Version:  "1.21.0",
			Path:     "/home/user/.go/versions/1.21.0",
			IsActive: true,
			Source:   model.SourceOnline,
			DownloadInfo: &model.DownloadInfo{
				URL:          "https://golang.org/dl/go1.21.0.linux-amd64.tar.gz",
				Filename:     "go1.21.0.linux-amd64.tar.gz",
				Size:         163553657,
				Checksum:     "abc123def456",
				ChecksumType: "sha256",
				DownloadedAt: now.Add(-2 * time.Hour),
				Duration:     30000,
				Speed:        5451788.57,
			},
			ExtractInfo: &model.ExtractInfo{
				ArchiveSize:   163553657,
				ExtractedSize: 500000000,
				FileCount:     1500,
				Duration:      5 * time.Second,
				RootDir:       "go",
			},
			ValidationInfo: &model.ValidationInfo{
				ChecksumValid:   true,
				ExecutableValid: true,
				VersionValid:    true,
			},
			SystemInfo: &model.SystemInfo{
				OS:       "linux",
				Arch:     "amd64",
				Version:  "1.21.0",
				Filename: "go1.21.0.linux-amd64.tar.gz",
				URL:      "https://golang.org/dl/go1.21.0.linux-amd64.tar.gz",
			},
			CreatedAt:       now.Add(-2 * time.Hour),
			UpdatedAt:       now.Add(-1 * time.Hour),
			LastUsedAt:      &now,
			InstallDuration: 35 * time.Second,
			Tags:            []string{"stable", "lts"},
			Notes:           "当前使用的稳定版本",
		},
		{
			Version:  "1.20.7",
			Path:     "/home/user/.go/versions/1.20.7",
			IsActive: false,
			Source:   model.SourceOnline,
			DownloadInfo: &model.DownloadInfo{
				URL:          "https://golang.org/dl/go1.20.7.linux-amd64.tar.gz",
				Filename:     "go1.20.7.linux-amd64.tar.gz",
				Size:         145000000,
				Checksum:     "def456ghi789",
				ChecksumType: "sha256",
				DownloadedAt: now.Add(-7 * 24 * time.Hour),
				Duration:     25000,
				Speed:        5800000,
			},
			ExtractInfo: &model.ExtractInfo{
				ArchiveSize:   145000000,
				ExtractedSize: 450000000,
				FileCount:     1400,
				Duration:      4 * time.Second,
				RootDir:       "go",
			},
			ValidationInfo: &model.ValidationInfo{
				ChecksumValid:   true,
				ExecutableValid: true,
				VersionValid:    true,
			},
			CreatedAt:       now.Add(-7 * 24 * time.Hour),
			UpdatedAt:       now.Add(-7 * 24 * time.Hour),
			LastUsedAt:      timePtr(now.Add(-3 * 24 * time.Hour)),
			InstallDuration: 29 * time.Second,
			Tags:            []string{"stable"},
			Notes:           "之前的稳定版本",
		},
		{
			Version:  "1.19.13",
			Path:     "/home/user/.go/versions/1.19.13",
			IsActive: false,
			Source:   model.SourceLocal,
			ExtractInfo: &model.ExtractInfo{
				ExtractedSize: 380000000,
				FileCount:     1200,
			},
			ValidationInfo: &model.ValidationInfo{
				ChecksumValid:   false,
				ExecutableValid: true,
				VersionValid:    true,
				Error:           "无法验证校验和（本地导入）",
			},
			CreatedAt:  now.Add(-30 * 24 * time.Hour),
			UpdatedAt:  now.Add(-30 * 24 * time.Hour),
			LastUsedAt: timePtr(now.Add(-15 * 24 * time.Hour)),
			Tags:       []string{"legacy"},
			Notes:      "从本地导入的旧版本",
		},
		{
			Version:  "1.22.0-beta1",
			Path:     "/home/user/.go/versions/1.22.0-beta1",
			IsActive: false,
			Source:   model.SourceOnline,
			DownloadInfo: &model.DownloadInfo{
				URL:          "https://golang.org/dl/go1.22.0-beta1.linux-amd64.tar.gz",
				Filename:     "go1.22.0-beta1.linux-amd64.tar.gz",
				Size:         170000000,
				Checksum:     "ghi789jkl012",
				ChecksumType: "sha256",
				DownloadedAt: now.Add(-1 * 24 * time.Hour),
				Duration:     35000,
				Speed:        4857142.86,
			},
			ExtractInfo: &model.ExtractInfo{
				ArchiveSize:   170000000,
				ExtractedSize: 520000000,
				FileCount:     1600,
				Duration:      6 * time.Second,
				RootDir:       "go",
			},
			ValidationInfo: &model.ValidationInfo{
				ChecksumValid:   true,
				ExecutableValid: true,
				VersionValid:    true,
			},
			CreatedAt:       now.Add(-1 * 24 * time.Hour),
			UpdatedAt:       now.Add(-1 * 24 * time.Hour),
			InstallDuration: 41 * time.Second,
			Tags:            []string{"beta", "experimental"},
			Notes:           "测试新功能的beta版本",
		},
	}
}

// formatComparisonResult 格式化比较结果
func formatComparisonResult(comp *model.VersionComparison) string {
	if comp.Error != "" {
		return fmt.Sprintf("错误: %s", comp.Error)
	}

	switch comp.Result {
	case -1:
		return fmt.Sprintf("%s < %s", comp.Version1, comp.Version2)
	case 0:
		return fmt.Sprintf("%s = %s", comp.Version1, comp.Version2)
	case 1:
		return fmt.Sprintf("%s > %s", comp.Version1, comp.Version2)
	default:
		return "未知结果"
	}
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

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) - minutes*60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) - hours*60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

// joinTags 连接标签
func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := tags[0]
	for i := 1; i < len(tags); i++ {
		result += ", " + tags[i]
	}
	return result
}

// findActiveVersion 查找激活版本
func findActiveVersion(versions []*model.GoVersion) *model.GoVersion {
	for _, v := range versions {
		if v.IsActive {
			return v
		}
	}
	return nil
}

// timePtr 辅助函数：返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}
