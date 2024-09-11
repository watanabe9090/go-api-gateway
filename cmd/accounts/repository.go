package accounts

import (
	"database/sql"
	"log"
	"time"
)

type AccountsRepository struct {
	db *sql.DB
}

func NewAccountsRepository(db *sql.DB) *AccountsRepository {
	return &AccountsRepository{db}
}

func (r *AccountsRepository) InitUsersTable() error {
	const query = `CREATE TABLE IF NOT EXISTS accounts (
		id SERIAL PRIMARY KEY, 
		username VARCHAR(256) NOT NULL UNIQUE, 
		role VARCHAR(32) NOT NULL,
		password VARCHAR(256) NOT NULL, 
		created_at BIGINT NOT NULL
	);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		log.Println(err)
		return err
	}
	stmt.Close()
	return nil
}

func (r *AccountsRepository) GetByUsername(username string) (*Account, error) {
	const query = `
		SELECT id, username, password, role, created_at 
		FROM accounts WHERE username = $1;
	`
	var user Account
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return &user, nil
}

func (r *AccountsRepository) Save(username string, password string, role string) error {
	const query = `
		INSERT INTO accounts (username, password, role, created_at) 
		VALUES ($1, $2, $3, $4);
	`
	tx, err := r.db.Begin()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	_, err = stmt.Exec(username, password, role, time.Now().UnixMilli())
	if err != nil {
		log.Println(err.Error())
		return err
	}
	stmt.Close()
	tx.Commit()
	return nil
}
