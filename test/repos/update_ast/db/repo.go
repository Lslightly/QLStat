package db

import "database/sql"

func (conn *Connection) ListRepoitories(only_interested bool) ([]string, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if only_interested {
		rows, err = conn.Query(`select repo_name FROM repository where interested=true ORDER BY repo_name`)
	} else {
		rows, err = conn.Query(`select repo_name FROM repository ORDER BY interested DESC, repo_name`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	repos := make([]string, 0, 1000)
	for rows.Next() {
		var repo string
		err = rows.Scan(&repo)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}
