package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Album struct {
	ID     int64
	Title  string
	Artist string
	Price  float32
}

type S struct{ n int }

func (s *S) Exec() int {
	return s.n
}

func main() {
	// Capture connection properties.
	cfg := mysql.Config{
		User:   "root",
		Passwd: "root",
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "recordings",
	}
	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// expr-stmt
	tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction0", "tx0", 49.99)`)

	// assign
	_, err = tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction0", "tx0", 49.99)`)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println(err)
		}
	}

	s := &S{1}
	s.Exec()

	// if
	if _, err = tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction0", "tx0", 49.99)`); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println(err)
		}
	}

	_, err = addAlbumTx(tx, Album{
		Title:  "transaction1",
		Artist: "tx1",
		Price:  59.99,
	})
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println(err)
		}
	}

	if _, err := addAlbumTx(tx, Album{
		Title:  "transaction2",
		Artist: "tx2",
		Price:  69.99,
	}); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

func addAlbumTx(tx *sql.Tx, alb Album) (int64, error) {
	result, err := tx.Exec("INSERT INTO album (title, artist, price) VALUES (?, ?, ?)", alb.Title, alb.Artist, alb.Price)
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	return id, nil
}
