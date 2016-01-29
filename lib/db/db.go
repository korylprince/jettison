package db

import "time"

type Device struct {
	id           uint32
	SerialNumber string
	MacAddress   string
}

type Report struct {
	id             uint32
	Device         *Device
	Location       string
	FileSetVersion uint32
	Time           time.Time
}
