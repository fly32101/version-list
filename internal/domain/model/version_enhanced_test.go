package model

import (
	"testing"
	"time"
)

func TestGoVersion_GetAge(t *testing.T) {
	version := &GoVersion{
		Version:   "1.21.0",
		CreatedAt: time.Now().Add(-24 * time.Hour), // 24小时前
	}

	age := version.GetAge()
	if age < 23*time.Hour || age > 25*time.Hour {
		t.Errorf("GetAge() = %v, 期望大约24小时", age)
	}
}

func TestGoVersion_GetUsageFrequency(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		lastUsedAt   *time.Time
		expectedFreq string
	}{
		{
			name:         "never used",
			lastUsedAt:   nil,
			expectedFreq: "never",
		},
		{
			name:         "used today",
			lastUsedAt:   &now,
			expectedFreq: "today",
		},
		{
			name:         "used this week",
			lastUsedAt:   timePtr(now.Add(-3 * 24 * time.Hour)),
			expectedFreq: "this_week",
		},
		{
			name:         "used this month",
			lastUsedAt:   timePtr(now.Add(-15 * 24 * time.Hour)),
			expectedFreq: "this_month",
		},
		{
			name:         "used this quarter",
			lastUsedAt:   timePtr(now.Add(-60 * 24 * time.Hour)),
			expectedFreq: "this_quarter",
		},
		{
			name:         "rarely used",
			lastUsedAt:   timePtr(now.Add(-120 * 24 * time.Hour)),
			expectedFreq: "rarely",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := &GoVersion{
				Version:    "1.21.0",
				LastUsedAt: tt.lastUsedAt,
			}

			freq := version.GetUsageFrequency()
			if freq != tt.expectedFreq {
				t.Errorf("GetUsageFrequency() = %s, 期望 %s", freq, tt.expectedFreq)
			}
		})
	}
}

func TestGoVersion_IsStale(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		lastUsedAt *time.Time
		days       int
		expected   bool
	}{
		{
			name:       "never used",
			lastUsedAt: nil,
			days:       30,
			expected:   true,
		},
		{
			name:       "used recently",
			lastUsedAt: &now,
			days:       30,
			expected:   false,
		},
		{
			name:       "used 20 days ago, threshold 30",
			lastUsedAt: timePtr(now.Add(-20 * 24 * time.Hour)),
			days:       30,
			expected:   false,
		},
		{
			name:       "used 40 days ago, threshold 30",
			lastUsedAt: timePtr(now.Add(-40 * 24 * time.Hour)),
			days:       30,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := &GoVersion{
				Version:    "1.21.0",
				LastUsedAt: tt.lastUsedAt,
			}

			isStale := version.IsStale(tt.days)
			if isStale != tt.expected {
				t.Errorf("IsStale(%d) = %v, 期望 %v", tt.days, isStale, tt.expected)
			}
		})
	}
}

func TestGoVersion_GetInstallationEfficiency(t *testing.T) {
	tests := []struct {
		name         string
		downloadInfo *DownloadInfo
		extractInfo  *ExtractInfo
		expected     float64
	}{
		{
			name:         "no info",
			downloadInfo: nil,
			extractInfo:  nil,
			expected:     0,
		},
		{
			name: "normal efficiency",
			downloadInfo: &DownloadInfo{
				Size: 100000,
			},
			extractInfo: &ExtractInfo{
				ExtractedSize: 300000,
			},
			expected: 3.0,
		},
		{
			name: "high compression",
			downloadInfo: &DownloadInfo{
				Size: 50000,
			},
			extractInfo: &ExtractInfo{
				ExtractedSize: 500000,
			},
			expected: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := &GoVersion{
				Version:      "1.21.0",
				DownloadInfo: tt.downloadInfo,
				ExtractInfo:  tt.extractInfo,
			}

			efficiency := version.GetInstallationEfficiency()
			if efficiency != tt.expected {
				t.Errorf("GetInstallationEfficiency() = %f, 期望 %f", efficiency, tt.expected)
			}
		})
	}
}

