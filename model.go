package main

import "time"

type solardata struct {
	ID      int       `json:"id"`
	Created time.Time `json:"created"`
	Voltage float64   `json:"voltage"`
	Current float64   `json:"current"`
	Temp1   float64   `json:"temp1"`
	Temp2   float64   `json:"temp2"`
	Lum1    float64   `json:"lum1"`
	Lum2    float64   `json:"lum2"`
	Deleted bool      `json:"deleted"`
}

type solardebug struct {
	ID      int
	Created time.Time
	Message string
}
