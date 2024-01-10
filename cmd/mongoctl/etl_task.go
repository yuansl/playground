package main

import (
	"context"
	"fmt"
	"time"

	"github.com/qbox/net-deftones/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/yuansl/playground/util"
)

type EtlTask struct {
	Cdn    string
	Domain string
	Hour   string
}

type LogEtlRepository interface {
	FindEtlTasks(ctx context.Context, begin, end time.Time) ([]EtlTask, error)
	DeleteEtlTasks(ctx context.Context, begin, end time.Time) error
}

type Taskdb struct{ *mongo.Collection }

type TaskFilter struct {
	CreatedAt struct {
		GTE time.Time `bson:"$gte"`
		LT  time.Time `bson:"$lt"`
	} `bson:"createdAt"`
}

func (*Taskdb) Find(ctx context.Context, _ *TaskFilter) ([]EtlTask, error) {
	panic("not implemented")
}

func (db *Taskdb) Delete(ctx context.Context, filter *TaskFilter) error {
	result, err := db.DeleteMany(ctx, bson.M{
		"createdAt": bson.M{"$gte": filter.CreatedAt.GTE, "$lt": filter.CreatedAt.LT}})
	if err != nil {
		return fmt.Errorf("%w: db.%s.DeleteMany: %v", ErrDatabase, db.Name(), err)
	}
	logger.FromContext(ctx).Infof("db.%s.deleteMany(%+v): %+v\n", db.Name(), filter, result)

	return nil
}

func NewTaskdb(mongouri, db, collection string) *Taskdb {
	coll, err := newMongodb(mongouri, db, collection)
	if err != nil {
		util.Fatal("Taskdb: newMongodb: ", err)
	}
	return &Taskdb{Collection: coll}
}

type logsysEtl struct {
	taskdb *Taskdb
}

// DeleteEtlTasks implements LogetlRepository.
func (etl *logsysEtl) DeleteEtlTasks(ctx context.Context, begin time.Time, end time.Time) error {
	for day := begin; day.Before(end); day = day.AddDate(0, 0, +1) {
		var filter TaskFilter

		filter.CreatedAt.GTE = day
		filter.CreatedAt.LT = day.AddDate(0, 0, +1)
		if err := etl.taskdb.Delete(ctx, &filter); err != nil {
			return err
		}
	}
	return nil
}

// FindEtlTasks implements LogetlRepository.
func (etl *logsysEtl) FindEtlTasks(ctx context.Context, begin time.Time, end time.Time) ([]EtlTask, error) {
	panic("unimplemented")
}

var _ LogEtlRepository = (*logsysEtl)(nil)

func NewLogEtlRepository(taskdb *Taskdb) LogEtlRepository {
	return &logsysEtl{taskdb: taskdb}
}
