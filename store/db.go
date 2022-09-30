package store

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

//go:embed create_tables.sql
var createTable string

var (
	DB_HOST   string
	DB_PORT   string
	DB_USER   string
	DB_NAME   string
	DB_PASSWD string
)

func init() {
	DB_HOST = os.Getenv("DB_HOST")
	DB_PORT = os.Getenv("DB_PORT")
	DB_USER = os.Getenv("DB_USER")
	DB_NAME = os.Getenv("DB_NAME")
	DB_PASSWD = os.Getenv("DB_PASSWD")
}

type DB struct {
	db *sqlx.DB
}

func NewDB() (*DB, error) {
	if DB_HOST == "" || DB_PORT == "" || DB_USER == "" || DB_NAME == "" {
		return nil, errors.New("DB environment variables not set")
	}
	d, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", DB_USER, DB_PASSWD, DB_HOST, DB_PORT, DB_NAME))
	if err != nil {
		return nil, err
	}
	db := &DB{db: d}
	if err := db.init(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) init() error {
	_, err := db.db.Exec(`SET GLOBAL tidb_multi_statement_mode='ON'`)
	if err != nil {
		return err
	}
	initialized, err := db.initialized()
	if err != nil {
		return err
	}
	if !initialized {
		if err := db.createTables(); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) initialized() (bool, error) {
	var tables []string
	if err := db.db.Select(&tables, "SHOW TABLES"); err != nil {
		return false, err
	}
	for _, t := range tables {
		if strings.EqualFold(t, "system") {
			return true, nil
		}
	}
	return false, nil
}

func (db *DB) createTables() error {
	_, err := db.db.Exec(createTable)
	return err
}

func (db *DB) Close() {
	db.db.Close()
}

func (db *DB) Log(level, message string) {
	log.Println(level, message)
	db.db.Exec("INSERT INTO logs (level, message) VALUES (?, ?)", level, message)
}
