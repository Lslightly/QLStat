### 使用方法

用 golang.org/x/tools/go/packages to parse packages，更新 Type、Function 和 Variable 表。

警告：运行本程序将清空表 Type、Function 和 Variable。

用法：

与 updata_pkg_file 的用法类似，先更新 config.json 文件，然后执行：
    go build
    ./astdb

先小规模试验。看数据情况，再微调表结构。

### 轻量级make capacity统计

见[LightWeightTraverseMake](analyzer/makeAndNew.go#L35)，会简单统计makeslice, makemap, makechan中的类型，分配的capacity数字（若未知或不可解析，则为-1）。统计命令`go run main.go -make=true`。

该选项只会更改`Make`表格。

### 统计make和new分配对象的大小

见[TraverseMakesliceAndNew](analyzer/makeAndNew.go#L115)，会通过类型检查知道类型大小，从而计算分配对象的大小（若未知或者cap不可解析，则为-1）。统计命令为`go run main.go -MakesliceAndNew=true`。

该选项只会更改`MakesliceAndNew`表格。

### -mem选项会修改的表

- Expression
- ExprIdent
- ExprMemAcc
- ExprTypeAssert
- ExprCall
- Statement
- StmtAssignLhs
- InsertStmtRhs
- StmtRetRes
- StmtTypeSwitch
- StmtTypeSwitchCase
- StmtDefer

