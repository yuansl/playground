package main

import (
	"log"

	"gopkg.in/mgo.v2"
)

func main() {
	session, err := mgo.Dial("admin:PiLIVdNMgo@127.0.0.1:3708")
	if err != nil {
		log.Fatalf("mgo.Dial error: %v\n", err)
	}
	info, err := session.BuildInfo()
	if err != nil {
		log.Fatalf("session.BuildInfo error: %v\n", err)
	}
	log.Printf("buildInfo: %+v\n", info)
}
