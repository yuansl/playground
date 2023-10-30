// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-09-17 10:15:31

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/yuansl/playground/util"
)

var fatal = util.Fatal

type Filter struct {
}

type Document struct {
	CreateAt time.Time
	UpdateAt time.Time
}

type Mongodb interface {
	Find(ctx context.Context, filter *Filter) (any, error)
	Update(ctx context.Context, doc Document) error
	Delete(ctx context.Context, filter *Filter) error
	Create(ctx context.Context, doc Document) error
}

type Entry struct {
	Key   string
	Value any
}

type Redis interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any) error
	HGet(ctx context.Context, key string) error
	HSet(ctx context.Context, e ...Entry) error
}

type Mysql interface {
	Query(ctx context.Context, sql string) error
	Execute(ctx context.Context, sql string) error
}

type Record struct {
	Name  string `db:"Tables_in_mysql"`
	Value string
}

//go:generate stringer -type Gender -linecomment
type Gender int

const (
	Male   Gender = iota + 1 // male
	Female                   // female
)

func lambda[T any](f func() (T, error)) T {
	r, err := f()
	if err != nil {
		fatal(err)
	}
	return r
}

func main() {
	db := lambda(func() (*sqlx.DB, error) {
		return sqlx.Open("mysql", "yuansl:admin@tcp(localhost:3306)/test?parseTime=true&loc=UTC")
	})

	rows := lambda(func() (*sqlx.Rows, error) {
		return db.NamedQuery("show tables from mysql", map[string]any{})
	})

	for rows.Next() {
		var row Record

		if err := rows.StructScan(&row); err != nil {
			fatal("rows.StructScan error:", err)
		}
		fmt.Printf("%+v\n", row)
	}
	lambda(func() (any, error) {
		return nil, rows.Err()
	})
	ctx := context.TODO()
	{
		tx := lambda(func() (*sqlx.Tx, error) {
			return db.BeginTxx(ctx, &sql.TxOptions{})
		})
		defer tx.Commit()

		var row = struct {
			Age    int    `db:"age"`
			Name   string `db:"name"`
			Gender Gender `db:"gender"`
		}{
			Age:    26,
			Name:   "sibada",
			Gender: Male,
		}

		r, err := tx.NamedExecContext(ctx, "INSERT INTO sufy(name,age,gender) VALUES(:name, :age, :gender)", &row)
		if err != nil {
			fmt.Println("database error:", err)
			tx.Rollback()
			return
		}
		id, err := r.LastInsertId()
		fmt.Printf("Insert id: #%d\n", id)
	}
}
