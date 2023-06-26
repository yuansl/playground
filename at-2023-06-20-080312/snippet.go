// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-20 08:03:12

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
	"math/rand"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

type Filter struct {
	Time   time.Time
	Domain string
	Idc    string
	Area   string
}

type TimePoint struct {
	Time  time.Time
	Value int64
}

type offdown struct {
	Domain     string
	Uid        uint32
	Hub        string
	Time       time.Time
	Idc        string
	Area       string
	Timeseries []TimePoint
}
type Object any
type Doc struct {
	Object
	UpdatedAt time.Time
}

func dbUpdate(ctx context.Context, db *mongo.Collection, filter *Filter, doc any) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := db.UpdateOne(ctx, filter,
		bson.M{
			"$set": doc,
			"$setOnInsert": bson.M{
				"CreatedAt": time.Now(),
			},
			"$currentDate": bson.M{"UpdatedAt": bson.M{"$type": "date"}},
		}, options.Update().SetUpsert(true))
	return err
}

func main() {
	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fatal("mongo.Connect:", err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		fatal("client.Ping:", err)
	}

	db := client.Database("pili").Collection("off_down_5min")

	date := time.Date(2023, 6, 20, 0, 0, 0, 0, time.Local)
	err = dbUpdate(ctx, db,
		&Filter{Domain: "a.com", Idc: "bc", Area: "apac", Time: date},
		offdown{Hub: "live360", Uid: 999999, Time: date, Domain: "a.com", Idc: "bc", Area: "apac", Timeseries: []TimePoint{
			{Time: time.Now(), Value: rand.Int63()},
			{Time: time.Now().Add(-10 * time.Second), Value: rand.Int63()},
		}})
	if err != nil {
		fatal("update:", err)
	}
}