func TestGoVersion_CompareVersion(t *testing.T) {
	version1 := &GoVersion{Version: "1.21.0"}
	version2 := &GoVersion{Version: "1.20.0"}
	version3 := &GoVersion{Version: "1.21.0"}

	// 测试大于
	comp1 := version1.CompareVersion(version2)
	if comp1.Result != 1 {
		t.Errorf("1.21.0 vs 1.20.0: 期望结果为1，实际为%d", comp1.Result)
	}

	// 测试小于
	comp2 := version2.CompareVersion(version1)
	if comp2.Result != -1 {
		t.Errorf("1.20.0 vs 1.21.0: 期望结果为-1，实际为%d", comp2.Result)
	}

	// 测试等于
	comp3 := version1.CompareVersion(version3)
	if comp3.Result != 0 {
		t.Errorf("1.21.0 vs 1.21.0: 期望结果为0，实际为%d", comp3.Result)
	}
}

func TestGoVersion_Matches(t *testing.T) {
	now := time.Now()
	version := &GoVersion{
		Version:   "1.21.0",
		IsActive:  true,
		Source:    SourceOnline,
		Tags:      []string{"stable", "lts"},
		CreatedAt: now.Add(-24 * time.Hour),
		ExtractInfo: &ExtractInfo{
			ExtractedSize: 500000,
		},
	}

	tests := []struct {
		name     string
		filter   *VersionFilter
		expected bool
	}{
		{
			name:     "nil filter",
			filter:   nil,
			expected: true,
		},
		{
			name: "match source",
			filter: &VersionFilter{
				Source: &[]InstallSource{SourceOnline}[0],
			},
			expected: true,
		},
		{
			name: "no match source",
			filter: &VersionFilter{
				Source: &[]InstallSource{SourceLocal}[0],
			},
			expected: false,
		},
		{
			name: "match active",
			filter: &VersionFilter{
				IsActive: &[]bool{true}[0],
			},
			expected: true,
		},
		{
			name: "match tag",
			filter: &VersionFilter{
				Tags: []string{"stable"},
			},
			expected: true,
		},
		{
			name: "no match tag",
			filter: &VersionFilter{
				Tags: []string{"beta"},
			},
			expected: false,
		},
		{
			name: "match size range",
			filter: &VersionFilter{
				MinSize: &[]int64{400000}[0],
				MaxSize: &[]int64{600000}[0],
			},
			expected: true,
		},
		{
			name: "no match size range",
			filter: &VersionFilter{
				MinSize: &[]int64{600000}[0],
			},
			expected: false,
		},
		{
			name: "match pattern",
			filter: &VersionFilter{
				Pattern: "1.21.*",
			},
			expected: true,
		},
		{
			name: "no match pattern",
			filter: &VersionFilter{
				Pattern: "1.20.*",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := version.Matches(tt.filter)
			if matches != tt.expected {
				t.Errorf("Matches() = %v, 期望 %v", matches, tt.expected)
			}
		})
	}
}

func TestCompareVersionStrings(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.21.0", "1.20.0", 1},
		{"1.20.0", "1.21.0", -1},
		{"1.21.0", "1.21.0", 0},
		{"1.21.1", "1.21.0", 1},
		{"2.0.0", "1.21.0", 1},
		{"1.19.0", "1.21.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			comp := CompareVersionStrings(tt.v1, tt.v2)
			if comp.Result != tt.expected {
				t.Errorf("CompareVersionStrings(%s, %s) = %d, 期望 %d",
					tt.v1, tt.v2, comp.Result, tt.expected)
			}
		})
	}
}

