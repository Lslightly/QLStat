#!/bin/bash
go list -export -f '{{if .Export}}packagefile {{.ImportPath}}={{.Export}}{{end}}' std > importcfg
go tool compile -d=panic -p=p -C -e -importcfg=./importcfg -o a.o -l -d=ssa/check/on -m=2 escape2.go &> m2.log
