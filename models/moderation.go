package models

import "errors"

// Moderation represents an moderations
type Moderation struct {
	ModerationID         int64 `xorm:"pk autoincr"`
	CreatedUnix          int64 `xorm:"created"`
	UpdatedUnix          int64 `xorm:"updated"`
	Admin                string
	Title                string `xorm:"text"`
	Description          string `xorm:"text"`
	DescriptionSensitive bool
	Reason               string `xorm:"text"`
}

// AddModeration inserts a new moderations into the database
func AddModeration(a *Moderation) (err error) {
	_, err = engine.Insert(a)
	return err
}

// GetModeration fetches an announcement based on the ModerationID
func GetModeration(id int64) (*Moderation, error) {
	a := new(Moderation)
	has, err := engine.ID(id).Get(a)
	if err != nil {
		return a, err
	} else if !has {
		return a, errors.New("Doesn't exist")
	}

	return a, nil
}

// GetModerations fetches all moderations in the database
func GetModerations() (moderations []Moderation) {
	engine.Find(&moderations)
	return moderations
}
