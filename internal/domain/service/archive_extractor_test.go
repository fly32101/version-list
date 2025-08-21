package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveExtractorImpl_GetArchiveType(t *testing.T) {
	extractor := NewArchiveExtractor(nil)

	testCases := []struct {
		filename     string
		expectedType ArchiveType
	}{
		{"test.zip", ArchiveTypeZip},
		{"test.ZIP", ArchiveTypeZip},
		{"test.tar.gz", ArchiveTypeTarGz},
		{"test.tgz", ArchiveTypeTarGz},
		{"test.TAR.GZ", ArchiveTypeTarGz},
		{"test.msi", ArchiveTypeMsi},
		{"test.MSI", ArchiveTypeMsi},
		{"test.txt", ArchiveTypeUnknown},
		{"test", ArchiveTypeUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			result := extractor.GetArchiveType(tc.filename)
			if result != tc.expectedType {
				t.Errorf("GetArchiveType(%s) = %v, 期望 %v", tc.filename, result, tc.expectedType)
			}
		})
	}
}

func TestArchiveExtractorImpl_ZipExtraction(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试ZIP文件
	zipPath := filepath.Join(tempDir, "test.zip")
	if err := createTestZip(zipPath); err != nil {
		t.Fatalf("创建测试ZIP文件失败: %v", err)
	}

	// 创建解压目录
	extractDir := filepath.Join(tempDir, "extracted")

	extractor := NewArchiveExtractor(&ArchiveExtractorOptions{
		PreservePermissions: true,
		OverwriteExisting:   true,
	})

	// 测试进度回调
	var progressCalls []ExtractionProgress
	progressCallback := func(progress ExtractionProgress) {
		progressCalls = append(progressCalls, progress)
	}

	// 执行解压
	err := extractor.Extract(zipPath, extractDir, progressCallback)
	if err != nil {
		t.Fatalf("解压ZIP文件失败: %v", err)
	}

	// 验证解压结果
	expectedFiles := []string{
		"test_dir/file1.txt",
		"test_dir/file2.txt",
		"root_file.txt",
	}

	for _, expectedFile := range expectedFiles {
		filePath := filepath.Join(extractDir, expectedFile)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("期望的文件不存在: %s", expectedFile)
		}
	}

	// 验证进度回调被调用
	if len(progressCalls) == 0 {
		t.Error("进度回调未被调用")
	}

	// 验证最后一次进度
	lastProgress := progressCalls[len(progressCalls)-1]
	if lastProgress.Percentage != 100 {
		t.Errorf("最终进度 = %.1f%%, 期望 100%%", lastProgress.Percentage)
	}
}

func TestArchiveExtractorImpl_TarGzExtraction(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试tar.gz文件
	tarGzPath := filepath.Join(tempDir, "test.tar.gz")
	if err := createTestTarGz(tarGzPath); err != nil {
		t.Fatalf("创建测试tar.gz文件失败: %v", err)
	}

	// 创建解压目录
	extractDir := filepath.Join(tempDir, "extracted")

	extractor := NewArchiveExtractor(nil)

	// 执行解压
	err := extractor.Extract(tarGzPath, extractDir, nil)
	if err != nil {
		t.Fatalf("解压tar.gz文件失败: %v", err)
	}

	// 验证解压结果
	expectedFiles := []string{
		"test_dir/file1.txt",
		"test_dir/file2.txt",
		"root_file.txt",
	}

	for _, expectedFile := range expectedFiles {
		filePath := filepath.Join(extractDir, expectedFile)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("期望的文件不存在: %s", expectedFile)
		}
	}
}

func TestArchiveExtractorImpl_ValidateArchive(t *testing.T) {
	tempDir := t.TempDir()

	extractor := NewArchiveExtractor(nil)

	// 测试有效的ZIP文件
	zipPath := filepath.Join(tempDir, "valid.zip")
	if err := createTestZip(zipPath); err != nil {
		t.Fatalf("创建测试ZIP文件失败: %v", err)
	}

	if err := extractor.ValidateArchive(zipPath); err != nil {
		t.Errorf("验证有效ZIP文件失败: %v", err)
	}

	// 测试有效的tar.gz文件
	tarGzPath := filepath.Join(tempDir, "valid.tar.gz")
	if err := createTestTarGz(tarGzPath); err != nil {
		t.Fatalf("创建测试tar.gz文件失败: %v", err)
	}

	if err := extractor.ValidateArchive(tarGzPath); err != nil {
		t.Errorf("验证有效tar.gz文件失败: %v", err)
	}

	// 测试无效文件
	invalidPath := filepath.Join(tempDir, "invalid.zip")
	if err := os.WriteFile(invalidPath, []byte("invalid content"), 0644); err != nil {
		t.Fatalf("创建无效文件失败: %v", err)
	}

	if err := extractor.ValidateArchive(invalidPath); err == nil {
		t.Error("期望验证无效文件时返回错误")
	}
}

