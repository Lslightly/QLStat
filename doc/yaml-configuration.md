# YAML Configuration Guide

## Overview

QLStat uses a YAML configuration file to drive the entire analysis pipeline: clone repositories → build CodeQL databases → generate external predicates → query → decode → collect results.

Typical configuration:

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

## Top-Level Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `logRoot` | No | `<projectRoot>/logs` | Runtime log directory (clone/build/query/decode, auto-archived) |
| `repoRoot` | No | `<projectRoot>/repos` | Source code storage directory |
| `dbRoot` | No | `<projectRoot>/codeql-db` | CodeQL database directory |
| `repositories` | Yes* | — | Repository list. Mutually exclusive with `sources` below. |
| `language` | Yes | — | Analysis language, passed to `codeql database create -l` (e.g. `go`, `cpp`, `javascript`) |
| `buildTimeout` | No | 3600 | Build timeout per repository (seconds) |
| `buildGrps` | Yes | — | Build group list ([see below](#buildgrps-defining-build-groups)) |
| `queryconfig` | Yes | — | Query configuration ([see below](#queryconfig-defining-queries)) |

---

## repositories: Defining Repositories

`repositories` is a list. Each entry specifies a group of repositories that share the same origin (same `urlPrefix` and directory prefix).

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `urlPrefix` | No | Remote repository URL prefix. If empty, the repo is local (skip clone). |
| `dir` | No | Subdirectory under `repoRoot`. If empty, repos go directly under `repoRoot/`. |
| `repos` | Yes | List of repository names. Remote repos use `org/repo`; local repos use directory names. |

### Remote Repositories

```yaml
repositories:
  - urlPrefix: https://github.com/
    dir: github.com
    repos:
      - rclone/rclone
      - cli/cli
```

Result:
- `git clone https://github.com/rclone/rclone.git` → `repoRoot/github.com/rclone/`
- `git clone https://github.com/cli/cli.git` → `repoRoot/github.com/cli/`

### Pinning a Branch or Commit

Append a space after the repo name, then write a branch name, tag, or commit hash:

```yaml
repos:
  - rclone/rclone v1.72.1          # tag
  - Lslightly/dolt heapvar_should_move  # branch
  - cli/cli 28c187b                # commit hash
```

If the repository already exists, `git checkout` is used to switch.

### Local Repositories

Omit `urlPrefix` and use `dir` to point to a subdirectory under `repoRoot`:

```yaml
repositories:
  - dir: test/
    repos:
      - false-sharing
      - escape
```

Result: uses existing code at `repoRoot/test/false-sharing/` and `repoRoot/test/escape/`, skipping clone.

### Behavior Matrix

| urlPrefix | dir | Behavior | On-disk path |
|-----------|-----|----------|--------------|
| `https://github.com/` | `github.com` | Remote clone | `repoRoot/github.com/org/repo/` |
| `https://github.com/` | empty | Remote clone | `repoRoot/org/repo/` |
| empty | `test/` | Local, skip clone | `repoRoot/test/repo/` |
| empty | empty | Local, skip clone | `repoRoot/repo/` |

---

## buildGrps: Defining Build Groups

`buildGrps` is a list of build groups. Each group defines a build command and an optional external predicate generation script, applied to a set of repositories.

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `buildRepos` | Yes | Selector or list of full repository names. See ["Selector"](#selector-buildrepos-and-querydbs) below. |
| `buildCmd` | No | Build command. See ["Build command types"](#build-command-types) below. Empty lets CodeQL auto-detect. |
| `extgenScript` | No | External predicate generation script. See ["External predicate generation"](#external-predicate-generation) below. Empty skips it. |
| `dbName` | No | Custom database name. Used to differentiate databases when building the same repo multiple times. |

### Selector (`buildRepos` and `queryDBs`)

| Value | Meaning |
|-------|---------|
| `["-"]` | All repositories defined in `repositories` |
| `["*"]` | All repositories that exist on disk under `repoRoot/` (auto-scan) |
| `["repo1", "repo2"]` | Only the specified repositories (matched by full name) |

> Selectors must be written as single-element lists: `["-"]`, not `"-"`.

### Build Command Types

```yaml
# Type 1: empty → let codeql auto-detect the build command
buildCmd: ""

# Type 2: custom build script (path relative to project root)
buildCmd: yaml-examples/build/actor-demo.sh

# Type 3: inline shell command (arguments split by whitespace; does not parse quotes)
buildCmd: go build -a ./...
```

When using a build script, the following environment variables are set:

| Variable | Description |
|----------|-------------|
| `REPO_DIR` | Repository root directory |
| `PROJROOT` | Project root directory |

### External Predicate Generation

```yaml
# Type 1: goescape → runs go build -gcflags=-m=2, parses escape analysis output
extgenScript: goescape

# Type 2: custom script (path relative to project root)
extgenScript: yaml-examples/extgens/dolt.sh
```

Environment variables set for custom scripts:

| Variable | Description |
|----------|-------------|
| `PROJROOT` | Project root directory |
| `REPO_DIR` | Repository root directory |
| `OUTPUT_DIR` | Directory for intermediate output |
| `DB_EXT_DIR` | External predicate database directory (`dbRoot/<repo>/ext/`) |

### dbName: Multiple Builds for the Same Repository

When the same repository needs to be built with different configurations, use `dbName` to avoid database conflicts:

```yaml
buildGrps:
  # Default build
  - buildRepos: [rclone/rclone]
    extgenScript: goescape            # Database: codeql-db/rclone/

  # Same repository, different build command → explicit dbName needed
  - buildRepos: [rclone/rclone]
    dbName: rclone-custom
    buildCmd: go build -a ./...
    extgenScript: goescape            # Database: codeql-db/rclone-custom/
```

> `dbName` must be unique across all build groups.

---

## queryconfig: Defining Queries

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

### Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `resultRoot` | No | `<projectRoot>/codeqlResult` | Query result root directory |
| `parallelCore` | No | 20 | Parallelism for query execution |

### queryGrps

`queryGrps` is a list of query groups. Each group defines which databases to query and which queries to run.

| Field | Required | Description |
|-------|----------|-------------|
| `queryRoot` | No | Query root directory. `"std"` or empty → `<projectRoot>/qlsrc`. Absolute or relative paths are accepted. |
| `externalRoot` | No | Root directory for external predicate template files. `"std"` or empty → `<projectRoot>/qlsrc`. |
| `queryDBs` | Yes | Target database selector. Same as `buildRepos` (`["-"]`, `["*"]`, or full name list). |
| `queries` | Yes | List of query file paths. Paths are relative to `queryRoot`. |
| `externals` | No | List of external predicate names. |
| `externalFiles` | No | List of YAML file paths. Each file contains a list of external predicate names, merged with `externals`. |

### externalRoot

`externalRoot` determines where relative paths in `externalFiles` are resolved from. Same behavior as `queryRoot`:

| Value | Resolution |
|-------|------------|
| `"std"` or empty | `<projectRoot>/qlsrc` |
| Absolute path | Used as-is |
| Relative path | Resolved from the current working directory |

Examples:

```yaml
# externalRoot is "std", externalFiles uses relative paths
# → reads <projectRoot>/qlsrc/yaml-template/pprof.yaml
queryGrps:
  - queryDBs: [malloc_test]
    externalRoot: std
    externalFiles:
      - yaml-template/pprof.yaml
```

```yaml
# externalRoot is an absolute path, externalFiles uses relative paths
# → reads /custom/templates/my_predicates.yaml
queryGrps:
  - queryDBs: [malloc_test]
    externalRoot: /custom/templates
    externalFiles:
      - my_predicates.yaml
```

### externals and externalFiles

`externals` directly specifies predicate names:

```yaml
externals: [movedToHeap, newEscapesToHeap]
```

`externalFiles` references template YAML files (with a top-level list):

```yaml
externalFiles:
  - yaml-template/pprof.yaml
```

Contents of template file `qlsrc/yaml-template/pprof.yaml`:

```yaml
- profile
- value_type
- sample
- sample_to_location_id
- ...
```

> Results from `externals` and `externalFiles` are merged and deduplicated. Useful when multiple query groups share the same set of external predicates.

### queryDBs Selector

Same as `buildRepos`:

| Value | Meaning |
|-------|---------|
| `["-"]` | Query all databases built in `buildGrps` |
| `["*"]` | Query all databases under `dbRoot/` |
| `["rclone", "rclone-custom"]` | Query only the specified databases (by name) |

### Database Name Resolution

`queryDBs` uses the **dbName** (explicitly set value or the default repo name), **not** the full name from `repositories.repos`:

```
repositories:
  - repos:
    - rclone/rclone       # full name
# default dbName: rclone

queryconfig:
  queryGrps:
    - queryDBs: [rclone]  # ✓ correct: matches default dbName
```

---

## Complete Examples

### Minimal Configuration

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

### Full Configuration (Remote + Local Mix)

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

## Common Patterns

### Sharing External Predicates Across Query Groups

```yaml
queryGrps:
  - queryDBs: [repo-a]
    queries:
      - escape_ext/query_a.ql
    externals: [movedToHeap,newEscapesToHeap]
  - queryDBs: [repo-b]
    queries:
      - escape_ext/query_b.ql
    externals: [movedToHeap,newEscapesToHeap]      # repeated
```

→ Extract the common part with `externalFiles`:

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

`yaml-template/escape.yaml`:

```yaml
- movedToHeap
- newEscapesToHeap
```

### Multi-Dimension Analysis of the Same Repository

```yaml
buildGrps:
  # Default build
  - buildRepos: [rclone/rclone]
  # Custom build (different compile options)
  - buildRepos: [rclone/rclone]
    dbName: rclone-custom
    buildCmd: go build -a ./...

queryGrps:
  - queryDBs: [rclone, rclone-custom]
    queries:
      - escape_ext/heapvar_should_move.ql
    externals: [movedToHeap]
```

### The "*" Selector: Scan Everything

```yaml
buildRepos: ["*"]     # Build all repos existing under repoRoot/
queryDBs: ["*"]       # Query all databases under dbRoot/
```

---

## Common Mistakes

### Confusing Repository Names with Database Names

```
repositories:
  - repos:
    - github.com/org/repo   # ✗ don't include URL prefix, just org/repo
```

```
queryGrps:
  - queryDBs: [org/repo]    # ✗ queryDBs uses dbName, not the full repository name
```

### Wrong Selector Format

```yaml
buildRepos: "-"   # ✗ selector must be a list
buildRepos: ["-"] # ✓
```

### Database Name Conflicts

```yaml
buildGrps:
  - buildRepos: [repo-a]    # default dbName = repo-a
  - buildRepos: [repo-b]    # default dbName = repo-b
  - buildRepos: [repo-a]    # ✗ conflicts with the first! Add dbName to distinguish
```

### Empty BuildCmd Causes Build Failure

Some languages/projects need an explicit build command. If `buildCmd: ""` fails:

```yaml
buildCmd: go build -a ./...       # Go project
buildCmd: yaml-examples/build/project.sh  # custom build script
```

---

## Storage Layout

Directory structure after a run:

```
repoRoot/
  github.com/
    rclone/
    cli/
  test/
    false-sharing/

codeql-db/
  rclone/                    # dbName
    ext/                     # external predicate CSV
      movedToHeap.csv
    extgen/                  # external predicate intermediate output
  false-sharing/
    ext/

codeqlResult/
  escape_ext/
    heapvar_should_move.ql/  # query path
      rclone.bqrs            # raw result
      rclone.csv             # decoded result
      false-sharing.bqrs
      false-sharing.csv
  escape_ext/
    heapvar_should_move.csv  # collect merged result

logs/
  clone/current/
  build/current/
  query/current/
  decode/current/
```

---

## Appendix: YAML Tag Reference

| YAML Field | Type | Go Struct | Belongs To |
|------------|------|-----------|------------|
| `logRoot` | string | `Artifact.LogRoot` | Top-level |
| `repoRoot` | string | `Artifact.RepoRoot` | Top-level |
| `dbRoot` | string | `Artifact.DBRoot` | Top-level |
| `repositories` | `[]*RepoGroup` | `Artifact.Repositories` | Top-level |
| `language` | string | `Artifact.Lang` | Top-level |
| `buildTimeout` | int | `Artifact.BuildTimeout` | Top-level |
| `buildGrps` | `[]BuildGroup` | `Artifact.BuildGrps` | Top-level |
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
