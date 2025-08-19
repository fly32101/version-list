# Go版本管理工具

这是一个基于DDD架构的Go多版本管理工具，支持安装、切换和管理多个Go版本。

## 功能特性

- 安装多个Go版本
- 在不同Go版本之间切换
- 管理Go相关的环境变量
- 查看当前使用的Go版本
- 移除不需要的Go版本

## 安装

```bash
git clone <repository-url>
cd version-list
go build -o go-version.exe
```

将生成的`go-version.exe`（或`go-version`）添加到系统PATH中，以便在任何地方使用。

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

### 列出所有已安装的Go版本

```bash
go-version list
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
