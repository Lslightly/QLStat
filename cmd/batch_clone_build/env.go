package main

const (
	REPO_DIR   = "REPO_DIR"   // the root directory of the repository
	OUTPUT_DIR = "OUTPUT_DIR" // the directory to store intermediate results for generating external predicate
	PROJROOT   = "PROJROOT"   // the root directory of the project
	DB_EXT_DIR = "DB_EXT_DIR" // the directory to store external predicate database
)

type envpair struct {
	name, value string
}

// genEnv converts envpairs to strings in the format of "name=value"
func genEnv(pairs []envpair) (res []string) {
	for _, pair := range pairs {
		res = append(res, pair.name+"="+pair.value)
	}
	return res
}
