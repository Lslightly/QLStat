# QLStat

Analyze real-world project batch with declarative static analysis provided by CodeQL for empirical study and statistical analysis to gain insight of patterns in real world projects.

## Usage

1. Create your `stat.yaml` config file according to [`example.yaml`](./example.yaml).
2. Run `go run ./cmd/batch_clone_build stat.yaml` to clone github repositories and use codeql to create databases for these repositories.
   1. Add `-noclone` option to disable cloning.
3. Create your queries in [`qlsrc`](./qlsrc/).
4. Run `go run ./cmd/codeql_qdriver -c stat.yaml -collect` to run your queries in former created databases.
   1. The result for each repository will be stored in `<resultRoot>/<path/to/query>/<repo>.csv`.
   2. `-collect` option collects all csv files of different repositories to one csv file with `repo_name` attribute added. You can import the csv to [ClickHouse](https://clickhouse.com/docs/integrations/data-formats/csv-tsv) or other databases for further analysis.

# Citation

```bibtex
@misc{qlstat,
    author = {Qingwei Li},
    title = {QLStat},
    howpublished = {\url{https://github.com/Lslightly/QLStat}},
}
```
