package db

import (
	"database/sql"
	"fmt"
	"log"
)

const (
	host     = "localhost"
	port     = 5432
	userPg   = "lucas"
	password = "teste"
	dbname   = "foo_dev"
)

type Config struct {
	conn *sql.DB
}

func NewConfig(host string, port int, user string, password string, dbName string) *Config {
	connString := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)
	conn, err := sql.Open("postgres", connString)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected")
	return &Config{conn}
}

func (c *Config) Connection() *sql.DB {
	return c.conn
}
