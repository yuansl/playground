package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	r := LogLinkRequest{
		Domains: "a",
		Day:     time.Now(),
	}
	data, _ := json.Marshal(&r)

	var r0 LogLinkRequest

	err := json.Unmarshal(data, &r0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, r.Domains, r0.Domains)

	slices := strings.Split("v2/media.cdn.kuwo.cn_2023-07-12-23_part-00000.gz", "_")
	assert.Equal(t, slices[1][:13], "2023-07-12-23")
}
