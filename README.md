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
