// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-06-19 07:55:25

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"context"
	"fmt"
	"log"

	hazelcast "github.com/hazelcast/hazelcast-go-client"
)

func main() {
	ctx := context.TODO()
	// create the client and connect to the cluster on localhost
	client, err := hazelcast.StartNewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// get a map
	people, err := client.GetMap(ctx, "people")
	if err != nil {
		log.Fatal(err)
	}
	personName := "Jane Doe"
	// set a value in the map
	if err = people.Set(ctx, personName, 30); err != nil {
		log.Fatal(err)
	}
	// get a value from the map
	age, err := people.Get(ctx, personName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s is %d years old.\n", personName, age)
	// stop the client to release resources
	client.Shutdown(ctx)
}
