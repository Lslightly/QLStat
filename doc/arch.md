# Project Architecture

## Components

```bash
cmd/
   batch_clone_build      # clone repos, build databases (with optional extgen)
   codeql_qdriver         # run queries, decode results, collect CSVs
   escape_adapter         # convert escape analysis log (m=2) to CSV external predicates
   pprof2qlcsv            # convert pprof profiles to CSV external predicates
   pprof-external-verify  # verify integrity of pprof external predicate CSV files
qlsrc/                 # CodeQL query sources
   escape_ext/         #   escape analysis queries (movedToHeap, escaped_loc, ...)
   pprof_ext/          #   pprof external predicate queries (malloc_time, ...)
```

## Storage Structure

```bash
repos/                 # source code (cloned by batch_clone_build)
   github.com/         #   hostname from GitSource.Prefix
      repo0/
   test/
      repo1/

codeql-db/             # CodeQL databases (built from repos)
   repo0/              #   database named by repo DirBaseName/specified customly
      ext/             #   external predicate CSV files
      extgen/          #   intermediate output directory for external predicates
   repo1/
      ext/

codeqlResult/          # query & decode output
   path/to/
      queryName/       #   per-query result directory
         repo0.bqrs    #     raw result
         repo0.csv     #     decoded result
      queryName.csv    #   collected results across all repos (optional)

logs/                  # runtime logs, archived automatically
   build/              # phase: build
      current/         #   latest run
         path/to/repo/ #     per-repo stdout/stderr
            out
            err
         repoTimes.csv #     timing summary
         repo_build.txt#     build status summary
      archive/<ts>/    #   historical runs
   clone/
      current/
         fail.log      #   clone failures
      archive/<ts>/
   query/
      current/query/path/
         repo0.out     #     per-repo query run stdout
         repo0.err     #     per-repo query run stderr
      archive/<ts>/
   decode/
      current/query/path/
         repo0.out     #     decode stdout
         repo0.err     #     decode stderr
      archive/<ts>/
```

## Workflow Pipeline

The toolchain has 6 stages. Each takes input from its predecessor and writes
to a well-known directory. Stages are independent — you can re-run any stage
without re-running earlier ones (provided the input data already exists).

```
                   sources[*].fullnames
                           │
                           ▼
                    ┌──────────────┐
                    │ 1. Clone     │──→ repoRoot/<host>/<repo>/
                    └──────────────┘
                           │
                           ▼
    BuildGrps[*].BuildCmd ─┤  (empty = default build)
                           ▼
                    ┌──────────────┐
                    │ 2. Build     │──→ dbRoot/<repo>/
                    └──────────────┘
                           │
    BuildGrps[*].ExtGenScript ─┤
                           ▼
                    ┌──────────────┐
                    │ 3. External  │──→ dbRoot/<repo>/ext/<pred>.csv
                    │    Gen       │
                    └──────────────┘
                           │
        queries[*] ────────┤
                           ▼
                    ┌──────────────┐
                    │ 4. Query     │──→ resultRoot/<queryPath>/<repo>.bqrs
                    └──────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ 5. Decode    │──→ resultRoot/<queryPath>/<repo>.csv
                    └──────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ 6. Collect   │──→ resultRoot/<queryPath>.csv
                    └──────────────┘
```

### 1. Clone — `batch_clone_build` (Phase: clone)

```
Input:   YAML.sources[*].fullnames          "org/repo [branch]"
         YAML.sources[*].prefix             "https://github.com/"
Process: git clone <prefix+fullname>.git    repoRoot/<host>/<DirBaseName>
         or (if exists) git checkout <branch>
Output:  repoRoot/<host>/<DirBaseName>/     source code on disk
```

The clone phase only touches `repoRoot/`. It populates the source tree that all
downstream phases depend on. Branch/commit pinning happens here via `Checkout`.

### 2. Build — `batch_clone_build` (Phase: build)

```
Input:   repoRoot/<host>/<DirBaseName>/      source code
         YAML.BuildGrps[*].BuildRepos        which repos to build
         YAML.BuildGrps[*].DBName            custom database name (optional, defaults to repo DirBaseName)
         YAML.BuildGrps[*].BuildCmd          command | script (empty/omitted = auto-detect)
         YAML.language                       "go", "cpp", ...
Process: codeql database create dbRoot/<DirBaseName>  \
           -l=<lang> --overwrite                         \
           -s=<repoDir>                                  \
           [ -c <buildCmd> ]
         If DBName is set, the database is created at dbRoot/<DBName>/
         instead of dbRoot/<DirBaseName>/. This is useful for
         combining a repo with a custom build command into one DB.
Output:  dbRoot/<DirBaseName>/               CodeQL database (snapshot)
```

