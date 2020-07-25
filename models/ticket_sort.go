package models

import (
	"time"
)

var hotnessDelta float64 = 86400 // NOTE: 86400 is 1 day in seconds

// HotTicket implements sort.Interface for []Ticket based on iota score diminished
// by time.
type HotTickets []Ticket

func (p HotTickets) Len() int {
	return len(p)
}

func (p HotTickets) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func getHotScore(p Ticket) float64 {
	t := time.Now().Sub(time.Unix(p.CreatedUnix, 0)).Seconds() / hotnessDelta
	if t < 1 {
		return float64(len(p.Voters))
	}
	return float64(len(p.Voters)) / t
}

func (p HotTickets) Less(i, j int) bool {
	return getHotScore(p[i]) > getHotScore(p[j])
}
