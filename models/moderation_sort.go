package models

import (
	"time"
)

// ModerationSort implements sort.Interface for []Moderation depending on date.
type ModerationSort []Moderation

func (p ModerationSort) Len() int {
	return len(p)
}

func (p ModerationSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p ModerationSort) Less(i, j int) bool {
	return time.Unix(p[i].CreatedUnix, 0).After(time.Unix(p[j].CreatedUnix, 0))
}
