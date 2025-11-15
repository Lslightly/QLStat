package main

import "log"

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func bypass[T any](v T, err error) T {
	fatal(err)
	return v
}
