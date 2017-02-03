package db

import "time"

//Report represents a client report
type Report struct {
	id             uint32
	MacAddress     string
	Location       string
	FileSetVersion uint32
	Time           time.Time
}
