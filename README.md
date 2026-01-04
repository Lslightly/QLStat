# QLStat

Analyze real-world project batch with declarative static analysis provided by CodeQL for empirical study and statistical analysis to gain insight of patterns in real world projects.

## Features Overview

QLStat provides a comprehensive framework for large-scale empirical analysis of software projects using CodeQL. Key features include:

- **Batch Processing**: Clone, build, and analyze multiple repositories in parallel
- **Flexible Configuration**: YAML-based configuration for defining analysis targets and parameters
- **Extensible Analysis**: Support for custom external predicates (e.g., escape analysis data)
- **Scalable Query Execution**: Parallel execution of CodeQL queries across repositories
- **Comprehensive Logging**: Detailed logging at each stage of the analysis pipeline
- **Data Collection**: Aggregation of results from multiple repositories into unified datasets
- **Language Support**: Currently focused on Go, with extensibility for other languages supported by CodeQL

## Usage

### 1. Configuration

Create your `stat.yaml` config file according to [`example.yaml`](./example.yaml). The configuration supports several key sections:

- `sources`: Define repository sources with prefixes and specific repositories
- `language`: Specify the programming language for analysis (e.g., go)
- `buildGrps`: Configure build groups with timeout and build commands
- `externalGenGrps`: Generate external predicates (like escape analysis data)
- `queryconfig`: Set up query execution with parallelization options
- `queryGrps`: Define query groups with specific queries and target repositories

### 2. Database Creation

Run `go run ./cmd/batch_clone_build stat.yaml` to clone repositories and create CodeQL databases:

```bash
go run ./cmd/batch_clone_build stat.yaml
```

Key options:
- `-noclone`: Skip cloning if repositories already exist
- `-nobuild`: Skip database creation if databases already exist
- `-noextgen`: Skip generation of external predicates

The tool supports three main phases:
1. **Cloning**: Download repositories from specified sources
2. **Building**: Create CodeQL databases using appropriate build commands
3. **External Predicate Generation**: Generate additional data sources like escape analysis results

### 3. Query Development

Create your queries in the [`qlsrc`](./qlsrc/) directory. Queries should follow CodeQL conventions and can leverage external predicates when needed.

### 4. Query Execution

Run `go run ./cmd/codeql_qdriver -collect stat.yaml` to execute queries on the created databases:

```bash
go run ./cmd/codeql_qdriver -collect stat.yaml
```

Available options:
- `-format`: Specify output format (text, csv, json, bqrs) - default: csv
- `-decode-only`: Only decode existing bqrs files without running queries
- `-collect`: Collect all CSV results into a single file with repository names

Results are processed in three stages:
1. **Query Execution**: Run CodeQL queries on each database
2. **Decoding**: Convert bqrs results to specified format (CSV, JSON, etc.)
3. **Collection**: Aggregate results from all repositories into a single dataset

## Extensions

### Go Escape Analysis Extension

QLStat supports extending CodeQL with escape analysis data through the escape adapter:

1. Configure `externalGenGrps` in your YAML with `genScript: goescape`.
   1. `goescape` is actually the command `go build -a -gcflags=all=-m=2 .`
   2. You can also specify your own script with only one constraint: Generate `m2.log` in `$logRoot/extgen/path/to/repo/m2.log`
2. This generates escape analysis data during the build phase
3. Reference the external predicate in your query group with `externals: - movedToHeap`
4. Use the external predicate in your CodeQL queries

For more details about how the escape analysis extension works, see [Escape Analysis Documentation](doc/adapters/escape_analysis.md).

## Architecture

For detailed information about the storage structure and architecture, please refer to the [Architecture Documentation](doc/arch.md).

# Citation

```bibtex
@misc{qlstat,
    author = {Qingwei Li},
    title = {QLStat},
    howpublished = {\url{https://github.com/Lslightly/QLStat}},
}
```