func TestArchiveExtractorImpl_GetArchiveInfo(t *testing.T) {
	tempDir := t.TempDir()
	extractor := NewArchiveExtractor(nil)

	// 测试ZIP文件信息
	zipPath := filepath.Join(tempDir, "test.zip")
	if err := createTestZip(zipPath); err != nil {
		t.Fatalf("创建测试ZIP文件失败: %v", err)
	}

	zipInfo, err := extractor.GetArchiveInfo(zipPath)
	if err != nil {
		t.Fatalf("获取ZIP文件信息失败: %v", err)
	}

	if zipInfo.Type != ArchiveTypeZip {
		t.Errorf("ZIP文件类型 = %v, 期望 %v", zipInfo.Type, ArchiveTypeZip)
	}

	if zipInfo.FileCount <= 0 {
		t.Errorf("ZIP文件数量 = %d, 期望 > 0", zipInfo.FileCount)
	}

	// 测试tar.gz文件信息
	tarGzPath := filepath.Join(tempDir, "test.tar.gz")
	if err := createTestTarGz(tarGzPath); err != nil {
		t.Fatalf("创建测试tar.gz文件失败: %v", err)
	}

	tarGzInfo, err := extractor.GetArchiveInfo(tarGzPath)
	if err != nil {
		t.Fatalf("获取tar.gz文件信息失败: %v", err)
	}

	if tarGzInfo.Type != ArchiveTypeTarGz {
		t.Errorf("tar.gz文件类型 = %v, 期望 %v", tarGzInfo.Type, ArchiveTypeTarGz)
	}

	if tarGzInfo.FileCount <= 0 {
		t.Errorf("tar.gz文件数量 = %d, 期望 > 0", tarGzInfo.FileCount)
	}
}

func TestArchiveExtractorImpl_PathTraversalSecurity(t *testing.T) {
	tempDir := t.TempDir()

	// 创建包含路径遍历的恶意ZIP文件
	maliciousZipPath := filepath.Join(tempDir, "malicious.zip")
	if err := createMaliciousZip(maliciousZipPath); err != nil {
		t.Fatalf("创建恶意ZIP文件失败: %v", err)
	}

	extractDir := filepath.Join(tempDir, "extracted")
	extractor := NewArchiveExtractor(nil)

	// 尝试解压恶意文件，应该失败
	err := extractor.Extract(maliciousZipPath, extractDir, nil)
	if err == nil {
		t.Error("期望解压恶意文件时返回错误")
	}

	if !strings.Contains(err.Error(), "不安全的文件路径") {
		t.Errorf("期望路径遍历错误，得到: %v", err)
	}
}

// createTestZip 创建测试用的ZIP文件
func createTestZip(zipPath string) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 添加目录
	if _, err := zipWriter.Create("test_dir/"); err != nil {
		return err
	}

	// 添加文件
	files := map[string]string{
		"test_dir/file1.txt": "This is file 1 content",
		"test_dir/file2.txt": "This is file 2 content",
		"root_file.txt":      "This is root file content",
	}

	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			return err
		}

		if _, err := io.WriteString(writer, content); err != nil {
			return err
		}
	}

	return nil
}

// createTestTarGz 创建测试用的tar.gz文件
func createTestTarGz(tarGzPath string) error {
	file, err := os.Create(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// 添加目录
	dirHeader := &tar.Header{
		Name:     "test_dir/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}
	if err := tarWriter.WriteHeader(dirHeader); err != nil {
		return err
	}

	// 添加文件
	files := map[string]string{
		"test_dir/file1.txt": "This is file 1 content",
		"test_dir/file2.txt": "This is file 2 content",
		"root_file.txt":      "This is root file content",
	}

	for filename, content := range files {
		header := &tar.Header{
			Name:     filename,
			Mode:     0644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if _, err := tarWriter.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}

// createMaliciousZip 创建包含路径遍历的恶意ZIP文件
func createMaliciousZip(zipPath string) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 添加包含路径遍历的文件
	writer, err := zipWriter.Create("../../../malicious.txt")
	if err != nil {
		return err
	}

	if _, err := io.WriteString(writer, "This is malicious content"); err != nil {
		return err
	}

	return nil
}
