package accounts

import (
	"database/sql"
	"testing"

	"github.com/watanabe9090/cerberus/internal"
)

func initDB() (*sql.DB, *AccountsRepository) {
	db, _ := internal.OpenPostgreSQLConnection(&internal.DBProps{
		Host:     "localhost",
		User:     "postgres",
		Password: "example",
		Port:     5432,
		DBName:   "cerberus_test",
	})
	repository := NewAccountsRepository(db)
	return db, repository
}

func TestGIVEN_EmptyDB_WHEN_Ok_THEN_InitAccountsTable(t *testing.T) {
	db, repository := initDB()
	defer db.Close()

	err := repository.InitUsersTable()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGIVEN_Params_WHEN_Ok_THEN_Save(t *testing.T) {
	db, repository := initDB()
	db.Exec(`DELETE * FROM accounts`)
	defer db.Close()

	err := repository.Save("foobar", "barfoo", "USER")
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGIVEN_Params_WHEN_Ok_THEN_GetByUsername(t *testing.T) {
	db, repository := initDB()
	db.Exec(`DELETE * FROM accounts`)
	defer db.Close()

	err := repository.Save("foobar", "barfoo", "USER")
	if err != nil {
		t.Error(err.Error())
	}
	acc, err := repository.GetByUsername("foobar")
	if err != nil {
		t.Error(err.Error())
	}
	if acc.ID == 0 {
		t.Error("account not found")
	}
}
