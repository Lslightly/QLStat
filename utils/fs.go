package utils

import (
	"log"
	"os"
	"path/filepath"
)

func CreateOutAndErr(sharedPathNoExt string) (outFile, errFile *os.File) {
	outFile = CreateFile(sharedPathNoExt + ".out")
	errFile = CreateFile(sharedPathNoExt + ".err")
	return
}

func MkdirAll(dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Panicf("error when creating dir %s: %v", dir, err)
	}
}

func CreateFile(file string) *os.File {
	dirname := filepath.Dir(file)
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		MkdirAll(dirname)
	}
	res, err := os.Create(file)
	if err != nil {
		log.Panicf("error when creating file %s: %v", res.Name(), err)
	}
	return res
}

func Remkdir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Panicf("error occurs when deleting dir %s: %v", dir, err)
	}
	MkdirAll(dir)
}

func OpenFile(f string) *os.File {
	res, err := os.Open(f)
	if err != nil {
		log.Panicf("error occurs when open file %s: %v", res.Name(), err)
	}
	return res
}
