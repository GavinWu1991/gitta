# Gitta 设置指南

本指南帮助你在现有 Git 仓库中快速使用 gitta：下载预构建二进制、校验完整性、初始化目录结构，以及运行发布与质量门禁流程。

## 1. 获取与安装

### 1.1 预构建二进制（推荐）

1) 打开 GitHub Releases 页面，选择目标版本。  
2) 下载对应平台压缩包：  
   - macOS：`gitta-<version>-darwin-amd64.tar.gz`（Intel）或 `darwin-arm64.tar.gz`（Apple Silicon）  
   - Linux：`gitta-<version>-linux-amd64.tar.gz` 或 `linux-arm64.tar.gz`  
   - Windows：`gitta-<version>-windows-amd64.zip` 或 `windows-arm64.zip`  
3) 校验完整性（推荐）：  
   ```bash
   shasum -a 256 gitta-<version>-<platform>-<arch>.tar.gz
   # 与发布页的 checksums.txt 对比
   ```
4) 解压并加入 PATH：  
   ```bash
   tar -xzf gitta-<version>-darwin-amd64.tar.gz   # macOS/Linux
   unzip gitta-<version>-windows-amd64.zip        # Windows
   sudo mv gitta /usr/local/bin/                  # 可选
   gitta --help
   ```

### 1.2 从源码构建

```bash
git clone https://github.com/GavinWu1991/gitta.git
cd gitta
go mod tidy
go build -o gitta ./cmd/gitta
./gitta --help
```

## 2. 初始化现有项目（init 脚本）

> 适用于任意已有 Git 仓库，生成 gitta 所需目录和示例任务。

```bash
# 假设已下载脚本
./scripts/init.sh

# 若已初始化且需重建
./scripts/init.sh --force                    # 备份现有目录后重建
./scripts/init.sh --example-sprint Sprint-02 # 自定义示例 Sprint 名称
```

脚本行为：
- 创建 `sprints/<Sprint-01>/` 与 `backlog/` 目录
- 写入示例任务文件（US-001、US-002），格式符合 gitta 要求
- 如果已存在目录：
  - 默认跳过并提示
  - 使用 `--force` 时会先备份再重建
- 需在 Git 仓库根目录运行（检测 `.git/`）

完成后运行：
```bash
gitta list       # 查看示例任务
gitta list --all # 查看 Sprint + backlog
```

## 3. 发布与回滚（维护者）

### 3.1 创建发布
```bash
# 确认工作区干净，测试通过
make verify

# 打 tag（语义化版本）
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions 会：
# - 运行质量门禁（make verify）
# - 调用 GoReleaser 构建 6 个平台/架构的产物
# - 生成 checksums.txt 并附到 Release
```

### 3.2 快速回滚
```bash
# 删除远端标签与发布（需在 GitHub 上删除 Release）
git push origin :refs/tags/v1.0.0
# 如需重新发布，修复问题后重新打相同或更高版本的 tag
```

## 4. 质量门禁与排查

- GitHub Actions `quality` 作业运行 `make verify`（含 gofmt/goimports/vet/staticcheck/govulncheck/tests/架构检查）。
- 失败时：发布作业会被跳过，查看 Actions 日志定位失败步骤。
- 本地复现：`make verify`，或用 `make release-snapshot` 进行本地 GoReleaser 干跑（不发布）。

## 5. 常见问题

- **不是 Git 仓库**：在项目根目录执行 `git init` 后再运行脚本。  
- **权限不足**：确保对仓库有写权限，必要时以合适权限运行脚本。  
- **校验和不一致**：重新下载，确认使用对应平台/架构文件并比对 checksums.txt。  
- **平台不支持**：若缺少你的平台，请按 README 的源码构建步骤本地编译。  

## 6. 发布说明与变更记录（参考）

- 发布说明来源于 Git 提交（GoReleaser changelog），自动附加到 Release。
- 如需人工补充，可在 GitHub Release 页面编辑说明。
- 推荐的发布说明模板（可复制到 Release 描述中）：  
  ```
  ## 变更摘要
  - 新功能：...
  - 修复：...
  - 文档/开发体验：...
  - 破坏性变更：无/有（说明迁移方式）

  ## 校验
  - 产物：macOS amd64/arm64，Linux amd64/arm64，Windows amd64/arm64
  - 校验和：附带 checksums.txt
  - 质量门禁：make verify 通过
  ```

## 7. 目录一览（初始化后）

```
<repo>/
├── sprints/
│   └── Sprint-01/
│       └── US-001.md   # 示例 Sprint 任务
└── backlog/
    └── US-002.md       # 示例 backlog 任务
```

完成以上步骤后，即可用 gitta 的常用命令：
```bash
gitta list        # 查看当前 Sprint
gitta list --all  # 查看 Sprint + backlog
gitta start <id>  # 为任务创建/切换分支
```
