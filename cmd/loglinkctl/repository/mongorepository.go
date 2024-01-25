package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/qbox/net-deftones/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const _LOGLINK_DB = "fusionlogv2"
const _LOGLINK_COLLECTION = "logv2"
const _LOG_TIME_FORMAT = "2006-01-02-15"

type fusionlogRepository struct {
	*mongo.Collection
}

// GetLink implements LinkRepository.
func (db *fusionlogRepository) GetLinks(ctx context.Context, domain string, begin time.Time, end time.Time, business Business) ([]LogLink, error) {
	var links []LogLink
	query := bson.M{
		"hour":     bson.M{"$gte": begin.Format(_LOG_TIME_FORMAT), "$lt": end.Format(_LOG_TIME_FORMAT)},
		"business": business.String(),
	}
	if domain != "" {
		query["domain"] = domain
	}
	it, err := db.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("db.Find: %v", err)
	}
	defer it.Close(ctx)

	err = it.All(ctx, &links)

	return links, err
}

// SetDownloadUrl implements LinkRepository.
func (db *fusionlogRepository) SetDownloadUrl(ctx context.Context, link *LogLink, url string) error {
	logger.FromContext(ctx).Infof("link: %+v, new url: %q\n", link, url)
	_id, err := primitive.ObjectIDFromHex(link.Id)
	_, err = db.UpdateByID(ctx, _id, bson.M{
		"$set": bson.M{
			"url":   url,
			"mtime": time.Now().Unix(),
		}},
	)
	return err
}

func NewMongoLinkRepository(monguri string) (LinkRepository, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(monguri))
	if err != nil {
		return nil, fmt.Errorf("mongo.Connect: %v", err)
	}
	return &fusionlogRepository{
		Collection: client.Database(_LOGLINK_DB).Collection(_LOGLINK_COLLECTION),
	}, nil
}
