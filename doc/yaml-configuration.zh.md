# YAML 配置文件编写指南

## 概述

QLStat 使用 YAML 配置文件驱动整个分析流水线：克隆仓库 → 构建 CodeQL 数据库 → 生成外部谓词 → 查询 → 解码 → 收集结果。

典型配置文件：

```yaml
logRoot: logs
repoRoot: repos
dbRoot: codeql-db

repositories:
  - urlPrefix: https://github.com/
    dir: github.com
    repos:
      - org/repo
      - org/repo branch-or-commit

language: go
buildTimeout: 3600
buildGrps:
  - buildRepos: ["-"]
    buildCmd: ""
    extgenScript: goescape

queryconfig:
  resultRoot: codeqlResult
  parallelCore: 20
  queryGrps:
    - queryRoot: std
      queryDBs: ["-"]
      queries:
        - escape_ext/heapvar_should_move.ql
      externals: [movedToHeap]
```

---

## 顶层字段

| 字段 | 必须 | 默认值 | 说明 |
|------|------|--------|------|
| `logRoot` | 否 | `<projectRoot>/logs` | 运行时日志目录（clone/build/query/decode，自动归档） |
| `repoRoot` | 否 | `<projectRoot>/repos` | 源代码存放目录 |
| `dbRoot` | 否 | `<projectRoot>/codeql-db` | CodeQL 数据库存放目录 |
| `repositories` | 是* | — | 仓库列表。与下文 `sources` 二选一 |
| `language` | 是 | — | 分析语言，传给 `codeql database create -l`（如 `go`、`cpp`、`javascript`） |
| `buildTimeout` | 否 | 3600 | 单个仓库构建超时（秒） |
| `buildGrps` | 是 | — | 构建组列表（[见下文](#buildgrps定义构建组)） |
| `queryconfig` | 是 | — | 查询配置（[见下文](#queryconfig定义查询)） |

---

## repositories：定义仓库

`repositories` 是一个列表，每个条目可以指定一组来源相同（相同 urlPrefix 和目录前缀）的仓库。

### 字段

| 字段 | 必须 | 说明 |
|------|------|------|
| `urlPrefix` | 否 | 远程仓库的 URL 前缀。为空表示本地仓库，跳过克隆。 |
| `dir` | 否 | `repoRoot` 下的子目录。为空则仓库直接放在 `repoRoot/` 下。 |
| `repos` | 是 | 仓库名列表。远程仓库写 `org/repo`，本地仓库写目录名。 |

### 远程仓库

```yaml
repositories:
  - urlPrefix: https://github.com/
    dir: github.com
    repos:
      - rclone/rclone
      - cli/cli
```

效果：
- `git clone https://github.com/rclone/rclone.git` → `repoRoot/github.com/rclone/`
- `git clone https://github.com/cli/cli.git` → `repoRoot/github.com/cli/`

### 指定分支或提交

在仓库名后加空格，再写分支名、标签或 commit hash：

```yaml
repos:
  - rclone/rclone v1.72.1          # tag
  - Lslightly/dolt heapvar_should_move  # branch
  - cli/cli 28c187b                # commit hash
```

若仓库已存在，会执行 `git checkout` 切换。

### 本地仓库

省略 `urlPrefix`，使用 `dir` 指向 `repoRoot` 下的子目录：

```yaml
repositories:
  - dir: test/
    repos:
      - false-sharing
      - escape
```

效果：使用 `repoRoot/test/false-sharing/` 和 `repoRoot/test/escape/` 下的已有代码，跳过克隆。

### 行为矩阵

| urlPrefix | dir | 行为 | 磁盘路径 |
|-----------|-----|------|---------|
| `https://github.com/` | `github.com` | 远程克隆 | `repoRoot/github.com/org/repo/` |
| `https://github.com/` | 空 | 远程克隆 | `repoRoot/org/repo/` |
| 空 | `test/` | 本地，跳过克隆 | `repoRoot/test/repo/` |
| 空 | 空 | 本地，跳过克隆 | `repoRoot/repo/` |

---

## buildGrps：定义构建组

`buildGrps` 是构建组列表。每组定义一个构建命令和可选的外部谓词生成脚本，应用于一组仓库。

### 字段

| 字段 | 必须 | 说明 |
|------|------|------|
| `buildRepos` | 是 | 选择器或仓库全名列表。[见下方"选择器"](#选择器buildrepos-和-querydbs)。 |
| `buildCmd` | 否 | 构建命令。[见下方"构建命令类型"](#构建命令类型)。为空则让 CodeQL 自动检测。 |
| `extgenScript` | 否 | 外部谓词生成脚本。[见下方"外部谓词生成"](#外部谓词生成)。为空则跳过。 |
| `dbName` | 否 | 自定义数据库名称。用于同一仓库构建多次时区分不同的数据库。 |

### 选择器（`buildRepos` 和 `queryDBs`）

| 值 | 含义 |
|----|------|
| `["-"]` | 所有在 `repositories` 中定义过的仓库 |
| `["*"]` | `repoRoot/` 下磁盘上存在的所有仓库（自动扫描） |
| `["repo1", "repo2"]` | 仅指定的仓库（全名匹配） |

> 选择器写为单元素列表：`["-"]` 而非 `"-"`。

### 构建命令类型

```yaml
# 类型 1：空 → 让 codeql 自动检测构建命令
buildCmd: ""

# 类型 2：自定义构建脚本（路径相对于项目根目录）
buildCmd: yaml-examples/build/actor-demo.sh

# 类型 3：内联 shell 命令（注意只按照空格自动分割，不对引号语义感知）
buildCmd: go build -a ./...
```

使用构建脚本时，会设置以下环境变量：

| 环境变量 | 说明 |
|---------|------|
| `REPO_DIR` | 仓库根目录 |
| `PROJROOT` | 项目根目录 |

### 外部谓词生成

```yaml
# 类型 1：goescape → 运行 go build -gcflags=-m=2，解析逃逸分析输出
extgenScript: goescape

# 类型 2：自定义脚本（路径相对于项目根目录）
extgenScript: yaml-examples/extgens/dolt.sh
```

使用自定义脚本时设置的环境变量：

| 环境变量 | 说明 |
|---------|------|
| `PROJROOT` | 项目根目录 |
| `REPO_DIR` | 仓库根目录 |
| `OUTPUT_DIR` | 中间结果输出目录 |
| `DB_EXT_DIR` | 外部谓词数据库存放目录（`dbRoot/<repo>/ext/`） |

### dbName：同仓库多构建

同一仓库需要以不同方式构建分析时，使用 `dbName` 避免数据库冲突：

```yaml
buildGrps:
  # 默认构建
  - buildRepos: [rclone/rclone]
    extgenScript: goescape            # 数据库: codeql-db/rclone/

  # 同一仓库，不同构建命令 → 需要显式 dbName
  - buildRepos: [rclone/rclone]
    dbName: rclone-custom
    buildCmd: go build -a ./...
    extgenScript: goescape            # 数据库: codeql-db/rclone-custom/
```

> `dbName` 必须在所有构建组中唯一。

---

## queryconfig：定义查询

```yaml
queryconfig:
  resultRoot: codeqlResult
  parallelCore: 20
  queryGrps:
    - queryRoot: std
      queryDBs: ["-"]
      queries:
        - escape_ext/heapvar_should_move.ql
      externals: [movedToHeap]
      externalFiles:
        - yaml-template/pprof.yaml
```

### 字段

| 字段 | 必须 | 默认值 | 说明 |
|------|------|--------|------|
| `resultRoot` | 否 | `<projectRoot>/codeqlResult` | 查询结果根目录 |
| `parallelCore` | 否 | 20 | 并行核数 |

### queryGrps

`queryGrps` 是查询组列表。每组定义查询哪些数据库、运行哪些查询。

| 字段 | 必须 | 说明 |
|------|------|------|
| `queryRoot` | 否 | 查询根目录。`"std"` 或空 → `<projectRoot>/qlsrc`。可指定绝对路径或相对路径。 |
| `externalRoot` | 否 | 外部谓词模板文件的根目录。`"std"` 或空 → `<projectRoot>/qlsrc`。 |
| `queryDBs` | 是 | 目标数据库选择器。与 `buildRepos` 相同（`["-"]`、`["*"]` 或全名列表）。 |
| `queries` | 是 | 查询文件路径列表。路径相对于 `queryRoot`。 |
| `externals` | 否 | 外部谓词名称列表。 |
| `externalFiles` | 否 | YAML 文件路径列表。每个文件包含一个外部谓词名列表，与 `externals` 合并。 |

### externalRoot

`externalRoot` 决定 `externalFiles` 中相对路径的解析起点。与 `queryRoot` 行为一致：

| 值 | 解析行为 |
|----|---------|
| `"std"` 或空 | `<projectRoot>/qlsrc` |
| 绝对路径 | 直接使用 |
| 相对路径 | 从当前工作目录解析 |

示例：

```yaml
# externalRoot 为 "std"，externalFiles 使用相对路径
# → 实际读取 <projectRoot>/qlsrc/yaml-template/pprof.yaml
queryGrps:
  - queryDBs: [malloc_test]
    externalRoot: std
    externalFiles:
      - yaml-template/pprof.yaml
```

```yaml
# externalRoot 为绝对路径，externalFiles 使用相对路径
# → 实际读取 /custom/templates/my_predicates.yaml
queryGrps:
  - queryDBs: [malloc_test]
    externalRoot: /custom/templates
    externalFiles:
      - my_predicates.yaml
```

### externals 和 externalFiles

`externals` 直接指定谓词名：

```yaml
externals: [movedToHeap, newEscapesToHeap]
```

`externalFiles` 引用模板 YAML 文件（文件内容为顶层列表）：

```yaml
externalFiles:
  - yaml-template/pprof.yaml
```

模板文件 `qlsrc/yaml-template/pprof.yaml` 内容：

```yaml
- profile
- value_type
- sample
- sample_to_location_id
- ...
```

> `externals` 和 `externalFiles` 的结果会合并去重。适用于多个查询组复用同一套外部谓词。

### queryDBs 选择器

与 `buildRepos` 相同：

| 值 | 含义 |
|----|------|
| `["-"]` | 查询所有在 `buildGrps` 中构建过的数据库 |
| `["*"]` | 查询 `dbRoot/` 下所有数据库 |
| `["rclone", "rclone-custom"]` | 仅查询指定名称的数据库 |

### 数据库名称解析

`queryDBs` 中使用的是数据库的 **dbName**（显式设置的值或默认的仓库名），**不是** `repositories.repos` 中写的全名：

```
repositories:
  - repos:
    - rclone/rclone       # 全名
# 默认 dbName: rclone

queryconfig:
  queryGrps:
    - queryDBs: [rclone]  # ✓ 正确：匹配默认 dbName
```

---

## 完整示例

### 最小配置

```yaml
logRoot: logs
repoRoot: repos
dbRoot: codeql-db

repositories:
  - dir: test/
    repos:
      - escape

language: go
buildTimeout: 3600
buildGrps:
  - buildRepos: ["-"]
    extgenScript: goescape

queryconfig:
  resultRoot: codeqlResult
  parallelCore: 20
  queryGrps:
    - queryDBs: ["-"]
      queries:
        - escape_ext/heapvar_should_move.ql
      externals: [movedToHeap]
```

### 全量配置（远程 + 本地混合）

```yaml
logRoot: logs
repoRoot: repos
dbRoot: codeql-db

repositories:
  - urlPrefix: https://github.com/
    dir: github.com
    repos:
      - rclone/rclone
      - Lslightly/dolt heapvar_should_move
  - dir: test/
    repos:
      - false-sharing
      - malloc_test

language: go
buildTimeout: 3600
buildGrps:
  - buildRepos: ["-"]
    buildCmd: ""
    extgenScript: goescape

queryconfig:
  resultRoot: codeqlResult
  parallelCore: 20
  queryGrps:
    - queryDBs: ["-"]
      queries:
        - escape_ext/heapvar_should_move.ql
      externals: [movedToHeap]
    - queryDBs: [malloc_test]
      queries:
        - pprof_ext/use_ext_directly/profile.ql
      externalFiles:
        - yaml-template/pprof.yaml
```

---

## 常用模式

### 不同查询组共享外部谓词

```yaml
queryGrps:
  - queryDBs: [repo-a]
    queries:
      - escape_ext/query_a.ql
    externals: [movedToHeap,newEscapesToHeap]
  - queryDBs: [repo-b]
    queries:
      - escape_ext/query_b.ql
    externals: [movedToHeap,newEscapesToHeap]      # 重复写
```

→ 用 `externalFiles` 提取公共部分：

```yaml
queryGrps:
  - queryDBs: [repo-a]
    queries:
      - escape_ext/query_a.ql
    externalFiles:
      - yaml-template/escape.yaml
  - queryDBs: [repo-b]
    queries:
      - escape_ext/query_b.ql
    externalFiles:
      - yaml-template/escape.yaml
```

`yaml-template/escape.yaml`：

```yaml
- movedToHeap
- newEscapesToHeap
```

### 同仓库多维度分析

```yaml
buildGrps:
  # 默认构建
  - buildRepos: [rclone/rclone]
  # 自定义构建（不同编译选项）
  - buildRepos: [rclone/rclone]
    dbName: rclone-custom
    buildCmd: go build -a ./...

queryGrps:
  - queryDBs: [rclone, rclone-custom]
    queries:
      - escape_ext/heapvar_should_move.ql
    externals: [movedToHeap]
```

### "*" 选择器：扫描全部

```yaml
buildRepos: ["*"]     # 构建 repoRoot/ 下所有存在的仓库
queryDBs: ["*"]       # 查询 dbRoot/ 下所有数据库
```

---

## 常见错误

### 仓库名 vs 数据库名混淆

```
repositories:
  - repos:
    - github.com/org/repo   # ✗ 不要写 URL 前缀，只要 org/repo
```

```
queryGrps:
  - queryDBs: [org/repo]    # ✗ queryDBs 用数据库名（dbName），不是仓库全名
```

### 选择器格式错误

```yaml
buildRepos: "-"   # ✗ 选择器必须为列表
buildRepos: ["-"] # ✓
```

### 数据库名冲突

```yaml
buildGrps:
  - buildRepos: [repo-a]    # 默认 dbName = repo-a
  - buildRepos: [repo-b]    # 默认 dbName = repo-b
  - buildRepos: [repo-a]    # ✗ 与第一个冲突！添加 dbName 区分
```

### 空 buildCmd 导致构建失败

部分语言/项目需要显式指定构建命令。若 `buildCmd: ""` 导致构建失败：

```yaml
buildCmd: go build -a ./...       # Go 项目
buildCmd: yaml-examples/build/project.sh  # 自定义构建脚本
```

---

## 存储结构

运行后目录结构：

```
repoRoot/
  github.com/
    rclone/
    cli/
  test/
    false-sharing/

codeql-db/
  rclone/                    # dbName
    ext/                     # 外部谓词 CSV
      movedToHeap.csv
    extgen/                  # 外部谓词中间输出
  false-sharing/
    ext/

codeqlResult/
  escape_ext/
    heapvar_should_move.ql/  # 查询路径
      rclone.bqrs            # 原始结果
      rclone.csv             # 解码结果
      false-sharing.bqrs
      false-sharing.csv
  escape_ext/
    heapvar_should_move.csv  # collect 合并结果

logs/
  clone/current/
  build/current/
  query/current/
  decode/current/
```

---

## 附录：所有 YAML 标签速查

| YAML 字段 | 类型 | Go 结构体 | 所属 |
|-----------|------|-----------|------|
| `logRoot` | string | `Artifact.LogRoot` | 顶层 |
| `repoRoot` | string | `Artifact.RepoRoot` | 顶层 |
| `dbRoot` | string | `Artifact.DBRoot` | 顶层 |
| `repositories` | `[]*RepoGroup` | `Artifact.Repositories` | 顶层 |
| `language` | string | `Artifact.Lang` | 顶层 |
| `buildTimeout` | int | `Artifact.BuildTimeout` | 顶层 |
| `buildGrps` | `[]BuildGroup` | `Artifact.BuildGrps` | 顶层 |
| `buildRepos` | `[]string` | `BuildGroup.BuildRepos` | buildGrps[]. |
| `buildCmd` | string | `BuildGroup.BuildCommand` | buildGrps[]. |
| `extgenScript` | string | `BuildGroup.ExtGenScript` | buildGrps[]. |
| `dbName` | string | `BuildGroup.DBName` | buildGrps[]. |
| `resultRoot` | string | `QueryConfig.ResultRoot` | queryconfig |
| `parallelCore` | int | `QueryConfig.ParallelCore` | queryconfig |
| `queryGrps` | `[]QueryGroup` | `QueryConfig.QueryGrps` | queryconfig |
| `queryRoot` | string | `QueryGroup.QueryRoot` | queryGrps[]. |
| `externalRoot` | string | `QueryGroup.ExternalRoot` | queryGrps[]. |
| `queryDBs` | `[]string` | `QueryGroup.QueryDBs` | queryGrps[]. |
| `queries` | `[]string` | `QueryGroup.Queries` | queryGrps[]. |
| `externals` | `[]string` | `QueryGroup.Externals` | queryGrps[]. |
| `externalFiles` | `[]string` | `QueryGroup.ExternalFiles` | queryGrps[]. |
