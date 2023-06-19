package main

import (
	"log"
	"os"
)

type logger struct {
	*log.Logger
	uuid string
}

func NewLogger() *logger {
	return &logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Llongfile),
		uuid:   "this-is-a-uniq-id",
	}
}

func (logger *logger) Info(format string, v ...any) {
	logger.Printf("["+logger.uuid+"][INFO] "+format, v...)
}
