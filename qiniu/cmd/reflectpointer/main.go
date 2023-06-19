package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/qbox/net-deftones/clients/perspective"
)

func main() {
	var some = struct {
		perspective.CommonRequest
		Padding int64 `json:"X"`
	}{
		CommonRequest: perspective.CommonRequest{
			Domains:   []string{"www.example.com"},
			StartDate: time.Now().Add(-1 * time.Hour),
			EndDate:   time.Now(),
		},
	}

	data, _ := json.Marshal(some)

	fmt.Printf("Data = %q\n", data)
}