func TestFilterVersions(t *testing.T) {
	versions := []*GoVersion{
		{
			Version:  "1.21.0",
			IsActive: true,
			Source:   SourceOnline,
			Tags:     []string{"stable"},
		},
		{
			Version:  "1.20.0",
			IsActive: false,
			Source:   SourceLocal,
			Tags:     []string{"beta"},
		},
		{
			Version:  "1.19.0",
			IsActive: false,
			Source:   SourceOnline,
			Tags:     []string{"stable", "lts"},
		},
	}

	// 测试按来源过滤
	filter1 := &VersionFilter{
		Source: &[]InstallSource{SourceOnline}[0],
	}
	filtered1 := FilterVersions(versions, filter1)
	if len(filtered1) != 2 {
		t.Errorf("按来源过滤: 期望2个版本，实际%d个", len(filtered1))
	}

	// 测试按标签过滤
	filter2 := &VersionFilter{
		Tags: []string{"stable"},
	}
	filtered2 := FilterVersions(versions, filter2)
	if len(filtered2) != 2 {
		t.Errorf("按标签过滤: 期望2个版本，实际%d个", len(filtered2))
	}

	// 测试按激活状态过滤
	filter3 := &VersionFilter{
		IsActive: &[]bool{true}[0],
	}
	filtered3 := FilterVersions(versions, filter3)
	if len(filtered3) != 1 {
		t.Errorf("按激活状态过滤: 期望1个版本，实际%d个", len(filtered3))
	}
}

func TestSortVersions(t *testing.T) {
	versions := []*GoVersion{
		{Version: "1.19.0", CreatedAt: time.Now().Add(-3 * time.Hour)},
		{Version: "1.21.0", CreatedAt: time.Now().Add(-1 * time.Hour)},
		{Version: "1.20.0", CreatedAt: time.Now().Add(-2 * time.Hour)},
	}

	// 测试按版本号升序排序
	sorter1 := &VersionSorter{
		Field:     "version",
		Direction: "asc",
	}
	sorted1 := SortVersions(versions, sorter1)
	if sorted1[0].Version != "1.19.0" || sorted1[2].Version != "1.21.0" {
		t.Error("版本号升序排序失败")
	}

	// 测试按版本号降序排序
	sorter2 := &VersionSorter{
		Field:     "version",
		Direction: "desc",
	}
	sorted2 := SortVersions(versions, sorter2)
	if sorted2[0].Version != "1.21.0" || sorted2[2].Version != "1.19.0" {
		t.Error("版本号降序排序失败")
	}

	// 测试按创建时间排序
	sorter3 := &VersionSorter{
		Field:     "created_at",
		Direction: "desc",
	}
	sorted3 := SortVersions(versions, sorter3)
	if sorted3[0].Version != "1.21.0" {
		t.Error("创建时间降序排序失败")
	}
}

func TestCalculateStatistics(t *testing.T) {
	now := time.Now()
	versions := []*GoVersion{
		{
			Version:         "1.21.0",
			IsActive:        true,
			Source:          SourceOnline,
			Tags:            []string{"stable", "lts"},
			CreatedAt:       now.Add(-1 * time.Hour),
			LastUsedAt:      &now,
			InstallDuration: 30 * time.Second,
			ExtractInfo: &ExtractInfo{
				ExtractedSize: 500000,
			},
		},
		{
			Version:         "1.20.0",
			IsActive:        false,
			Source:          SourceLocal,
			Tags:            []string{"beta"},
			CreatedAt:       now.Add(-2 * time.Hour),
			InstallDuration: 25 * time.Second,
			ExtractInfo: &ExtractInfo{
				ExtractedSize: 400000,
			},
		},
	}

	stats := CalculateStatistics(versions)

	if stats.TotalVersions != 2 {
		t.Errorf("TotalVersions = %d, 期望 2", stats.TotalVersions)
	}

	if stats.OnlineVersions != 1 {
		t.Errorf("OnlineVersions = %d, 期望 1", stats.OnlineVersions)
	}

	if stats.LocalVersions != 1 {
		t.Errorf("LocalVersions = %d, 期望 1", stats.LocalVersions)
	}

	if stats.ActiveVersion != "1.21.0" {
		t.Errorf("ActiveVersion = %s, 期望 1.21.0", stats.ActiveVersion)
	}

	if stats.TotalDiskUsage != 900000 {
		t.Errorf("TotalDiskUsage = %d, 期望 900000", stats.TotalDiskUsage)
	}

	if stats.TagStatistics["stable"] != 1 {
		t.Errorf("TagStatistics[stable] = %d, 期望 1", stats.TagStatistics["stable"])
	}

	if stats.AverageInstallTime != 27500*time.Millisecond {
		t.Errorf("AverageInstallTime = %v, 期望 27.5s", stats.AverageInstallTime)
	}
}

// timePtr 辅助函数：返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}
