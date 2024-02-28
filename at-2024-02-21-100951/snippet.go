// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2024-02-21 10:09:51

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	sql "github.com/jmoiron/sqlx"

	"github.com/yuansl/playground/util"
)

func main() {
	db, err := sql.Open("mysql", "traffic_admin:traffic@tcp(localhost:3306)/traffic?parseTime=true&Loc=UTC")
	if err != nil {
		util.Fatal(err)
	}
	_ = db
	sqlbuf := strings.Builder{}
	uids := []int{33, 1, 34, 35}
	query, args, err := sql.In("uid IN (?)", uids)
	if err != nil {
		util.Fatal(err)
	}
	fmt.Print("query:", query, " args:", args)
	query = db.Rebind(query)
	fmt.Fprintf(&sqlbuf, " AND %s", query)
	fmt.Print("\nsql:", sqlbuf.String())
}
