package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	f, err := os.Open("go.mod")
	if err != nil {
		log.Fatal(err)
	}
	var line string
	ptrf := &f
	falias := *ptrf
	ptrf2 := &falias
	fmt.Fscanf(*ptrf2, "%s", &line)
	fmt.Println(line)
	if line == "   " {
		f.Close()
	}
}
