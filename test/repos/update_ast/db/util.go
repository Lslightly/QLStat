package db

import (
	"database/sql"
	"log"

	"github.com/hashicorp/golang-lru/v2"
)

type InvalidName struct{}
func(_ InvalidName) Error() string {
	return "InvalidName: _"
}

type NotFoundError struct{}
func(_ NotFoundError) Error() string {
	return "not found"
}

type ChildMissingError struct{}
func(_ ChildMissingError) Error() string {
	return "child mising"
}

const TYPE_CACHE_SIZE = 65536

func new_cache[Key comparable, Value any]() *lru.Cache[Key,Value] {
	cache, err := lru.New[Key,Value](TYPE_CACHE_SIZE)
	if err != nil {
		log.Fatal(err)
	}
	return cache
}

type TypeIsNil struct{}
func(_ TypeIsNil) Error()string {
	return "type is nil"
}

func to_null_int64[Int ~int|~int64](i Int) sql.NullInt64 {
	if i==-1 {
		return sql.NullInt64{Valid: false}
	} else {
		return sql.NullInt64{Valid: true, Int64: int64(i)}
	}
}

func to_null_str(s string) sql.NullString {
	if s=="" || s=="_" {
		return sql.NullString{Valid: false}
	} else {
		return sql.NullString{String: s, Valid: true}
	}
}
