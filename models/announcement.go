package models

import "errors"

// Announcement represents an announcement
type Announcement struct {
	AnnouncementID int64  `xorm:"pk autoincr"`
	Summary        string `xorm:"-"`
	Title          string `xorm:"text"`
	Tags           string `xorm:"text"`
	CreatedUnix    int64  `xorm:"created"`
	UpdatedUnix    int64  `xorm:"updated"`
	Description    string `xorm:"text"`
}

// AddAnnouncement inserts a new announcement into the database
func AddAnnouncement(a *Announcement) (err error) {
	_, err = engine.Insert(a)
	return err
}

// UpdateAnnouncement updates an announcement in the database.
func UpdateAnnouncement(a *Announcement) (err error) {
	_, err = engine.ID(a.AnnouncementID).Update(a)
	return
}

// GetAnnouncement fetches an announcement based on the AnnouncementID
func GetAnnouncement(id int64) (*Announcement, error) {
	a := new(Announcement)
	has, err := engine.ID(id).Get(a)
	if err != nil {
		return a, err
	} else if !has {
		return a, errors.New("Doesn't exist")
	}

	return a, nil
}

// GetAnnouncements fetches all announcement in the database
func GetAnnouncements() (announcements []Announcement) {
	engine.Find(&announcements)
	return announcements
}

// DelAnnouncement deletes a announcement based on the AnnouncementID
func DelAnnouncement(id int64) (err error) {
	_, err = engine.ID(id).Delete(&Announcement{})
	return err
}

// UpdateAnnouncementCols updates an announcement in the database including the
// specified columns, even if the fields are empty.
func UpdateAnnouncementCols(a *Announcement, cols ...string) error {
	_, err := engine.ID(a.AnnouncementID).Cols(cols...).Update(a)
	return err
}
