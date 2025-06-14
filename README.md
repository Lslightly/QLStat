## ql使用说明

1. 安装codeql，加入PATH中。
2. 以下步骤在不知道脚本在干什么的情况下均不要使用**默认参数**
   1. 使用[codeql_build dirBuild.sh](http://222.195.92.204:1480/vm/empirical-go/-/blob/master/scripts/codeql_build/dirBuild.sh?ref_type=heads)构建仓库数据库
   2. 使用[codeql qdriver](http://222.195.92.204:1480/vm/empirical-go/-/tree/master/scripts/codeql_qdriver?ref_type=heads)以及编写相应的查询配置文件进行查询
   3. 使用[codeql result parser](http://222.195.92.204:1480/vm/empirical-go/-/tree/master/scripts/codeql_result_parser?ref_type=heads)以及编写相应的分析方式对分析结果进行汇总分析

以`pkgcall`的测试为例：

```bash
cd codeql_build
./dirBuild.sh -repo pkgcall -o pkgcall.csv -l go -db <输出数据库的根目录> <待分析仓库根目录，这里为codeql-queries/test/repos>

cd codeql_qdriver
./qdriver -c test_pkgcall.yaml # 需要修改InDBRoot为上一个脚本的Root，修改QueryRoot和OutResultRoot

cd 
```

## ql说明

用于存储Go语言相关的codeql-queries

- [连续解引用`**p`](../codeql-queries/consecutiveDerefTimes.ql)
- [连续域访问`a.f1.f2.f3`](../codeql-queries/consecutiveFieldAccTimes.ql)
- [解引用操作`*a`, 隐式`(*x).f`](../codeql-queries/derefAcc.ql)
- [域访问表达式`a.f`，不包含函数](../codeql-queries/fieldAcc.ql)
- [赋值语句LHS包含解引用操作`*a, *b = ...`, `(*a).f = ...`](../codeql-queries/lhsDeref.ql)
  - 如果有多赋值的情况，则会算作两个赋值
- [赋值语句LHS不包含解引用操作](../codeql-queries/lhsNoDeref.ql)
- [分析仓库的lines of code](../codeql-queries/loc.ql)
- [分析仓库中没有报错的代码的lines of code](../codeql-queries/locOfFilesAnalyzed.ql)
- [`m[i]`where `m` is map type](../codeql-queries/mapIdx.ql)
- [成员函数对于成员进行更新但无影响](../codeql-queries/nonEffectUpdateToMember.ql)
- [`interface = data` where `data` is not pointer type](../codeql-queries/nonPtrAssignToInterface.ql)
- [指针类型嵌套深度`**int, 2`](../codeql-queries/pointer-type-level.ql)
- [`interface = data` where `data` is pointer type](../codeql-queries/ptrAssignToInterface.ql)
- [`&a`](../codeql-queries/refAcc.ql)
- [`s[i]` where `s` is slice type](../codeql-queries/sliceIdx.ql)
- [变量直接访问`a`](../codeql-queries/variableAcc.ql)
- [变量类型，包括方法接收者、函数形参、`a := init`](../codeql-queries/varType.ql)

codeql官方提供查询语句

- [DatabaseCallInLoop](../codeql-queries/codeql-ql-src/DatabaseCallInLoop.ql)
  - 如果在循环中使用database调用，则可能存在[N+1问题](https://planetscale.com/blog/what-is-n-1-query-problem-and-how-to-solve-it)
- [DivideByZero](../codeql-queries/codeql-ql-src/DivideByZero.ql)
  - [ ] 暂时存在问题，codeql自己提供的存在除零错误的go例子也没有被识别
- [WebCacheDeception](../codeql-queries/CWE-525/WebCacheDeception.ql)
  - 通过修改Cache-Control Header将受害者访问的网站cache到所有用户中。
- [SensitiveConditionBypass](../codeql-queries/CWE-807/SensitiveConditionBypass.ql)
  - 一些条件语句中包含和关键信息比对相关的东西，如密码比较，如果正确，则进入if分支。这种可能可以通过某种方式绕过，存在安全隐患。

## codeql build

使用 [dirBuild.sh](./dirBuild.sh) 构建codeql数据库

## codeql qdriver

通过编写 yaml 配置文件，指定要查询的数据库、查询语句、输出结果等，使用 [qdriver](./cmd/codeql_qdriver/qdriver) 运行查询。

## codeql result parser

使用 [qParser](./cmd/codeql_result_parser/qParser) 解析查询结果，并收集整合到csv文件中。

## codeql rename query

使用 [codeql_rename](./cmd/codeql_rename_query/codeql_rename) 对查询语句进行重命名，同时重命名结果目录。
