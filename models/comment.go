package models

import (
	"errors"
	"html/template"
)

// Comment represents an anonymous comment to a ticket.
type Comment struct {
	CommentID     int64         `xorm:"pk autoincr"`
	TicketID      int64         `xorm:"notnull"`
	PosterID      string        `xorm:"notnull"`
	Text          string        `xorm:"notnull"`
	FormattedText template.HTML `xorm:"-" json:"-"`
	CreatedUnix   int64         `xorm:"created"`
	UpdatedUnix   int64         `xorm:"updated"`
}

// AddComment adds a new Comment to the database.
func AddComment(c *Comment) (err error) {
	_, err = engine.Insert(c)
	return err
}

// UpdateComment updates a comment in the database.
func UpdateComment(c *Comment) (err error) {
	_, err = engine.ID(c.CommentID).Update(c)
	return
}

// GetComment gets a comment based on the ID.
// It will return the pointer to the Comment, and whether there was an error.
func GetComment(id string) (*Comment, error) {
	c := new(Comment)
	has, err := engine.ID(id).Get(c)
	if err != nil {
		return c, err
	} else if !has {
		return c, errors.New("Comment does not exist")
	}
	return c, nil
}

// DeleteComment deletes a comment from the database.
func DeleteComment(id string) (err error) {
	_, err = engine.ID(id).Delete(&Comment{})
	return
}
