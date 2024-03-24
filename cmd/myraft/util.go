package main

import (
	"math/rand/v2"
	"time"
)

func electionTimeout() time.Duration {
	return time.Duration(rand.Int()%151+150) * time.Millisecond
}
