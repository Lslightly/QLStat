package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/goccy/go-yaml"

	fscopy "github.com/otiai10/copy"
)

type RenamePair struct {
	OldName string `yaml:"old"`
	NewName string `yaml:"new"`
}

type ConfigTy struct {
	QueryRoot   string       `yaml:"QueryRoot"`
	Queries     []RenamePair `yaml:"Queries"`
	ResultRoots []string     `yaml:"ResultRoots"`
}

var configPath string

func init() {
	flag.StringVar(&configPath, "c", "./go.yaml", "specify the configuration file")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "NOTICE: this tool is not well tested.\nrename query. The result in ResultRoots will also be renamed for consistency.")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	config := parseConfig()
	for _, pair := range config.Queries {
		oldName, newName := pair.OldName, pair.NewName

		qlExt := ".ql"
		// rename Query
		rename(path.Join(config.QueryRoot, oldName)+qlExt, path.Join(config.QueryRoot, newName)+qlExt)

		// rename results
		for _, root := range config.ResultRoots {
			rename(path.Join(root, oldName), path.Join(root, newName))
		}
	}
}

func parseConfig() (config ConfigTy) {
	bs, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.Unmarshal(bs, &config)
	if err != nil {
		log.Fatalln(err)
	}
	return
}

func rename(oldPath string, newPath string) {
	fmt.Println("renaming", oldPath, "->", newPath)
	if _, err := os.Stat(newPath); err == nil {
		var overwrite int = 2
		for overwrite != 1 && overwrite != 0 {
			fmt.Print(newPath, "already exists, please input 1 to overwrite it or 0 to do nothing and continue: ")
			fmt.Scanf("%d", &overwrite)
		}
		if overwrite == 0 {
			fmt.Println("do nothing for", newPath)
			return
		} else {
			backupPath := newPath + "_bak"
			err := fscopy.Copy(newPath, backupPath)
			if err != nil {
				fmt.Println("error occurs when backup", newPath)
				return
			}
			err = os.RemoveAll(newPath)
			if err != nil {
				fmt.Println("error occurs when removing newPath", newPath, ", backup path is", backupPath)
				return
			}
			err = os.Rename(oldPath, newPath)
			if err != nil {
				fmt.Println("error occurs when renaming", oldPath, ", backup of newPath is", backupPath)
				return
			}
			err = os.RemoveAll(backupPath)
			if err != nil {
				fmt.Println("error occurs when removing backupPath", backupPath)
				return
			}
		}
	} else if os.IsNotExist(err) {
		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Println("error occurs when renaming", oldPath)
			return
		}
	}
}
