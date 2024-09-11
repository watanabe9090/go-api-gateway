package tokens

import (
	"database/sql"
	"log"
	"time"
)

type TokensRepository struct {
	db *sql.DB
}

func NewTokensRepository(db *sql.DB) *TokensRepository {
	return &TokensRepository{db}
}

func (r *TokensRepository) InitTokenTable() error {
	const query = `
		CREATE TABLE IF NOT EXISTS tokens (
		id SERIAL PRIMARY KEY,
		account_username VARCHAR NOT NULL,
		type VARCHAR(2) NOT NULL,
		token VARCHAR(1024) NOT NULL,
		state VARCHAR(16) NOT NULL,
		created_at BIGINT NOT NULL,
		CONSTRAINT fk_account_username
			FOREIGN KEY(account_username ) REFERENCES accounts(username)
		);
	`
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

func (r *TokensRepository) GetByAccountUsername(username string) ([]Token, error) {
	const query = `
		SELECT id, account_username, type, token, state, created_at 
		FROM tokens 
		WHERE account_username = $1
		ORDER BY created_at DESC;
	`
	var tokens []Token

	rows, err := r.db.Query(query, username)
	for rows.Next() {
		var token Token
		rows.Scan(&token.ID, &token.AccountUsername, &token.Type, &token.Token, &token.State, &token.CreatedAt)
		tokens = append(tokens, token)
	}
	if err != nil {
		log.Println(err.Error())
		return tokens, err
	}
	return tokens, nil
}

func (r *TokensRepository) GetByToken(token string) ([]Token, error) {
	const query = `
		SELECT id, account_username, type, token, state, created_at 
		FROM tokens 
		WHERE token = $1
		ORDER BY created_at DESC;
	`
	var tokens []Token

	rows, err := r.db.Query(query, token)
	for rows.Next() {
		var token Token
		rows.Scan(&token.ID, &token.AccountUsername, &token.Type, &token.Token, &token.State, &token.CreatedAt)
		tokens = append(tokens, token)
	}
	if err != nil {
		log.Println(err.Error())
		return tokens, err
	}
	return tokens, nil
}

func (r *TokensRepository) SaveToken(username string, token string) error {
	const query = `
		INSERT INTO tokens (account_username, type, token, state, created_at) 
		VALUES ($1, $2, $3, $4, $5);
	`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	_, err = stmt.Exec(username, "TK", token, "ACTIVE", time.Now().UnixMilli())
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (r *TokensRepository) UpdateTokensState(username string, token string, state string) error {
	const query = `
		UPDATE tokens
		SET state = $1
		WHERE account_username = $2
		AND token = $3;
	`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	_, err = stmt.Exec(state, username, token)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}
