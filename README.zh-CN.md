# Gitta - Git 任务助手

轻量、原生 Git 的任务管理工具，把仓库作为唯一事实来源。任务以 Markdown 存储，分支状态自动反映进度，无需额外服务。

[![CI](https://github.com/GavinWu1991/gitta/actions/workflows/ci.yml/badge.svg)](https://github.com/GavinWu1991/gitta/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-brightgreen.svg)](go.mod)

---

## 目录

- [Gitta - Git 任务助手](#gitta---git-任务助手)
  - [目录](#目录)
  - [Gitta 是什么？](#gitta-是什么)
    - [功能](#功能)
  - [快速开始](#快速开始)
    - [先决条件](#先决条件)
    - [下载预构建二进制](#下载预构建二进制)
    - [安装](#安装)
    - [一行安装 + 初始化（自动下载 + init 脚本）](#一行安装--初始化自动下载--init-脚本)
    - [构建](#构建)
    - [首次命令](#首次命令)
  - [可用命令](#可用命令)
    - [快速示例](#快速示例)
  - [常见工作流](#常见工作流)
    - [入门流程（安装 → 列表 → 开始 → 验证）](#入门流程安装--列表--开始--验证)
    - [日常流程（更新 → 列表 → 开始/继续 → 提交）](#日常流程更新--列表--开始继续--提交)
    - [Sprint 规划（Sprint 与 backlog）](#sprint-规划sprint-与-backlog)
  - [架构](#架构)
  - [开发](#开发)
    - [项目结构](#项目结构)
    - [测试](#测试)
    - [新增命令](#新增命令)
  - [贡献](#贡献)
  - [文档](#文档)
  - [支持](#支持)
  - [许可证](#许可证)

---

## Gitta 是什么？

Gitta 是一款 Git 任务助手，把任务存成带 YAML Frontmatter 的 Markdown，并用 Git 分支推导状态。无需服务器或外部服务，只要有 Git 就能工作。

### 功能

- **零基础设施**：无须部署任何服务，开箱即用。
- **Git 原生**：任务保存在仓库的 Markdown 文件中。
- **分支感知**：分支状态自动驱动任务状态。
- **命令行优先**：快速 CLI 流程，未来支持 TUI。
- **离线优先**：本地即可使用，适合低联网场景。

---

## 快速开始

### 先决条件

- Go 1.21 或更高
- Git
- Make（可选，用于开发）

### 下载预构建二进制

> 推荐：最快 2 分钟即可运行，无需 Go 环境。

1. 访问 GitHub Releases：选择需要的版本  
2. 下载适合平台的压缩包：  
   - macOS：`gitta-<version>-darwin-amd64.tar.gz`（Intel）或 `darwin-arm64.tar.gz`（Apple Silicon）  
   - Linux：`gitta-<version>-linux-amd64.tar.gz` 或 `linux-arm64.tar.gz`  
   - Windows：`gitta-<version>-windows-amd64.zip` 或 `windows-arm64.zip`
3. 校验完整性（推荐）：  
   ```bash
   shasum -a 256 gitta-<version>-<platform>-<arch>.tar.gz
   # 或使用 checksums.txt 中的值比对
   ```
4. 解压并添加到 PATH：  
   ```bash
   tar -xzf gitta-<version>-darwin-amd64.tar.gz   # macOS/Linux
   unzip gitta-<version>-windows-amd64.zip        # Windows
   sudo mv gitta /usr/local/bin/                  # 可选
   gitta --help
   ```

### 安装

```bash
# 克隆仓库
git clone https://github.com/GavinWu1991/gitta.git
cd gitta

# 安装依赖
go mod tidy

# 验证安装
make verify  # 运行全部检查
```

### 一行安装 + 初始化（自动下载 + init 脚本）

```bash
curl -sSf https://raw.githubusercontent.com/GavinWu1991/gitta/main/scripts/remote-init.sh | bash
# 强制重建或自定义 Sprint 名：
curl -sSf https://raw.githubusercontent.com/GavinWu1991/gitta/main/scripts/remote-init.sh | bash -s -- --force --example-sprint Sprint-02
```

### 构建

```bash
# 构建二进制
go build -o gitta ./cmd/gitta

# 验证可用
./gitta --help
./gitta version
```

### 首次命令

```bash
# 查看当前 Sprint 任务
gitta list

# 同时查看 Sprint + backlog
gitta list --all

# 开始一个任务
gitta start US-001

# 查看版本
gitta version
```

---

## 可用命令

| 命令 | 描述 | 基本用法 | 文档 |
|------|------|----------|------|
| `gitta list` | 显示当前 Sprint 任务；`--all` 包含 backlog | `gitta list [--all]` | [docs/cli/list.md](docs/cli/list.md) |
| `gitta start` | 为任务创建/切换分支，可选更新 assignee | `gitta start <task-id|file-path> [--assignee <name>]` | [docs/cli/start.md](docs/cli/start.md) |
| `gitta version` | 输出版本、提交、构建时间、Go 版本 | `gitta version [--json]` | [docs/cli/version.md](docs/cli/version.md) |

### 快速示例

```bash
# 仅 Sprint
gitta list

# Sprint + backlog
gitta list --all

# 通过任务 ID 开始
gitta start US-001

# 通过文件路径开始
gitta start sprints/Sprint-01/US-001.md

# JSON 版本信息
gitta version --json
```

---

## 常见工作流

### 入门流程（安装 → 列表 → 开始 → 验证）
1) 按“快速开始”安装和构建  
2) 查看 Sprint：`gitta list`  
3) 开始任务：`gitta start US-001`  
4) 验证：检查当前分支与任务 frontmatter

### 日常流程（更新 → 列表 → 开始/继续 → 提交）
1) 更新代码：`git pull`  
2) 查看 Sprint：`gitta list`  
3) 开始或继续：`gitta start <task-id>`  
4) 随进度提交/推送；分支代表状态

### Sprint 规划（Sprint 与 backlog）
1) Sprint 列表：`gitta list`  
2) Sprint + backlog：`gitta list --all`  
3) 调整任务：通过移动 Markdown 位置管理 Sprint/backlog，执行 `gitta list --all` 验证

---

## 架构

六边形架构（端口-适配器）：
- **领域**：`internal/core`，`internal/services`
- **适配器**：`cmd/`（CLI），`infra/`（Git/文件系统），`ui/`（未来 TUI）
- **共享**：`pkg/` 工具库

详见 [docs/architecture.md](docs/architecture.md)。

---

## 开发

### 项目结构

```
cmd/gitta/          # CLI（Cobra）
internal/           # 领域逻辑
  core/             # 接口
  services/         # 实现
infra/              # Git、文件系统适配器
pkg/                # 工具库
tools/              # 开发工具
docs/               # 文档
```

### 测试

```bash
go test ./...
make verify  # 包含测试与 lint
```

### 新增命令

1) 创建命令文件：`cmd/gitta/<command>.go`  
2) 在 `cmd/gitta/root.go` 注册  
3) 在 `internal/services/` 实现服务  
4) 在 `docs/cli/<command>.md` 补充文档

更多见 `cmd/README.md`。

---

## 贡献

- 设置并验证：`go mod tidy && make verify`
- 遵循六边形边界（业务逻辑不放在 `cmd/`）
- 非 trivial 逻辑采用表驱动测试，CLI 流程补充集成测试
- 提 PR 时关联对应的 spec/plan，说明修改范围
- 架构参考： [docs/architecture.md](docs/architecture.md)  
- 命令参考： [cmd/README.md](cmd/README.md)

---

## 文档

- [架构指南](docs/architecture.md)
- [CLI 参考](docs/cli/)
- [快速开始](docs/quickstart.md)

---

## 支持

- 问题反馈：在 GitHub 提 issue，附带复现步骤和 CLI 输出
- 排查：重新运行 `gitta list --all` 检查任务位置和状态

---

## 许可证

使用 [MIT 许可证](LICENSE)。