`codeql database create` invokes the language-specific extractor. With an empty
or omitted `buildCmd`, the extractor infers the build system. With a custom
command or script, it wraps that invocation. The output is a relocatable CodeQL
database under `dbRoot/`.

### 3. External Predicate Generation — `batch_clone_build` (Phase: extgen)

```
Input:   repoRoot/<host>/<DirBaseName>/     source code (for compilation)
         YAML.BuildGrps[*].ExtGenScript     "goescape" | custom script | "" (skip)
         YAML.BuildGrps[*].BuildRepos       which repos to process
         dbRoot/<DirBaseName>/              target database (for ext dir)
Process: "goescape":
           go build -a -gcflags=-m=2 ./... 2> log
           → escape_adapter -dir dbRoot/<DirBaseName>/ext/  \
                          -src <repoDir>                    \
                          -movedToHeap -newEscapesToHeap <log>
         custom script:
           <script> with env REPO_DIR, OUTPUT_DIR, PROJROOT, DB_EXT_DIR
Output:  dbRoot/<DirBaseName>/ext/<pred>.csv  external predicate tables
            (or dbRoot/<DBName>/ext/ if DBName is set)
```

External predicates extend CodeQL's reach into compiler-level information
(escape analysis, allocation sites, etc). The `extgenScript` field is optional
on each `BuildGroup`; if empty or omitted, extgen is skipped for that group.
The `goescape` scriptlet captures the Go compiler's `-m=2` diagnostics and
converts them via `escape_adapter`. Custom scripts can produce arbitrary CSV
tables.

The `-dir` parameter (or `$DB_EXT_DIR` env var) points to a database's `ext/`
directory, making the external data logically part of that database.

### 4. Query — `codeql_qdriver` (Phase: query)

```
Input:   YAML.QueryGrps[*].QueryDBs          which databases to query
         YAML.QueryGrps[*].Queries           which query scripts to run
         YAML.QueryGrps[*].Externals         external predicate names (optional)
         YAML.QueryGrps[*].ExternalFiles     YAML file listing externals (optional, alternative to Externals)
         qlsrc/<query.ql>                    query script
         dbRoot/<DirBaseName>/ext/<pred>.csv  external predicates
Process: codeql query run                    \
           -d=dbRoot/<DirBaseName>            \
           --search-path=qlsrc/lib            \
           qlsrc/<query.ql>                   \
           --output=resultRoot/<queryPathNoExt>/<DirBaseName>.bqrs  \
           [--external=<pred>=dbRoot/.../ext/<pred>.csv ...]
         ExternalFiles loads a YAML array of predicate names; those names are
         merged into Externals. This avoids repeating the same list in every
         query group config.
Output:  resultRoot/<queryPathNoExt>/<DirBaseName>.bqrs  raw BQRS result
```

Each (database × query) pair runs independently, in parallel. The BQRS format
is CodeQL's binary result format, compact and fast to write.

### 5. Decode — `codeql_qdriver` (Phase: decode)

```
Input:   resultRoot/<queryPathNoExt>/*.bqrs  raw results
         YAML.targetDecodeFmt              "csv" | "json" | "text" | "bqrs"
Process: codeql bqrs decode --format=<fmt>   \
           <bqrsPath>                        \
           --output=<bqrsPathWithoutExt>.<fmt>
Output:  resultRoot/<queryPathNoExt>/<DirBaseName>.<fmt>  decoded results
```

BQRS → human-readable format. Each `.bqrs` file produces one output file in
the same directory. Existing decoded files are overwritten.

### 6. Collect (optional) — `codeql_qdriver`

```
Input:   resultRoot/<queryPathNoExt>/*.csv   per-DB decoded results
Process: merge CSVs, validate headers match, append "repo" column
Output:  resultRoot/<queryPathNoExt>.csv     single combined table
```

Only works when decode format is `csv`. The collection concatenates all CSVs
for the same query, prepending each row with the database name as the `repo`
column. Header mismatch (e.g. query version differences) causes a fatal error.
