package dbrepository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"playground/at-2023-06-05-232009/types"
)

type mongodb struct {
	*mongo.Collection
}

func (db *mongodb) Save(ctx context.Context, domains []types.Domain) (_ error) {
	for _, domain := range domains {
		_, err := db.UpdateOne(context.TODO(),
			bson.M{"domain": domain.Domain},
			bson.M{"$set": bson.M{
				"updateAt": time.Now(),
				"cname":    "www.shifen.com.",
			}},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			return fmt.Errorf("mongodb.UpdateMany: %v", err)
		}
	}
	return nil
}

type domainDoc struct {
	Cname string `bson:"cname"`
}

func (db *mongodb) Find(ctx context.Context, domain string, opts ...any) (domains []types.Domain, err error) {
	it, err := db.Collection.Find(ctx, bson.M{"domain": domain})
	if err != nil {
		return nil, fmt.Errorf("mongodb.find: %v", err)
	}
	defer it.Close(ctx)

	for it.Next(ctx) {
		doc := domainDoc{}

		if err = it.Decode(&doc); err != nil {
			return nil, fmt.Errorf("mongodb.Decode: %v", err)
		}
		domains = append(domains, types.Domain{Cname: doc.Cname})
	}
	return domains, nil
}
