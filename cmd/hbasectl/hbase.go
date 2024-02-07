package main

import (
	"context"
	"fmt"
	"time"

	"github.com/qbox/net-deftones/logger"
	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
)

const _HBASE_TABLENAME = "unify_flux"

type hbaseclient struct {
	gohbase.Client
	table string
}

// DeleteAll implements Hbase.
func (*hbaseclient) DeleteAll(ctx context.Context, table string, rowkey string) error {
	fmt.Printf("deleteall '%s', '%s'\n", table, rowkey)
	return nil
}

func (hbase *hbaseclient) doGet(ctx context.Context, table, rowkey string, cf map[string][]string) ([]*hrpc.Cell, error) {
	op, err := hrpc.NewGet(ctx, []byte(table), []byte(rowkey), hrpc.Families(cf), hrpc.CacheBlocks(true))
	if err != nil {
		return nil, fmt.Errorf("hbase.hrpc.NewGet: %w", err)
	}
	result, err := hbase.Client.Get(op)
	if err != nil {
		return nil, fmt.Errorf("hbase.Client.Get: %w", err)
	}
	return result.Cells, nil
}

// Get implements Hbase.
func (hbase *hbaseclient) Get(ctx context.Context, table string, rowkey string, colfamily ...string) ([]Cell, error) {
	logger.FromContext(ctx).Infof("get '%s', '%s'\n", table, rowkey)
	var cells []Cell
	var cf = make(map[string][]string)

	for _, _cf := range colfamily {
		cf["cf"] = append(cf["cf"], _cf)
	}
	hrpcCells, err := hbase.doGet(ctx, table, rowkey, cf)
	if err != nil {
		return nil, fmt.Errorf("hbase.doGet: %w", err)
	}
	for _, cell := range hrpcCells {
		cells = append(cells, Cell{
			Row:       string(cell.Row),
			Column:    string(cell.Family) + ":" + string(cell.Qualifier),
			Timestamp: time.Unix(int64(*cell.Timestamp)/1e6, 0),
			Value:     string(cell.Value)})
	}
	return cells, nil
}

func (hbase *hbaseclient) Close() {
	hbase.Client.Close()
}

var _ Hbase = (*hbaseclient)(nil)

func OpenHbase(addr string) *hbaseclient {
	client := gohbase.NewClient(addr)

	return &hbaseclient{Client: client, table: _HBASE_TABLENAME}
}
