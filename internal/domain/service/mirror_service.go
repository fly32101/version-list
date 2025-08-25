package service

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// MirrorService 镜像管理服务接口
type MirrorService interface {
	GetAvailableMirrors() []Mirror
	TestMirrorSpeed(ctx context.Context, mirror Mirror) (*MirrorTestResult, error)
	SelectFastestMirror(ctx context.Context, mirrors []Mirror) (*Mirror, error)
	ValidateMirror(ctx context.Context, mirror Mirror) error
	GetMirrorByName(name string) (*Mirror, error)
	AddCustomMirror(mirror Mirror) error
	RemoveCustomMirror(name string) error
	SaveConfig(configPath string) error
}

// Mirror 镜像配置
type Mirror struct {
	Name        string `json:"name"`
	BaseURL     string `json:"base_url"`
	Description string `json:"description"`
	Region      string `json:"region"`
	Priority    int    `json:"priority"`
}

// MirrorTestResult 镜像测试结果
type MirrorTestResult struct {
	Mirror       Mirror        `json:"mirror"`
	ResponseTime time.Duration `json:"response_time"`
	Available    bool          `json:"available"`
	Error        error         `json:"error,omitempty"`
}

// mirrorServiceImpl 镜像服务实现
type mirrorServiceImpl struct {
	httpClient *http.Client
	mirrors    []Mirror
	config     *MirrorConfig
	configPath string
	mu         sync.RWMutex
}

// NewMirrorService 创建镜像服务实例
func NewMirrorService() MirrorService {
	config := NewMirrorConfig()

	return &mirrorServiceImpl{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		mirrors: getDefaultMirrors(),
		config:  config,
	}
}

// NewMirrorServiceWithConfig 使用指定配置创建镜像服务实例
func NewMirrorServiceWithConfig(configPath string) (MirrorService, error) {
	config := NewMirrorConfig()
	if err := config.LoadFromFile(configPath); err != nil {
		return nil, fmt.Errorf("加载镜像配置失败: %w", err)
	}

	return &mirrorServiceImpl{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		mirrors:    getDefaultMirrors(),
		config:     config,
		configPath: configPath,
	}, nil
}

// GetAvailableMirrors 获取可用镜像列表
func (s *mirrorServiceImpl) GetAvailableMirrors() []Mirror {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 合并默认镜像和自定义镜像
	allMirrors := make([]Mirror, len(s.mirrors))
	copy(allMirrors, s.mirrors)

	// 添加自定义镜像
	customMirrors := s.config.GetCustomMirrors()
	allMirrors = append(allMirrors, customMirrors...)

	// 按优先级排序
	sort.Slice(allMirrors, func(i, j int) bool {
		return allMirrors[i].Priority < allMirrors[j].Priority
	})

	return allMirrors
}

// TestMirrorSpeed 测试镜像速度
func (s *mirrorServiceImpl) TestMirrorSpeed(ctx context.Context, mirror Mirror) (*MirrorTestResult, error) {
	// 检查缓存
	if cached, exists := s.config.GetCachedResult(mirror.Name, 5*time.Minute); exists {
		result := &MirrorTestResult{
			Mirror:       mirror,
			Available:    cached.Available,
			ResponseTime: cached.ResponseTime,
		}
		if cached.Error != "" {
			result.Error = fmt.Errorf(cached.Error)
		}
		return result, nil
	}

	result := &MirrorTestResult{
		Mirror:    mirror,
		Available: false,
	}

	// 构建测试URL
	testURL := mirror.BaseURL
	if testURL[len(testURL)-1] != '/' {
		testURL += "/"
	}

	start := time.Now()

	// 先尝试HEAD请求，如果失败则尝试GET请求
	methods := []string{"HEAD", "GET"}
	var lastErr error

	for _, method := range methods {
		req, err := http.NewRequestWithContext(ctx, method, testURL, nil)
		if err != nil {
			lastErr = fmt.Errorf("创建%s请求失败: %w", method, err)
			continue
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("%s请求失败: %w", method, err)
			continue
		}
		defer resp.Body.Close()

		result.ResponseTime = time.Since(start)

		// 检查响应状态
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			result.Available = true
			// 缓存成功结果
			s.config.CacheTestResult(mirror.Name, result)
			return result, nil
		} else if resp.StatusCode == 405 && method == "HEAD" {
			// HEAD方法不支持，尝试GET
			continue
		} else {
			lastErr = fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
		}
	}

	result.Error = lastErr
	// 缓存失败结果
	s.config.CacheTestResult(mirror.Name, result)

	return result, nil
}

