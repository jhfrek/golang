package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "Simbirka"
	dbname   = "tracker_db"
)

func main() {

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	defer db.Close()

	err = db.Ping()
	CheckError(err)

	fmt.Println("Соединение установлено!")

	insertStmt := `TRUNCATE "Employee"`
	_, err = db.Exec(insertStmt)
	CheckError(err)

	insertStmt = `insert into "Employee"("Name", "EmpId") values('Rohit', 21)`
	_, err = db.Exec(insertStmt)
	CheckError(err)

	insertStmt = `insert into "Employee"("Name", "EmpId") values($1, $2)`
	_, ee := db.Exec(insertStmt, "krish", 10)
	CheckError(ee)

	fmt.Println("Ура!!! Заработало!!!")
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
