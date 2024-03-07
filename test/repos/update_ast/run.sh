./astdb -interest -trunc
./astdb -interest -def -lightWeightAnalyze=false
mv update_ast.log update_ast-def.log
./astdb -interest -mem  -lightWeightAnalyze=false
mv update_ast.log update_ast-mem.log
