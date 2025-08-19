package model

import (
	"time"
)

// GoVersion 表示一个Go版本
type GoVersion struct {
	Version   string    // 版本号，如 "1.21.0"
	Path      string    // 安装路径
	IsActive  bool      // 是否为当前激活版本
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// Environment 表示环境变量配置
type Environment struct {
	GOROOT string // Go安装目录
	GOPATH string // 工作目录
	GOBIN  string // 二进制文件目录
}
