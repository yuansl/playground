package types

import "time"

type Domain struct {
	Domain   string
	Cname    string
	CreateAt time.Time
}
