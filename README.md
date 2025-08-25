# GoVersion - Go版本管理工具

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey.svg)](README.md)

一个现代化的Go多版本管理工具，基于DDD架构设计，提供简单易用的命令行界面来安装、切换和管理多个Go版本。

## 📋 目录

- [✨ 亮点特性](#-亮点特性)
- [🚀 功能特性](#-功能特性)
- [📦 安装](#-安装)
- [⚡ 快速开始](#-快速开始)
- [📖 使用方法](#-使用方法)
- [🛠️ 故障排除](#️-故障排除)
- [⚡ 性能优化](#-性能优化)
- [🏗️ 架构设计](#️-架构设计)
- [👨‍💻 开发说明](#-开发说明)
- [🤝 贡献](#-贡献)
- [❓ 常见问题](#-常见问题)
- [📄 许可证](#-许可证)

## ✨ 亮点特性

- 🚀 **一键在线安装** - 自动从Go官网下载安装任意版本
- ⚡ **极速性能** - 优化算法，315,269文件/秒的移动速度
- 🎯 **智能检测** - 自动识别系统架构，无需手动配置
- 📊 **实时反馈** - 详细的进度显示和状态信息
- 🛡️ **可靠稳定** - 完整性验证、断点续传、自动重试
- 🔄 **无缝切换** - 符号链接技术，切换版本无需重启终端

## 🎬 演示

<!-- 这里可以添加 GIF 演示 -->
<!-- ![Demo](docs/demo.gif) -->

```bash
# 一键安装最新版本
$ go-version install 1.25.0
🚀 开始在线安装 Go 1.25.0...
📋 阶段: 检测系统 ⏱️ 20.0%
💬 检测到系统: windows/amd64
✅ 安装完成! 📊 总耗时: 2m15s

# 快速切换版本
$ go-version use 1.25.0
✅ 已切换到 Go 1.25.0

# 验证当前版本
$ go version
go version go1.25.0 windows/amd64
```

## 🚀 功能特性

### 核心功能

- 🚀 **在线安装**：从Go官网自动下载并安装任意版本
- 🔄 **版本切换**：在不同Go版本之间快速切换
- 📋 **版本管理**：查看、移除和管理已安装的Go版本
- 🌍 **环境管理**：自动管理Go相关的环境变量
- 📦 **本地导入**：导入系统中已安装的Go版本
- 🚀 **镜像源管理**：管理和优化Go下载镜像源，提高下载速度

### 增强特性

- 🎨 **彩色输出**：增强用户体验，直观区分不同类型信息
- 📊 **实时进度**：下载和安装过程的详细进度显示
- ⚡ **高性能**：优化的并行处理和智能缓存机制
- 🛡️ **可靠性**：自动重试、错误恢复和完整性验证
- 🐳 **Docker支持**：提供容器化部署方案
- 🔧 **灵活配置**：支持自定义安装路径和各种选项
- 🌐 **多镜像源支持**：内置5个高质量镜像源，支持自定义镜像源
- 🚀 **智能镜像选择**：自动测试并选择最快的镜像源

## 📦 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/your-username/go-version-manager.git
cd go-version-manager

# 构建可执行文件
go build -o go-version.exe

# 或者使用 make（如果有 Makefile）
make build
```

将生成的`go-version.exe`（或`go-version`）添加到系统PATH中，以便在任何地方使用。

### 预编译二进制文件

从 [Releases](https://github.com/your-username/go-version-manager/releases) 页面下载适合您系统的预编译二进制文件：

- **Windows**: `go-version-windows-amd64.exe`
- **Linux**: `go-version-linux-amd64`
- **macOS**: `go-version-darwin-amd64`

## ⚡ 快速开始

### 5分钟上手指南

1. **下载并安装工具**

   ```bash
   # 方式1: 从源码构建
   git clone https://github.com/your-username/go-version-manager.git
   cd go-version-manager
   go build -o go-version.exe
   
   # 方式2: 下载预编译二进制文件
   # 从 GitHub Releases 下载对应平台的文件
   ```

2. **设置环境变量**

   ```bash
   # Windows
   setup-env.bat
   
   # Linux/macOS  
   ./setup-env.sh
   ```

3. **安装Go版本**

   ```bash
   # 安装最新版本
   go-version install 1.25.0
   ```

4. **开始使用**

   ```bash
   # 验证安装
   go version
   
   # 查看已安装版本
   go-version list
   ```

### 常用命令速查

```bash
# 安装和管理
go-version install 1.25.0          # 在线安装
go-version install 1.24.0 --force  # 强制重新安装
go-version list                     # 查看所有版本
go-version use 1.25.0              # 切换版本
go-version current                  # 查看当前版本
go-version remove 1.24.0           # 移除版本

# 镜像源管理
go-version mirror list              # 查看所有镜像源
go-version mirror test              # 测试所有镜像源速度
go-version mirror fastest          # 选择最快镜像源
go-version install 1.21.0 --mirror goproxy-cn    # 使用指定镜像安装
go-version install 1.21.0 --auto-mirror          # 自动选择最快镜像

# 高级选项
go-version install 1.25.0 --path "C:\Go1.25.0"  # 自定义路径
go-version install 1.25.0 --timeout 600         # 设置超时
go-version import "C:\Go"                        # 导入现有安装
```

### Docker安装

```bash
# 构建镜像
docker build -t go-version .

# 运行容器
docker run -it --rm -v /usr/local/go-versions:/go-versions go-version list
```

注意：使用Docker时，需要将Go版本安装目录挂载到容器的`/go-versions`路径。

## 📖 使用方法

### 初始设置

首次使用前，需要设置环境变量：

**Linux/macOS:**

```bash
chmod +x setup-env.sh
./setup-env.sh
```

**Windows:**
以管理员身份运行：

```cmd
setup-env.bat
```

#### 环境变量说明

本工具依赖以下环境变量，它们会在运行安装脚本时自动设置：

1. **GO_VERSIONS_PATH**
   - 描述：指定Go版本的安装目录
   - 默认值（Windows）：`%USERPROFILE%\.go-versions`
   - 默认值（Linux/macOS）：`$HOME/.go-versions`
   - 示例：`C:\Users\username\.go-versions`

2. **GO_VERSIONS_CURRENT**
   - 描述：记录当前使用的Go版本
   - 位置：存储在GO_VERSIONS_PATH目录下的`.current-version`文件中
   - 示例：`1.21.0`

3. **GOROOT**
   - 描述：Go的安装根目录
   - 设置方式：自动指向当前使用的Go版本的安装路径
   - 示例：`C:\Users\username\.go-versions\1.21.0`

4. **PATH**
   - 描述：系统可执行文件搜索路径
   - 设置方式：自动添加当前Go版本的bin目录
   - 示例：`%GOROOT%\bin` 或 `$GOROOT/bin`

#### 手动配置环境变量

如果自动安装脚本无法使用，您可以手动配置这些环境变量：

**Windows:**

```cmd
# 设置Go版本安装目录
setx GO_VERSIONS_PATH "%USERPROFILE%\.go-versions"

# 创建目录
mkdir "%USERPROFILE%\.go-versions"

# 添加到PATH（需要重启终端生效）
setx PATH "%PATH%;%USERPROFILE%\.go-versions\current\bin"
```

**Linux/macOS:**

```bash
# 添加到~/.bashrc或~/.zshrc
echo 'export GO_VERSIONS_PATH="$HOME/.go-versions"' >> ~/.bashrc
echo 'export PATH="$PATH:$GO_VERSIONS_PATH/current/bin"' >> ~/.bashrc

# 创建目录
mkdir -p "$HOME/.go-versions"

# 应用更改
source ~/.bashrc
```

### 列出所有已安装的Go版本

```bash
go-version list
```

输出示例（带颜色）：

```text
版本      路径                              状态      安装来源    最后使用
1.25.0    ~/.go/versions/1.25.0            当前使用   在线安装    2024-01-15
1.24.0    ~/.go/versions/1.24.0                      在线安装    2024-01-10  
1.21.0    /usr/local/go                              本地导入    2023-12-20
```

### 安装指定版本的Go

#### 在线安装（推荐）

从Go官方网站自动下载并安装指定版本：

```bash
# 基本用法：安装最新版本
go-version install 1.25.0

# 安装到自定义路径
go-version install 1.25.0 --path "D:\tools\go\go1.25.0"

# 强制重新安装（覆盖已存在的版本）
go-version install 1.25.0 --force

# 跳过文件完整性验证（加快安装速度）
go-version install 1.25.0 --skip-verification

# 设置下载超时时间（秒）
go-version install 1.25.0 --timeout 600

# 设置最大重试次数
go-version install 1.25.0 --max-retries 5
```

**在线安装特性：**

- 🚀 **自动系统检测**：自动识别操作系统和CPU架构
- 📥 **智能下载**：从Go官网下载对应版本的安装包
- 📊 **实时进度**：显示下载和解压进度，包括速度和预计剩余时间
- ✅ **完整性验证**：自动验证下载文件的完整性
- 🔄 **断点续传**：支持网络中断后的断点续传
- ⚡ **高性能解压**：优化的并行解压算法，支持大文件快速处理
- 🛡️ **错误恢复**：自动重试和回滚机制，确保安装可靠性
- ⏰ **超时保护**：防止安装过程无限期挂起

**安装过程示例：**

```text
🚀 开始在线安装 Go 1.25.0...

📋 阶段: 检测系统
⏱️  进度: 20.0%
💬 检测到系统: windows/amd64
💬 下载URL: https://go.dev/dl/go1.25.0.windows-amd64.zip

📋 阶段: 下载文件  
⏱️  进度: 45.0%
💬 下载中... 67.3% (15.2 MB/s) 预计剩余: 1m30s
💬 已下载: 98.5 MB / 146.2 MB

📋 阶段: 解压文件
⏱️  进度: 85.0%
💬 解压中... 100.0% (14,496/14,496 文件)
💬 已移动 14,496 个文件到目标目录

📋 阶段: 配置安装
⏱️  进度: 95.0%
💬 验证 Go 二进制文件...
💬 更新版本记录...

✅ 安装完成! 
📊 总耗时: 2m15s
📁 安装路径: ~/.go/versions/1.25.0
🎯 使用 'go-version use 1.25.0' 切换到此版本
```

#### 离线安装

如果您已经下载了Go安装包，可以使用离线安装：

```bash
# 从本地压缩包安装
go-version install --local "path/to/go1.25.0.windows-amd64.zip"

# 从本地目录安装
go-version install --local "path/to/extracted/go/"
```

#### 命令选项详解

| 选项 | 简写 | 描述 | 默认值 | 示例 |
|------|------|------|--------|------|
| `--path` | `-p` | 自定义安装路径 | `~/.go/versions/{version}` | `--path "D:\Go\1.25.0"` |
| `--force` | `-f` | 强制重新安装已存在的版本 | `false` | `--force` |
| `--skip-verification` | `-s` | 跳过文件完整性验证 | `false` | `--skip-verification` |
| `--timeout` | `-t` | 下载超时时间（秒） | `300` | `--timeout 600` |
| `--max-retries` | `-r` | 最大重试次数 | `3` | `--max-retries 5` |
| `--local` | `-l` | 从本地文件安装 | - | `--local "go1.25.0.zip"` |
| `--help` | `-h` | 显示帮助信息 | - | `--help` |

#### 支持的Go版本

工具支持安装Go 1.13及以上的所有版本，包括：

- **稳定版本**：1.21.0, 1.22.0, 1.23.0, 1.24.0, 1.25.0等
- **测试版本**：1.25rc1, 1.25beta1等（如果可用）
- **历史版本**：1.13.x, 1.14.x, 1.15.x等

#### 支持的平台

| 操作系统 | 架构 | 文件格式 | 状态 |
|----------|------|----------|------|
| Windows | amd64 | .zip | ✅ 完全支持 |
| Windows | 386 | .zip | ✅ 完全支持 |
| Windows | arm64 | .zip | ✅ 完全支持 |
| Linux | amd64 | .tar.gz | ✅ 完全支持 |
| Linux | 386 | .tar.gz | ✅ 完全支持 |
| Linux | arm64 | .tar.gz | ✅ 完全支持 |
| macOS | amd64 | .tar.gz | ✅ 完全支持 |
| macOS | arm64 | .tar.gz | ✅ 完全支持 |

### 切换到指定版本的Go

```bash
go-version use 1.21.0
```

**注意：** 现在使用符号链接方式，切换版本后无需重启终端！

### 查看当前使用的Go版本

```bash
go-version current
```

### 移除指定版本的Go

```bash
go-version remove 1.20.0
```

### 导入本地已安装的Go版本

```bash
go-version import "C:\Go"
```

此命令用于导入系统中已经安装的Go版本，特别适用于以下场景：

- 系统中已预装Go，但希望通过go-version工具进行管理
- 已手动安装了某个Go版本，希望将其纳入go-version的管理范围

注意：导入路径必须是Go的安装根目录，包含bin、src等子目录。

### 镜像源管理

`mirror`命令提供了完整的镜像源管理功能，帮助您优化Go版本下载速度。

#### 查看所有可用的镜像源

```bash
# 基本列表
go-version mirror list

# 显示详细信息
go-version mirror list --details
```

输出示例：

```text
可用的镜像源:

  official     Go官方下载源 (global)
               https://golang.org/dl/

  goproxy-cn   七牛云Go代理镜像 (china)
               https://goproxy.cn/golang/

  aliyun       阿里云镜像源 (china)
               https://mirrors.aliyun.com/golang/

  tencent      腾讯云镜像源 (china)
               https://mirrors.cloud.tencent.com/golang/

  huawei       华为云镜像源 (china)
               https://mirrors.huaweicloud.com/golang/
```

#### 测试镜像源速度和可用性

```bash
# 测试所有镜像源
go-version mirror test

# 测试指定镜像源
go-version mirror test --name goproxy-cn

# 设置超时时间（60秒）
go-version mirror test --timeout 60

# 强制测试（忽略缓存）
go-version mirror test --force
```

测试输出示例：

```text
正在测试镜像源...

测试 official (Go官方下载源)...
  ✅ 可用 (响应时间: 245ms)

测试 goproxy-cn (七牛云Go代理镜像)...
  ✅ 可用 (响应时间: 156ms)

测试 aliyun (阿里云镜像源)...
  ✅ 可用 (响应时间: 189ms)

测试结果总结:
  1. goproxy-cn - 156ms
  2. aliyun - 189ms
  3. official - 245ms
```

#### 自动选择最快的镜像源

```bash
# 基本用法
go-version mirror fastest

# 显示详细测试过程
go-version mirror fastest --details

# 设置超时时间
go-version mirror fastest --timeout 60
```

输出示例：

```text
正在测试所有镜像源以选择最快的...

✅ 最快的镜像源: goproxy-cn
描述: 七牛云Go代理镜像
地区: china
URL: https://goproxy.cn/golang/

使用此镜像的示例:
  go-version install 1.21.0 --mirror goproxy-cn
```

#### 验证镜像源可用性

```bash
# 验证指定镜像源
go-version mirror validate --name official
go-version mirror validate --name goproxy-cn
```

#### 添加自定义镜像源

```bash
# 添加公司内部镜像
go-version mirror add \
  --name mycompany \
  --url "https://mirrors.mycompany.com/golang/" \
  --description "公司内部镜像" \
  --region "内网" \
  --priority 1

# 添加其他自定义镜像
go-version mirror add \
  --name university \
  --url "https://mirrors.university.edu/golang/" \
  --description "大学镜像源" \
  --region "教育网" \
  --priority 10
```

**参数说明：**

| 参数 | 必需 | 描述 | 示例 |
|------|------|------|------|
| `--name` | 是 | 镜像源名称（唯一标识） | `mycompany` |
| `--url` | 是 | 镜像源URL | `https://mirrors.example.com/golang/` |
| `--description` | 是 | 镜像源描述 | `"公司内部镜像"` |
| `--region` | 是 | 镜像源地区 | `"内网"` |
| `--priority` | 否 | 镜像源优先级（数字越小优先级越高） | `1` |

#### 移除自定义镜像源

```bash
# 移除指定的自定义镜像源
go-version mirror remove --name mycompany
go-version mirror remove --name university
```

**注意：**
- 只能移除自定义添加的镜像源
- 不能移除内置的默认镜像源（official、goproxy-cn、aliyun、tencent、huawei）

#### 在安装时使用镜像源

```bash
# 使用指定镜像源安装
go-version install 1.21.0 --mirror goproxy-cn
go-version install 1.21.0 --mirror aliyun
go-version install 1.21.0 --mirror mycompany

# 自动选择最快镜像源安装
go-version install 1.21.0 --auto-mirror

# 组合使用其他选项
go-version install 1.21.0 --mirror goproxy-cn --force --timeout 600
```

#### 镜像源配置管理

自定义镜像源的配置保存在 `mirrors.json` 文件中：

```bash
# 查看配置文件位置
go-version mirror list --config "path/to/mirrors.json"

# 使用自定义配置文件
go-version mirror add --config "path/to/mirrors.json" \
  --name custom \
  --url "https://example.com/golang/" \
  --description "Custom mirror" \
  --region "custom"
```

#### 镜像源特性

- **智能缓存**：测试结果会被缓存，避免重复测试
- **并发测试**：同时测试多个镜像源，提高效率
- **错误重试**：自动重试失败的测试，提高可靠性
- **多种测试方法**：支持HEAD和GET请求，兼容性更好
- **灵活配置**：支持自定义配置文件路径

注意：导入路径必须是Go的安装根目录，包含bin、src等子目录。

## 🛠️ 故障排除

### 在线安装问题

#### 网络连接问题

```bash
# 如果遇到网络超时，可以增加超时时间和重试次数
go-version install 1.25.0 --timeout 900 --max-retries 5
```

#### 解压过程卡死

如果安装过程在解压阶段卡死，这通常是由于：

- 磁盘空间不足
- 防病毒软件干扰
- 文件权限问题

**解决方案：**

1. 确保有足够的磁盘空间（至少500MB）
2. 临时禁用防病毒软件的实时保护
3. 以管理员权限运行命令
4. 使用 `--force` 选项重新安装

#### 下载速度慢

```bash
# 可以跳过验证以加快安装速度（不推荐用于生产环境）
go-version install 1.25.0 --skip-verification
```

#### 自定义安装位置

```bash
# 安装到指定目录
go-version install 1.25.0 --path "C:\CustomPath\Go1.25.0"
```

### 环境变量问题

如果切换版本后Go命令不可用，请检查：

1. **PATH环境变量**：确保包含 `%GO_VERSIONS_PATH%\current\bin`
2. **GOROOT环境变量**：应该指向当前版本的安装目录
3. **重启终端**：某些情况下需要重启终端使环境变量生效

### 权限问题

**Windows:**

```cmd
# 以管理员身份运行PowerShell或命令提示符
runas /user:Administrator cmd
```

**Linux/macOS:**

```bash
# 使用sudo运行（如果需要）
sudo go-version install 1.25.0
```

### 清理和重置

如果遇到严重问题，可以完全重置：

```bash
# 移除所有已安装的版本
go-version remove --all

# 清理环境变量（需要手动操作）
# Windows: 在系统设置中删除相关环境变量
# Linux/macOS: 编辑 ~/.bashrc 或 ~/.zshrc 删除相关配置
```

## ⚡ 性能优化

### 安装性能

本工具经过深度优化，提供卓越的安装性能：

- **智能目录移动**：优先使用系统级目录重命名，速度可达 300,000+ 文件/秒
- **并发处理**：多线程并行文件操作，充分利用多核CPU
- **优化缓冲区**：使用大缓冲区（1MB）提高I/O性能
- **断点续传**：网络中断后自动恢复下载
- **超时保护**：防止安装过程无限期挂起（默认10分钟超时）

### 性能基准

在标准测试环境下的性能表现：

| 操作类型 | 文件数量 | 耗时 | 速度 |
|---------|---------|------|------|
| 目录移动 | 516个文件 | 1.6ms | 315,269 文件/秒 |
| 并发解压 | 14,496个文件 | <30秒 | 483+ 文件/秒 |
| 下载 | 150MB安装包 | 取决于网速 | 支持断点续传 |

### 内存使用

- **下载阶段**：流式处理，内存使用稳定在64MB以下
- **解压阶段**：并行处理，内存使用根据并发数动态调整
- **文件移动**：批量处理，内存使用优化到最低

## 🏗️ 架构设计

本项目采用DDD（领域驱动设计）架构，分为以下几层：

- **Domain层**：包含核心业务逻辑和模型
  - `model`：包含领域模型，如Go版本、环境变量等
  - `repository`：定义仓库接口，抽象数据访问
  - `service`：包含领域服务，实现核心业务逻辑

- **Infrastructure层**：包含技术实现细节
  - `persistence`：实现仓库接口，处理数据持久化

- **Application层**：应用服务，协调领域对象
  - `version_app_service`：提供应用级别的服务接口

- **Interface层**：用户接口
  - `cli`：实现命令行界面

## 👨‍💻 开发说明

### 项目结构

```text
go-version-manager/
├── 📁 cmd/                     # 命令行入口
├── 📁 internal/                # 内部包
│   ├── 📁 domain/              # 领域层
│   │   ├── 📁 model/           # 领域模型
│   │   ├── 📁 repository/      # 仓库接口
│   │   └── 📁 service/         # 领域服务
│   ├── 📁 infrastructure/      # 基础设施层
│   │   └── 📁 persistence/     # 数据持久化
│   ├── 📁 application/         # 应用层
│   └── 📁 interface/           # 接口层
│       ├── 📁 cli/             # 命令行界面
│       └── 📁 ui/              # 用户界面组件
├── 📁 examples/                # 示例代码
├── 📁 docs/                    # 文档
├── 📁 scripts/                 # 构建脚本
├── 📄 go.mod                   # Go模块文件
├── 📄 main.go                  # 主入口
├── 📄 Makefile                 # 构建配置
├── 📄 Dockerfile               # Docker配置
└── 📄 README.md                # 项目说明
```

### 技术栈

- **语言**: Go 1.21+
- **架构**: DDD (领域驱动设计)
- **CLI框架**: Cobra
- **测试**: Go标准测试库 + Testify
- **构建**: Go Modules
- **部署**: 单二进制文件 / Docker

### 扩展功能

如果需要扩展功能，可以按照以下步骤进行：

1. 在`domain/model`中定义或修改领域模型
2. 在`domain/repository`中定义或修改仓库接口
3. 在`domain/service`中实现或修改领域服务
4. 在`infrastructure/persistence`中实现仓库接口
5. 在`application`中修改或添加应用服务
6. 在`interface/cli`中添加或修改命令行接口

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

### 开发环境设置

```bash
# 克隆仓库
git clone https://github.com/your-username/go-version-manager.git
cd go-version-manager

# 安装依赖
go mod download

# 运行测试
go test ./...

# 构建
go build -o go-version.exe
```

## 📝 更新日志

### v2.0.0 (2024-01-15)
- ✨ 新增在线安装功能
- ⚡ 性能优化：315,269文件/秒的移动速度
- 🛡️ 增强错误处理和恢复机制
- 📊 改进进度显示和用户体验
- 🔧 支持更多命令行选项

### v1.0.0 (2023-12-01)
- 🎉 首次发布
- 📦 基本的版本管理功能
- 🔄 版本切换和环境变量管理

## ❓ 常见问题

### Q: 支持哪些Go版本？
A: 支持Go 1.13及以上的所有版本，包括稳定版、RC版和Beta版。

### Q: 可以同时安装多个版本吗？
A: 是的，可以安装任意数量的Go版本，并在它们之间快速切换。

### Q: 安装的Go版本存储在哪里？
A: 默认存储在 `~/.go/versions/` 目录下，可以通过 `--path` 选项自定义。

### Q: 如何卸载工具？
A: 删除可执行文件和 `~/.go/versions/` 目录，然后清理环境变量即可。

## 📞 支持

- 🐛 [报告Bug](https://github.com/your-username/go-version-manager/issues)
- 💡 [功能请求](https://github.com/your-username/go-version-manager/issues)
- 📖 [文档](https://github.com/your-username/go-version-manager/wiki)
- 💬 [讨论](https://github.com/your-username/go-version-manager/discussions)

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE) - 查看 LICENSE 文件了解详情。

---

<div align="center">
  <p>如果这个项目对您有帮助，请给个 ⭐️ Star 支持一下！</p>
  <p>Made with ❤️ by Go developers, for Go developers</p>
</div>
