package utils

import (
	"fmt"
	"io"
	"os/exec"
)

var Verbose bool = true

func Runcmd(dir string, name string, args ...string) error {
	return RuncmdWithBuf(dir, nil, nil, name, args...)
}

func RuncmdWithBuf(dir string, outbuf, errbuf io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if outbuf != nil {
		cmd.Stdout = outbuf
	}
	if errbuf != nil {
		cmd.Stderr = errbuf
	}
	cmd.Dir = dir
	if Verbose {
		fmt.Println(cmd.String())
	}
	return cmd.Run()
}
