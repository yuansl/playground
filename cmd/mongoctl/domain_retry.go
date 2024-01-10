package main

import (
	"context"
	"fmt"

	"github.com/qbox/net-deftones/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yuansl/playground/util"
)

// DomainRetry represents mongo collection 'log-validation.domain_retry'
type DomainRetry struct {
	*mongo.Collection
}

// DeleteDomainRetry implements Mongodb.
func (db *DomainRetry) Delete(ctx context.Context, filter *DomainRetryFilter) error {
	result, err := db.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("%w: db.%s.DeleteMany: %v", ErrDatabase, db.Name(), err)
	}
	if result.DeletedCount <= 0 {
		return fmt.Errorf("db.%s.DeleleMany deletes nothing: %+v", db.Name(), result)
	}
	logger.FromContext(ctx).Infof("db.%s.deleteMany(%+v) result: %+v\n", db.Name(), filter, result)
	return nil
}

// FindDomainRetry implements Mongodb.
func (db *DomainRetry) Find(ctx context.Context, filter *DomainRetryFilter) ([]DomainRetryInfo, error) {
	assert(filter != nil)

	var result []DomainRetryInfo

	it, err := db.Collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%w: db.find: %v", ErrDatabase, err)
	}
	if err = it.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("%w: cursor.All: %v", ErrDatabase, err)
	}
	return result, nil
}

func newMongodb(mongouri string, db, collection string) (*mongo.Collection, error) {
	if db == "" || collection == "" || mongouri == "" {
		return nil, fmt.Errorf("%w: neither 'mongouri', 'db' nor 'collection' can be empty", ErrInvalid)
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongouri))
	if err != nil {
		return nil, fmt.Errorf("%w: mongo.Connect: %v", ErrDatabase, err)
	}
	return client.Database(db).Collection(collection), nil
}

func NewDomainRetry(mongouri string, db, collection string) *DomainRetry {
	coll, err := newMongodb(mongouri, db, collection)
	if err != nil {
		util.Fatal("NewDomainRetry: newMongodb: ", err)
	}
	return &DomainRetry{Collection: coll}
}
