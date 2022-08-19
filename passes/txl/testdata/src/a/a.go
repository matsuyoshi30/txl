package a

import (
	"database/sql"
	"fmt"
	"log"
)

func testcase1() {
	var db *sql.DB

	tx, err := db.Begin() // want "transaction variable declared here is used 3 times until COMMIT"
	if err != nil {
		log.Fatal(err)
	}

	tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction", "tx",10)`) // want "INSERT album"

	_, err = tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction", "tx",10)`) // want "INSERT album"
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println(err)
		}
	}

	if _, err = tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction", "tx",10)`); err != nil { // want "INSERT album"
		if err := tx.Rollback(); err != nil {
			log.Println(err)
		}
	}

	// TODO
	// runTx := func(tx *sql.Tx, n int) (sql.Result, error) {
	// 	_, err := tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction", "tx", 10)`)
	// 	return nil, err
	// }

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

type S struct{ n int }

func (s *S) Exec() int {
	return s.n
}

type T struct{ n int }

func (t *T) Exec(s string) (sql.Result, error) {
	return nil, nil
}

func testcase2() {
	var db *sql.DB

	tx, err := db.Begin() // want "transaction variable declared here is used 2 times until COMMIT"
	if err != nil {
		log.Fatal(err)
	}

	tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction", "tx",10)`) // want "INSERT album"

	{
		tx := &T{1}
		tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction", "tx",10)`)
	}

	_, err = tx.Exec(`INSERT INTO album (title, artist, price) VALUES ("transaction", "tx",10)`) // want "INSERT album"
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

type Album struct {
	ID     int64
	Title  string
	Artist string
	Price  float32
}

func testcase3() {
	var db *sql.DB

	tx, err := db.Begin() // want "transaction variable declared here is used 2 times until COMMIT"
	if err != nil {
		log.Fatal(err)
	}

	_, err = addAlbumTx(tx, Album{ // want "INSERT album"
		Title:  "transaction1",
		Artist: "tx1",
		Price:  59.99,
	})
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println(err)
		}
	}

	if _, err = addAlbumTx(tx, Album{ // want "INSERT album"
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

type txRunner struct {
	tx *sql.Tx
}

func newTxRunner(tx *sql.Tx) *txRunner { return &txRunner{tx} }

func (tr *txRunner) f(title, artist string, price float64) (sql.Result, error) {
	return tr.tx.Exec(`INSERT INTO album (title, artist, price) VALUES (?, ?, ?)`, title, artist, price)
}

// TODO
// func testcase4() {
// 	var db *sql.DB

// 	tx, err := db.Begin() // "transaction variable declared here is used 2 times until COMMIT"
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	tr := newTxRunner(tx)

// 	_, err = tr.f("transaction", "tx", 10.0) // "INSERT album"
// 	if err != nil {
// 		if err := tx.Rollback(); err != nil {
// 			log.Println(err)
// 		}
// 	}

// 	_, err = tr.f("transaction", "tx", 10.0) // "INSERT album"
// 	if err != nil {
// 		if err := tx.Rollback(); err != nil {
// 			log.Println(err)
// 		}
// 	}

// 	if err := tx.Commit(); err != nil {
// 		log.Println(err)
// 	}
// }
