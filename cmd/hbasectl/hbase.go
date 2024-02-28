package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"

	"github.com/qbox/net-deftones/logger"
)

type Cell struct {
	Row          string
	Columnfamily string
	Timestamp    time.Time
	Value        []byte
}

type HBase interface {
	// Get row or cell contents. pass table name, rowkey, and optional column family
	Get(_ context.Context, table, key string, col ...string) ([]Cell, error)
	Delete(ctx context.Context, table, key string, col ...string) error
}

// hbaseclient implements HBase interface
type hbaseclient struct {
	gohbase.Client
}

// DeleteAll implements Hbase.
func (hbase *hbaseclient) Delete(ctx context.Context, table string, key string, col ...string) error {
	var command strings.Builder
	fmt.Fprintf(&command, "delete '%s','%s'", table, key)
	if len(col) > 0 {
		fmt.Fprintf(&command, ",'%s'", col[0])
	}
	command.WriteByte('\n')
	fmt.Print(command.String())
	mu, err := hrpc.NewDel(ctx, []byte(table), []byte(key), nil)
	if err != nil {
		return fmt.Errorf("hrpc.NewDel: %w", err)
	}
	_, err = hbase.Client.Delete(mu)
	return err
}

func (hbase *hbaseclient) doGet(ctx context.Context, table, key string, cf map[string][]string) ([]*hrpc.Cell, error) {
	op, err := hrpc.NewGet(ctx, []byte(table), []byte(key), hrpc.Families(cf))
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
func (hbase *hbaseclient) Get(ctx context.Context, table string, key string, col ...string) ([]Cell, error) {
	logger.FromContext(ctx).Infof("get '%s', '%s'\n", table, key)
	var cells []Cell
	var cf = make(map[string][]string)

	for _, _cf := range col {
		cfd_qual := strings.SplitN(_cf, ":", 2)
		colfamily, qual := cfd_qual[0], cfd_qual[1]
		cf[colfamily] = append(cf[colfamily], qual)
	}
	hrpcCells, err := hbase.doGet(ctx, table, key, cf)
	if err != nil {
		return nil, fmt.Errorf("hbase.doGet: %w", err)
	}
	for _, cell := range hrpcCells {
		cells = append(cells, Cell{
			Row:          string(cell.Row),
			Columnfamily: string(cell.Family) + ":" + string(cell.Qualifier),
			Timestamp:    time.Unix(int64(*cell.Timestamp)/1e6, 0),
			Value:        cell.Value,
		})
	}
	return cells, nil
}

func (hbase *hbaseclient) Close() {
	hbase.Client.Close()
}

var _ HBase = (*hbaseclient)(nil)

func OpenHbase(addr string) *hbaseclient {
	client := gohbase.NewClient(addr, gohbase.EffectiveUser("qboxserver"))

	return &hbaseclient{Client: client}
}
