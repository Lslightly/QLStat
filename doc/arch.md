# Project Architecture

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
