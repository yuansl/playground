package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/qbox/net-deftones/util"
)

func Open(parquetfile string) *sqlx.DB {
	db, err := sqlx.Open("duckdb", "")
	if err != nil {
		util.Fatal(err)
	}
	_, err = db.Exec("INSTALL parquet; LOAD parquet; ")
	if err != nil {
		util.Fatal("INSTALL parquet: ", err)
	}
	return db
}

func InspectParquetFile(parfile string) error {
	db := Open(parfile)
	var counter atomic.Int32
	start := time.Now()

	rows, err := db.Queryx("select server_ip, domain,request_time,bytes_sent,request_id from read_parquet('" + parfile + "') ;")
	if err != nil {
		util.Fatal("db.Query:", err)
	}
	for rows.Next() {
		var stdlog StandardizedLog

		err = rows.StructScan(&stdlog)
		if err != nil {
			util.Fatal("Rows.Scan:", err)
		}
		counter.Add(+1)

		if stdlog.Domain == "qn-pcdngw.cdn.huya.com" {
			fmt.Printf("Read %+v\n", stdlog)
		}
	}
	rows.Close()
	fmt.Printf("read %d lines from parquest file %s in %v\n", counter.Load(), parfile, time.Since(start))
	return rows.Err()
}
