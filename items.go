package main

import "time"

type Item struct {
	Brand     string
	Model     string
	CaseSize  string
	CaseColor string
	BandColor string
	PhotoURL  string
	price     int
	DateAcq   time.Time
}
