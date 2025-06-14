# codeql查询结果分析

通过[qlSumConfig.yaml](qlSumConfig.yaml)指定对每个类型的ql查询结果的统计collect方式。目前实现：
- 条目数量统计
- group by数量统计，需指定列数

- [ ] 后续可引入csv上的QL查询(但是目前没有这个需求)

```bash
$ ./qParser --help        
accept the root directory path of codeql result as the last argument
  -c string
        the analyzer yaml configuration file (default "qlSumConfig.yaml")
```

# 测试

使用[test.sh](./test.sh)，根据[test.yaml](./test.yaml)统计分析[../../codeql-queries/test/codeqlResult/](../../codeql-queries/test/codeqlResult/)中的结果，结果在[../../codeql-queries/test/codeqlResult/`qlname`/analyze/](../../codeql-queries/test/codeqlResult/consecutiveDerefTimes/analyze/)中。

