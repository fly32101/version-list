package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// ToJSON 将GoVersion转换为JSON字符串
func (v *GoVersion) ToJSON() (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化GoVersion失败: %v", err)
	}
	return string(data), nil
}

// FromJSON 从JSON字符串创建GoVersion
func GoVersionFromJSON(jsonStr string) (*GoVersion, error) {
	var version GoVersion
	if err := json.Unmarshal([]byte(jsonStr), &version); err != nil {
		return nil, fmt.Errorf("反序列化GoVersion失败: %v", err)
	}
	return &version, nil
}

// ToMap 将GoVersion转换为map[string]interface{}
func (v *GoVersion) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"version":    v.Version,
		"path":       v.Path,
		"is_active":  v.IsActive,
		"source":     v.Source,
		"created_at": v.CreatedAt.Format(time.RFC3339),
		"updated_at": v.UpdatedAt.Format(time.RFC3339),
		"tags":       v.Tags,
		"notes":      v.Notes,
	}

	if v.LastUsedAt != nil {
		result["last_used_at"] = v.LastUsedAt.Format(time.RFC3339)
	}

	if v.InstallDuration > 0 {
		result["install_duration"] = v.InstallDuration.String()
	}

	if v.DownloadInfo != nil {
		result["download_info"] = map[string]interface{}{
			"url":           v.DownloadInfo.URL,
			"filename":      v.DownloadInfo.Filename,
			"size":          v.DownloadInfo.Size,
			"checksum":      v.DownloadInfo.Checksum,
			"checksum_type": v.DownloadInfo.ChecksumType,
			"downloaded_at": v.DownloadInfo.DownloadedAt.Format(time.RFC3339),
			"duration":      v.DownloadInfo.Duration,
			"speed":         v.DownloadInfo.Speed,
		}
	}

	if v.ExtractInfo != nil {
		result["extract_info"] = map[string]interface{}{
			"archive_size":   v.ExtractInfo.ArchiveSize,
			"extracted_size": v.ExtractInfo.ExtractedSize,
			"file_count":     v.ExtractInfo.FileCount,
			"duration":       v.ExtractInfo.Duration.String(),
			"root_dir":       v.ExtractInfo.RootDir,
		}
	}

	if v.ValidationInfo != nil {
		result["validation_info"] = map[string]interface{}{
			"checksum_valid":   v.ValidationInfo.ChecksumValid,
			"executable_valid": v.ValidationInfo.ExecutableValid,
			"version_valid":    v.ValidationInfo.VersionValid,
			"error":            v.ValidationInfo.Error,
		}
	}

	if v.SystemInfo != nil {
		result["system_info"] = map[string]interface{}{
			"os":       v.SystemInfo.OS,
			"arch":     v.SystemInfo.Arch,
			"version":  v.SystemInfo.Version,
			"filename": v.SystemInfo.Filename,
			"url":      v.SystemInfo.URL,
		}
	}

	return result
}

// Clone 创建GoVersion的深拷贝
func (v *GoVersion) Clone() *GoVersion {
	clone := &GoVersion{
		Version:         v.Version,
		Path:            v.Path,
		IsActive:        v.IsActive,
		Source:          v.Source,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
		InstallDuration: v.InstallDuration,
		Notes:           v.Notes,
	}

	// 复制LastUsedAt
	if v.LastUsedAt != nil {
		lastUsed := *v.LastUsedAt
		clone.LastUsedAt = &lastUsed
	}

	// 复制Tags
	if v.Tags != nil {
		clone.Tags = make([]string, len(v.Tags))
		copy(clone.Tags, v.Tags)
	}

	// 复制DownloadInfo
	if v.DownloadInfo != nil {
		clone.DownloadInfo = &DownloadInfo{
			URL:          v.DownloadInfo.URL,
			Filename:     v.DownloadInfo.Filename,
			Size:         v.DownloadInfo.Size,
			Checksum:     v.DownloadInfo.Checksum,
			ChecksumType: v.DownloadInfo.ChecksumType,
			DownloadedAt: v.DownloadInfo.DownloadedAt,
			Duration:     v.DownloadInfo.Duration,
			Speed:        v.DownloadInfo.Speed,
		}
	}

	// 复制ExtractInfo
	if v.ExtractInfo != nil {
		clone.ExtractInfo = &ExtractInfo{
			ArchiveSize:   v.ExtractInfo.ArchiveSize,
			ExtractedSize: v.ExtractInfo.ExtractedSize,
			FileCount:     v.ExtractInfo.FileCount,
			Duration:      v.ExtractInfo.Duration,
			RootDir:       v.ExtractInfo.RootDir,
		}
	}

	// 复制ValidationInfo
	if v.ValidationInfo != nil {
		clone.ValidationInfo = &ValidationInfo{
			ChecksumValid:   v.ValidationInfo.ChecksumValid,
			ExecutableValid: v.ValidationInfo.ExecutableValid,
			VersionValid:    v.ValidationInfo.VersionValid,
			Error:           v.ValidationInfo.Error,
		}
	}

	// 复制SystemInfo
	if v.SystemInfo != nil {
		clone.SystemInfo = &SystemInfo{
			OS:       v.SystemInfo.OS,
			Arch:     v.SystemInfo.Arch,
			Version:  v.SystemInfo.Version,
			Filename: v.SystemInfo.Filename,
			URL:      v.SystemInfo.URL,
		}
	}

	return clone
}

// Validate 验证GoVersion数据的完整性
func (v *GoVersion) Validate() error {
	if v.Version == "" {
		return fmt.Errorf("版本号不能为空")
	}

	if v.Path == "" {
		return fmt.Errorf("安装路径不能为空")
	}

	if v.Source != SourceLocal && v.Source != SourceOnline {
		return fmt.Errorf("无效的安装来源: %d", v.Source)
	}

	// 如果是在线安装，必须有下载信息
	if v.Source == SourceOnline && v.DownloadInfo == nil {
		return fmt.Errorf("在线安装版本必须包含下载信息")
	}

	// 验证下载信息
	if v.DownloadInfo != nil {
		if v.DownloadInfo.URL == "" {
			return fmt.Errorf("下载URL不能为空")
		}
		if v.DownloadInfo.Filename == "" {
			return fmt.Errorf("下载文件名不能为空")
		}
		if v.DownloadInfo.Size <= 0 {
			return fmt.Errorf("下载文件大小必须大于0")
		}
	}

	// 验证解压信息
	if v.ExtractInfo != nil {
		if v.ExtractInfo.FileCount <= 0 {
			return fmt.Errorf("解压文件数量必须大于0")
		}
		if v.ExtractInfo.ExtractedSize <= 0 {
			return fmt.Errorf("解压后大小必须大于0")
		}
	}

	return nil
}

// Summary 返回版本的摘要信息
func (v *GoVersion) Summary() string {
	status := "本地"
	if v.Source == SourceOnline {
		status = "在线"
	}

	active := ""
	if v.IsActive {
		active = " (当前)"
	}

	tags := ""
	if len(v.Tags) > 0 {
		tags = fmt.Sprintf(" [%s]", v.Tags[0])
	}

	return fmt.Sprintf("Go %s%s - %s安装%s", v.Version, active, status, tags)
}
