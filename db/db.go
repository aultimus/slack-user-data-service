package db

import "github.com/jmoiron/sqlx"

func NewPostgres(dbConn *sqlx.DB) *Postgres {
	return &Postgres{dbConn: dbConn}
}

// Postgres implements the Storer interface
type Postgres struct {
	dbConn *sqlx.DB
}

func (p *Postgres) CreateUser() {

}

func (p *Postgres) UpdateUser() {

}

func (p *Postgres) GetUsers() {

}
