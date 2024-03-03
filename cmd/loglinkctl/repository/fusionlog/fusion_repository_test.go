package fusionlog

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yuansl/playground/cmd/loglinkctl/repository"
)

type mockLoglinkCURD struct {
	links []repository.LogLink
}

// Delete implements LoglinkStorage.
func (m *mockLoglinkCURD) Delete(ctx context.Context, link *repository.LogLink) error {
	panic("unimplemented")
}

// Query implements LoglinkCURD.
func (m *mockLoglinkCURD) Query(ctx context.Context, filter *Filter) ([]repository.LogLink, error) {
	return m.links, nil
}

// Update implements LoglinkCURD.
func (m *mockLoglinkCURD) Update(ctx context.Context, link *repository.LogLink) error {
	panic("unimplemented")
}

var _ LoglinkStorage = (*mockLoglinkCURD)(nil)

func TestGetLinks(t *testing.T) {
	fusionlog := &mockLoglinkCURD{links: []repository.LogLink{{Url: "kodofs:///home/log/www.example.com-20240201.gz",
		Timestamp: time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local),
	}}}
	repo, err := NewLinkRepository(fusionlog)
	if err != nil {
		t.Fatal(err)
	}
	begin, end := time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local), time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local)

	links, err := repo.GetLinks(context.TODO(), begin, end, &repository.LinkOptions{Filter: func(link *repository.LogLink) bool {
		return !strings.HasPrefix(link.Url, "http") && (link.Timestamp.Compare(begin) >= 0 && link.Timestamp.Before(end))
	}})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(links))
}
