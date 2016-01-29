package main

import (
	"math/rand"
	"time"
)

func NewInterval() time.Duration {
	interval := config.Interval * 60
	offsetRange := interval / 5
	offset := (rand.Int() % offsetRange) - (offsetRange / 2)
	return time.Second * time.Duration(interval+offset)
}
