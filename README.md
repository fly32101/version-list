# Go版本管理工具

这是一个基于DDD架构的Go多版本管理工具，支持安装、切换和管理多个Go版本。

## 功能特性

### 核心功能

- 🚀 **在线安装**：从Go官网自动下载并安装任意版本
- 🔄 **版本切换**：在不同Go版本之间快速切换
- 📋 **版本管理**：查看、移除和管理已安装的Go版本
- 🌍 **环境管理**：自动管理Go相关的环境变量
- 📦 **本地导入**：导入系统中已安装的Go版本

### 增强特性

- 🎨 **彩色输出**：增强用户体验，直观区分不同类型信息
- 📊 **实时进度**：下载和安装过程的详细进度显示
- ⚡ **高性能**：优化的并行处理和智能缓存机制
- 🛡️ **可靠性**：自动重试、错误恢复和完整性验证
- 🐳 **Docker支持**：提供容器化部署方案
- 🔧 **灵活配置**：支持自定义安装路径和各种选项

## 安装

### 本地安装

```bash
git clone <repository-url>
cd version-list
go build -o go-version.exe
```

将生成的`go-version.exe`（或`go-version`）添加到系统PATH中，以便在任何地方使用。

## 快速开始

### 5分钟上手指南

1. **下载并安装工具**

   ```bash
   # 克隆仓库
   git clone <repository-url>
   cd version-list
   go build -o go-version.exe
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

## 使用方法

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

```
版本      路径                    状态      
1.21.0    /usr/local/go-versions/1.21.0   
1.20.5    /usr/local/go-versions/1.20.5   当前使用
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
开始在线安装Go 1.25.0...
📋 阶段: 检测系统
⏱️  进度: 20.0%
💬 检测到系统: windows/amd64

📋 阶段: 下载文件  
⏱️  进度: 45.0%
💬 下载中... 67.3% (15.2 MB/s)

📋 阶段: 解压文件
⏱️  进度: 85.0%
💬 解压中... 100.0% (14496/14496 文件)

📋 阶段: 配置安装
⏱️  进度: 95.0%
💬 配置Go环境...

✅ 安装完成! 耗时: 2m15s
```

#### 本地安装

如果您已经下载了Go安装包，可以使用本地安装：

```bash
go-version install --local "path/to/go1.25.0.windows-amd64.zip"
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

## 故障排除

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

## 性能优化

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

## 架构设计

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

## 开发说明

### 项目结构

```
version-list/
├── internal/
│   ├── domain/
│   │   ├── model/          # 领域模型
│   │   ├── repository/     # 仓库接口
│   │   └── service/        # 领域服务
│   ├── infrastructure/
│   │   └── persistence/    # 数据持久化实现
│   ├── application/
│   │   └── version_app_service.go  # 应用服务
│   └── interface/
│       └── cli/            # 命令行界面
├── go.mod
├── main.go
└── README.md
```

### 扩展功能

如果需要扩展功能，可以按照以下步骤进行：

1. 在`domain/model`中定义或修改领域模型
2. 在`domain/repository`中定义或修改仓库接口
3. 在`domain/service`中实现或修改领域服务
4. 在`infrastructure/persistence`中实现仓库接口
5. 在`application`中修改或添加应用服务
6. 在`interface/cli`中添加或修改命令行接口

## 许可证

MIT
