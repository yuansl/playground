package main

import "time"

func alignoftime(t time.Time, alignas time.Duration) time.Time {
	if alignas > time.Hour {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}
	nanosecs := t.UnixNano() / alignas.Nanoseconds() * alignas.Nanoseconds()
	return time.Unix(nanosecs/1e9, nanosecs%1e9)
}