// SelectFastestMirror 选择最快的镜像
func (s *mirrorServiceImpl) SelectFastestMirror(ctx context.Context, mirrors []Mirror) (*Mirror, error) {
	if len(mirrors) == 0 {
		return nil, fmt.Errorf("没有可用的镜像")
	}

	// 并发测试所有镜像
	results := make(chan *MirrorTestResult, len(mirrors))
	var wg sync.WaitGroup

	for _, mirror := range mirrors {
		wg.Add(1)
		go func(m Mirror) {
			defer wg.Done()
			result, _ := s.TestMirrorSpeed(ctx, m)
			results <- result
		}(mirror)
	}

	// 等待所有测试完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集可用的镜像结果
	var availableResults []*MirrorTestResult
	for result := range results {
		if result.Available {
			availableResults = append(availableResults, result)
		}
	}

	if len(availableResults) == 0 {
		return nil, fmt.Errorf("没有可用的镜像")
	}

	// 按响应时间排序，选择最快的
	sort.Slice(availableResults, func(i, j int) bool {
		return availableResults[i].ResponseTime < availableResults[j].ResponseTime
	})

	return &availableResults[0].Mirror, nil
}

// ValidateMirror 验证镜像可用性
func (s *mirrorServiceImpl) ValidateMirror(ctx context.Context, mirror Mirror) error {
	result, err := s.TestMirrorSpeed(ctx, mirror)
	if err != nil {
		return err
	}

	if !result.Available {
		if result.Error != nil {
			return result.Error
		}
		return fmt.Errorf("镜像不可用")
	}

	return nil
}

// GetMirrorByName 根据名称获取镜像
func (s *mirrorServiceImpl) GetMirrorByName(name string) (*Mirror, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 先搜索默认镜像
	for _, mirror := range s.mirrors {
		if mirror.Name == name {
			return &mirror, nil
		}
	}

	// 再搜索自定义镜像
	customMirrors := s.config.GetCustomMirrors()
	for _, mirror := range customMirrors {
		if mirror.Name == name {
			return &mirror, nil
		}
	}

	return nil, fmt.Errorf("未找到名为 '%s' 的镜像", name)
}

// getDefaultMirrors 获取预设镜像配置
func getDefaultMirrors() []Mirror {
	return []Mirror{
		{
			Name:        "official",
			BaseURL:     "https://golang.org/dl/",
			Description: "Go官方下载源",
			Region:      "global",
			Priority:    1,
		},
		{
			Name:        "goproxy-cn",
			BaseURL:     "https://goproxy.cn/golang/",
			Description: "七牛云Go代理镜像",
			Region:      "china",
			Priority:    2,
		},
		{
			Name:        "aliyun",
			BaseURL:     "https://mirrors.aliyun.com/golang/",
			Description: "阿里云镜像源",
			Region:      "china",
			Priority:    3,
		},
		{
			Name:        "tencent",
			BaseURL:     "https://mirrors.cloud.tencent.com/golang/",
			Description: "腾讯云镜像源",
			Region:      "china",
			Priority:    4,
		},
		{
			Name:        "huawei",
			BaseURL:     "https://mirrors.huaweicloud.com/golang/",
			Description: "华为云镜像源",
			Region:      "china",
			Priority:    5,
		},
	}
}

// AddCustomMirror 添加自定义镜像
func (s *mirrorServiceImpl) AddCustomMirror(mirror Mirror) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查名称是否已存在（包括默认镜像和自定义镜像）
	for _, existing := range s.mirrors {
		if existing.Name == mirror.Name {
			return fmt.Errorf("镜像名称 '%s' 已存在（默认镜像）", mirror.Name)
		}
	}

	customMirrors := s.config.GetCustomMirrors()
	for _, existing := range customMirrors {
		if existing.Name == mirror.Name {
			return fmt.Errorf("镜像名称 '%s' 已存在（自定义镜像）", mirror.Name)
		}
	}

	// 添加到配置中
	s.config.AddCustomMirror(mirror)

	return nil
}

// RemoveCustomMirror 移除自定义镜像
func (s *mirrorServiceImpl) RemoveCustomMirror(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否为默认镜像（不能删除）
	for _, defaultMirror := range s.mirrors {
		if defaultMirror.Name == name {
			return fmt.Errorf("不能移除默认镜像源 '%s'", name)
		}
	}

	// 从配置中移除
	if !s.config.RemoveCustomMirror(name) {
		return fmt.Errorf("未找到自定义镜像源 '%s'", name)
	}

	return nil
}

// SaveConfig 保存配置到文件
func (s *mirrorServiceImpl) SaveConfig(configPath string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := configPath
	if path == "" {
		path = s.configPath
	}

	if path == "" {
		return fmt.Errorf("未指定配置文件路径")
	}

	return s.config.SaveToFile(path)
}
