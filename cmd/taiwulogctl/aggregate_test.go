package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extraDomainFrom(t *testing.T) {
	text := []byte(`audiosdk.xmcdn.com_202312260120-0000.json`)
	match := extraDomainFrom(string(text))
	assert.Equal(t, "audiosdk.xmcdn.com", match)
}
