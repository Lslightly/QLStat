# 查询驱动

通过指定yaml文件执行查询语句。

```bash
$ ./qdriver --help 
  -c string
        the configuration file (default "./go.yaml")
  -format string
        target format of decode, including text, csv, json, bqrs (default "csv")
  -only-decode
        only decoding bqrs files
```

yaml模板：

```yaml
%YAML 1.2
---
InDBRoot: /data/github_go/codeql-db
Repos: []
QueryRoot: ../../codeql-queries
Queries:
  # - consecutiveDerefTimes.ql
  # - consecutiveFieldAccTimes.ql
  - varType.ql
OutResultRoot: /data/github_go/codeqlResult
ErrorLog: error.log
```

说明：
1. InDBRoot 输入数据库的根目录
2. Repos 指定需要分析的仓库，如果为空，则分析InDBRoot下的所有仓库
3. QueryRoot 指定query的根目录
4. Queries 指定要分析的query
5. OutResultRoot 指定结果输出目录，结果目录下先以query脚本的名称创建目录，目录中包含：
   1. `repo.bqrs`, `repo.csv`, `repo.json`，均为`.bqrs`结果文件的一种解释形式
   2. `analyze`目录，用于存放聚合之后的结果，聚合分析见[../codeql_result_parser](../codeql_result_parser)
   3. `log`目录，用于存放查询时的日志

# 测试

使用[test.sh](./test.sh)可以根据[test.yaml](./test.yaml)配置文件进行相应查询。将查询结果按照ql名称放在[../../codeql-queries/test/codeqlResult/](../../codeql-queries/test/codeqlResult/)目录下。

# yaml文件说明

- [egSliceExprNotSetNil.yaml](./egSliceExprNotSetNil.yaml)用于说明WEF-sliceExpr这个问题模式在aws-sdk-go-v2中的表现情况
- [go.yaml](./go.yaml)用于检测不同query语句的查询结果
- [newHaveRepo.yaml](./newHaveRepo.yaml)用于检测996个仓库的数据
- [prototype.yaml](./prototype.yaml)用于分析为软件原型大赛所准备的仓库数据库
- [sliceExprNotSetNil.yaml](./sliceExprNotSetNil.yaml)用于在996个仓库的数据库中分析WEF-sliceExprNotSet的情况
- [test.yaml](./test.yaml)用于测试[codeql-queries](../../codeql-queries/)中的仓库和query语句

