package db

import "log"

func(conn *Connection) TruncateTables() {
	_, err := conn.Exec(`TRUNCATE "Type" RESTART IDENTITY CASCADE`)
	if err != nil {
		log.Println("Fail to truncate table Type:", err)
	}
	_, err = conn.Exec(`TRUNCATE "Function" RESTART IDENTITY CASCADE`)
	if err != nil {
		log.Println("Fail to truncate table Function:", err)
	}
	_, err = conn.Exec(`TRUNCATE "Variable" RESTART IDENTITY CASCADE`)
	if err != nil {
		log.Println("Fail to truncate table Variable:", err)
	}
	_, err = conn.Exec(`TRUNCATE "Statement" RESTART IDENTITY CASCADE`)
	if err != nil {
		log.Println("Fail to truncate table Statement:", err)
	}
	_, err = conn.Exec(`TRUNCATE "Expression" RESTART IDENTITY CASCADE`)
	if err != nil {
		log.Println("Fail to truncate table Expression:", err)
	}
}

func(conn *Connection) TruncateExprMemAcc() {
	_, err := conn.Exec(`TRUNCATE "ExprMemAcc" RESTART IDENTITY CASCADE`)
	if err != nil {
		log.Println("Fail to truncate table ExprMemAcc:", err)
	}
}
