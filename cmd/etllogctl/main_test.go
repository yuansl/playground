package main

import (
	"reflect"
	"testing"
	"time"
)

func Test_aligned5minTime(t *testing.T) {
	type args struct {
		t time.Time
	}
	now := time.Date(2023, 11, 12, 19, 27, 0, 3, time.Local)
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{args: args{t: now}, want: time.Date(2023, 11, 12, 19, 25, 0, 0, time.Local)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := aligned5minTime(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("aligned5minTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
