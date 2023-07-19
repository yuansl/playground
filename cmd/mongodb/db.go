package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Iterator interface {
	Next() (any, error)
}

type Mongodb interface {
	Find(ctx context.Context, selector any) (Iterator, error)
	Update(ctx context.Context, selector any) error
	Insert(ctx context.Context, selector any) error
	Delete(ctx context.Context, selector any) error
}

type mongodb struct {
	db *mongo.Collection
}

func (db *mongodb) Find(ctx context.Context, selector any) (Iterator, error) {
	panic("not implemented") // TODO: Implement
}

func (db *mongodb) Update(ctx context.Context, selector any) error {
	panic("not implemented") // TODO: Implement
}

func (db *mongodb) Insert(ctx context.Context, selector any) error {
	panic("not implemented") // TODO: Implement
}

func (db *mongodb) Delete(ctx context.Context, selector any) error {
	panic("not implemented") // TODO: Implement
}

func fatal(format string, v ...any) {
	fmt.Fprintf(os.Stderr, "fatal error: "+format, v...)
	os.Exit(1)
}

func New(mongouri, db, coll string) Mongodb {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongouri))
	if err != nil {
		fatal("mongo.Connect error: %v\n", err)
	}
	return &mongodb{db: client.Database(db).Collection(coll)}
}

func main() {
	New("monodb://uri", "deftone_monitor_v3", "user_defined_rule")
	redis.NewClient(nil)
}
