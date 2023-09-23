package main

import (
	"fmt"
	"os"
)

type onSignal struct {
	sig os.Signal
}

// Error implements error.
func (e *onSignal) Error() string {
	return fmt.Sprintf("The signal %v received, shutdown ...", e.sig)
}

var _ error = (*onSignal)(nil)
