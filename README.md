# QLStat

Analyze real-world project batch with declarative static analysis provided by CodeQL for empirical study and statistical analysis to gain insight of patterns in real world projects.

## Usage

1. Create your `stat.yaml` config file according to [`example.yaml`](./example.yaml).
2. Run `go run ./cmd/batch_clone_build stat.yaml` to clone github repositories and use codeql to create databases for these repositories.
   1. Add `-noclone` option to disable cloning.
3. Create your queries in [`qlsrc`](./qlsrc/).
4. Run `go run ./cmd/codeql_qdriver -collect stat.yaml` to run your queries in former created databases.
   1. The result for each repository will be stored in `<resultRoot>/<path/to/query>/<repo>.csv`.
   2. `-collect` option collects all csv files of different repositories to one csv file with `repo_name` attribute added. You can import the csv to [ClickHouse](https://clickhouse.com/docs/integrations/data-formats/csv-tsv) or other databases for further analysis.


## Storage Structure

```bash
cmd/
   batch_clone_build # clone repositories and build databases for these repositories
   codeql_qdriver    # run queries in codeql database
   escape_adapter    # adapter to convert escape analysis log to csv file for generating external predicates to extend CodeQL ability
qlsrc/   # sources of query
repos/   # repositories
   hostname/   # repositories in hostname, typically github.com/gitlab.com/...
      repo0/
      ...
   test/       # test repositories. Assume the hostname is test
      repo0/
      ...
codeql-db/     # database root
   hostname/
      repo0/   # repo0 database
         ext/  # external predicate databases(csv files)
      ...
   test/
   ${lang}_log.txt
   repoTimes.csv
codeqlResult/  # root for results of queries
   path/to/
      queryName/  # query result for each repository
         repo0.bqrs
         repo0.csv
      queryName.csv  # collected query result for path/to/queryName
logs/          # logs for different stages
   build/      # log for database building
      ${time}/
         hostname/repo/
            out
            err
         repo_build.txt
         repoTimes.csv
   query/      # log for query
      ${time}/path/to/query/
         repo0.err
         repo0.out
   decode/
      ${time}/path/to/query   
         repo0.out
         repo0.err
```

# Citation

```bibtex
@misc{qlstat,
    author = {Qingwei Li},
    title = {QLStat},
    howpublished = {\url{https://github.com/Lslightly/QLStat}},
}
```
