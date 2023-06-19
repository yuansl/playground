// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-04-17 11:41:20

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func main() {
	sess, err := mgo.Dial("mongodb://127.0.0.1:27017")
	if err != nil {
		fatal("mgo.Dial:", err)
	}
	sess.Ping()

	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI("mongodb://localhost:27017").SetConnectTimeout(2*time.Second))
	if err != nil {
		fatal(err)
	}
	fmt.Println("session.Ping:", client.Ping(context.TODO(), nil))
}
