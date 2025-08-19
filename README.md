# Go版本管理工具

这是一个基于DDD架构的Go多版本管理工具，支持安装、切换和管理多个Go版本。

## 功能特性

- 安装多个Go版本
- 在不同Go版本之间切换
- 管理Go相关的环境变量
- 查看当前使用的Go版本
- 移除不需要的Go版本
- **彩色命令行输出**：增强用户体验，直观区分不同类型信息
- **Docker支持**：提供容器化部署方案

## 安装

### 本地安装

```bash
git clone <repository-url>
cd version-list
go build -o go-version.exe
```

将生成的`go-version.exe`（或`go-version`）添加到系统PATH中，以便在任何地方使用。

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

```bash
go-version install 1.21.0
```

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
