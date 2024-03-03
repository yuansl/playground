package fusionlog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/qbox/net-deftones/util"
)

const _LOGLINK_DB = "fusionlogv2"
const _LOGLINK_COLLECTION = "logv2"
const _LOGV2_TIME_FORMAT = "2006-01-02-15"

var ErrUnavailable = errors.New("fusionlog.db: service unavaiable")

type fusionlogv2 struct {
	coll *mongo.Collection
}

type linkdoc struct {
	LogLink   `bson:",inline"`
	Timestamp int64 `bson:"timeStamp"`
}

func (db *fusionlogv2) Query(ctx context.Context, filter *Filter) ([]LogLink, error) {
	var links []LogLink
	var query = make(bson.M)

	if filter.ID != "" {
		_id, _ := primitive.ObjectIDFromHex(filter.ID)
		query["_id"] = _id
	}
	if filter.Domain != "" {
		query["domain"] = filter.Domain
	}
	var hour bson.M
	if !filter.HourBegin.IsZero() && !filter.HourEnd.IsZero() {
		hour = bson.M{"$gte": filter.HourBegin.Format(_LOGV2_TIME_FORMAT), "$lt": filter.HourEnd.Format(_LOGV2_TIME_FORMAT)}
	} else if !filter.HourBegin.IsZero() {
		hour = bson.M{"$eq": filter.HourBegin.Format(_LOGV2_TIME_FORMAT)}
	} else if !filter.HourBegin.IsZero() {
		hour = bson.M{"eq": filter.HourEnd.Format(_LOGV2_TIME_FORMAT)}
	}
	if len(hour) > 0 {
		query["hour"] = hour
	}
	opts := []*options.FindOptions{options.Find().SetHint("domain_1_hour_1").SetNoCursorTimeout(true)}
	if filter.Limit > 0 {
		opts = append(opts, options.Find().SetLimit(int64(filter.Limit)))
	}
	it, err := db.coll.Find(ctx, query, opts...)
	if err != nil {
		switch {
		case mongo.IsTimeout(err):
			return nil, fmt.Errorf("%w: mongodb.coll.find({%v}): %w", context.DeadlineExceeded, query, err)
		case mongo.IsNetworkError(err):
			fallthrough
		default:
			return nil, fmt.Errorf("%w: db.coll.Find({%v}): %w", ErrUnavailable, query, err)
		}
	}
	defer it.Close(ctx)

	for it.Next(ctx) {
		var doc linkdoc
		if err = it.Decode(&doc); err != nil {
			return nil, fmt.Errorf("%w: db.Decode: %w", ErrUnavailable, err)
		}
		doc.LogLink.Timestamp = time.Unix(doc.Timestamp, 0)
		links = append(links, doc.LogLink)
	}
	return links, nil
}

func (db *fusionlogv2) Update(ctx context.Context, link *LogLink) error {
	_id, err := primitive.ObjectIDFromHex(link.Id)
	if err != nil {
		return fmt.Errorf("%w: primitive.ObjectIDFromHex: %w", ErrInvalid, err)
	}
	_, err = db.coll.UpdateByID(ctx,
		_id,
		bson.M{
			"$set": bson.M{
				"url":   link.Url,
				"mtime": time.Now().Unix(),
			},
		})
	return err
}

func (db *fusionlogv2) Delete(ctx context.Context, link *LogLink) error {
	_id, err := primitive.ObjectIDFromHex(link.Id)
	if err != nil {
		return fmt.Errorf("%w: primitive.ObjectIDFromHex: %w", ErrInvalid, err)
	}
	_, err = db.coll.DeleteOne(ctx, bson.M{"_id": _id})
	return err
}

func NewFusionlogdb(mongouri string) *fusionlogv2 {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongouri).SetRetryReads(true).SetReadPreference(readpref.SecondaryPreferred()))
	if err != nil {
		util.Fatal("mongo.Connect: %v", err)
	}
	return &fusionlogv2{
		coll: client.Database(_LOGLINK_DB).Collection(_LOGLINK_COLLECTION),
	}
}
