package internal

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func OpenPostgreSQLConnection(props *DBProps) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", props.Host, props.Port, props.User, props.Password, props.DBName))
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println(err.Error())
		db.Close()
		return nil, err
	}
	return db, nil
}
