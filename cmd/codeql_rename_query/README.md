# 重命名codeql-query

当重命名query时，除了重命名codeql query的名字之外，还需要改变codeqlResult中的名称。

如果目标名字已经存在，则需要手动确认是否覆盖。

```yaml
%YAML 1.2
---
QueryRoot: ../../codeql-queries
Queries:
  - old: type-depth
    new: pointer-type-level
ResultRoots:
  - /data/github_go/codeqlResult
  - ../../codeql-queries/test/codeqlResult
```

- QueryRoot query根目录
- Queries 要重命名的queries
  - old 旧名称
  - new 新名称
- ResultRoots 要重命名的结果根目录
