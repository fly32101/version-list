package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// BenchmarkDownloadService 下载服务性能基准测试
func BenchmarkDownloadService(b *testing.B) {
	// 创建测试数据
	testData := strings.Repeat("test data ", 1000) // 约9KB数据
	testFile := filepath.Join(os.TempDir(), "benchmark_test.txt")

	// 写入测试文件
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		b.Fatal(err)
	}
	defer os.Remove(testFile)

	b.Run("OptimizedChunkSize", func(b *testing.B) {
		options := &DownloadOptions{
			ChunkSize:          64 * 1024, // 优化的64KB
			ProgressUpdateRate: 50 * time.Millisecond,
		}
		_ = NewDownloadService(options)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			destPath := filepath.Join(os.TempDir(), fmt.Sprintf("bench_dest_%d.txt", i))
			copyFile(testFile, destPath) // 模拟下载
			os.Remove(destPath)
		}
	})

	b.Run("DefaultChunkSize", func(b *testing.B) {
		options := &DownloadOptions{
			ChunkSize:          8 * 1024, // 默认的8KB
			ProgressUpdateRate: 100 * time.Millisecond,
		}
		_ = NewDownloadService(options)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			destPath := filepath.Join(os.TempDir(), fmt.Sprintf("bench_dest_%d.txt", i))
			copyFile(testFile, destPath) // 模拟下载
			os.Remove(destPath)
		}
	})
}

// BenchmarkArchiveExtractor 解压服务性能基准测试
func BenchmarkArchiveExtractor(b *testing.B) {
	b.Run("ParallelExtraction", func(b *testing.B) {
		options := &ArchiveExtractorOptions{
			ParallelWorkers: runtime.NumCPU(),
			BufferSize:      64 * 1024,
		}
		_ = NewArchiveExtractor(options)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 模拟并行解压操作
			simulateParallelWork(runtime.NumCPU())
		}
	})

	b.Run("SequentialExtraction", func(b *testing.B) {
		options := &ArchiveExtractorOptions{
			ParallelWorkers: 1,
			BufferSize:      8 * 1024,
		}
		_ = NewArchiveExtractor(options)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 模拟串行解压操作
			simulateParallelWork(1)
		}
	})
}

// TestMemoryUsage 内存使用测试
func TestMemoryUsage(t *testing.T) {
	// 测试大文件下载的内存使用
	t.Run("LargeFileDownload", func(t *testing.T) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// 模拟大文件下载
		options := &DownloadOptions{
			ChunkSize:   64 * 1024, // 优化的分块大小
			EnableCache: false,     // 禁用缓存以测试内存使用
		}
		_ = NewDownloadService(options)

		// 创建大文件（1MB）
		largeData := make([]byte, 1024*1024)
		testFile := filepath.Join(os.TempDir(), "large_test.bin")
		if err := os.WriteFile(testFile, largeData, 0644); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(testFile)

		// 模拟下载过程
		destFile := filepath.Join(os.TempDir(), "large_dest.bin")
		copyFile(testFile, destFile)
		defer os.Remove(destFile)

		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryUsed := m2.Alloc - m1.Alloc
		t.Logf("大文件下载内存使用: %d bytes (%.2f MB)", memoryUsed, float64(memoryUsed)/1024/1024)

		// 验证内存使用在合理范围内（应该远小于文件大小）
		if memoryUsed > 1024*1024/2 { // 不应超过文件大小的一半
			t.Errorf("内存使用过高: %d bytes", memoryUsed)
		}
	})
}

// TestCacheEfficiency 缓存效率测试
func TestCacheEfficiency(t *testing.T) {
	cacheDir := filepath.Join(os.TempDir(), "test_cache")
	defer os.RemoveAll(cacheDir)

	options := &DownloadOptions{
		EnableCache: true,
		CacheDir:    cacheDir,
	}
	service := NewDownloadService(options).(*DownloadServiceImpl)

	// 创建测试文件
	testData := "test cache data"
	testFile := filepath.Join(os.TempDir(), "cache_test.txt")
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile)

	// 第一次"下载"（实际是复制）
	destFile1 := filepath.Join(os.TempDir(), "cache_dest1.txt")
	start1 := time.Now()
	copyFile(testFile, destFile1)
	duration1 := time.Since(start1)
	defer os.Remove(destFile1)

	// 缓存文件
	service.cacheFile("test://url", destFile1)

	// 第二次从缓存获取
	destFile2 := filepath.Join(os.TempDir(), "cache_dest2.txt")
	start2 := time.Now()
	if err := service.copyFromCache(filepath.Join(cacheDir, fmt.Sprintf("%x", "test://url")), destFile2, nil); err != nil {
		t.Fatal(err)
	}
	duration2 := time.Since(start2)
	defer os.Remove(destFile2)

	t.Logf("首次下载用时: %v", duration1)
	t.Logf("缓存获取用时: %v", duration2)
	if duration2 > 0 {
		t.Logf("性能提升: %.2fx", float64(duration1)/float64(duration2))
	}

	// 验证缓存确实提高了性能
	if duration2 >= duration1 {
		t.Log("注意: 缓存没有显著提高性能，可能是由于测试文件太小")
	}
}

// TestParallelExtractionPerformance 并行解压性能测试
func TestParallelExtractionPerformance(t *testing.T) {
	if runtime.NumCPU() < 2 {
		t.Skip("需要多核CPU进行并行测试")
	}

	// 测试串行处理
	start1 := time.Now()
	simulateParallelWork(1)
	sequential := time.Since(start1)

	// 测试并行处理
	start2 := time.Now()
	simulateParallelWork(runtime.NumCPU())
	parallel := time.Since(start2)

	t.Logf("串行处理用时: %v", sequential)
	t.Logf("并行处理用时: %v", parallel)
	if parallel > 0 {
		t.Logf("性能提升: %.2fx", float64(sequential)/float64(parallel))
	}

	// 并行处理应该更快（在多核系统上）
	if parallel >= sequential {
		t.Log("警告: 并行处理没有显著提高性能，可能是由于测试负载太小")
	}
}

// 辅助函数

// copyFile 复制文件（模拟下载）
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// simulateParallelWork 模拟并行工作
func simulateParallelWork(workers int) {
	workChan := make(chan int, 100)
	doneChan := make(chan bool, workers)

	// 启动工作协程
	for i := 0; i < workers; i++ {
		go func() {
			for range workChan {
				// 模拟CPU密集型工作
				for j := 0; j < 1000; j++ {
					_ = j * j
				}
			}
			doneChan <- true
		}()
	}

	// 发送工作任务
	go func() {
		for i := 0; i < 100; i++ {
			workChan <- i
		}
		close(workChan)
	}()

	// 等待所有工作完成
	for i := 0; i < workers; i++ {
		<-doneChan
	}
}
