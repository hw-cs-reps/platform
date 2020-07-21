package models

// Ticket represents an issue
type Ticket struct {
	TicketID      int       `xorm:"pk"`
	Title         string    `xorm:"text"`
	Tags          string    `xorm:"text"`
	CreatedUnix   int64     `xorm:"created"`
	UpdatedUnix   int64     `xorm:"updated"`
	Description   string    `xorm:"text"`
	Upvotes       int       `xorm:"default 0"`
	IsRep         bool      `xorm:"bool"` // Used for adding badge to emphasise rep tickets
	CommentsCount int       `xorm:"-"`
	Comments      []Comment `xorm:"-"`
}

// AddTicket inserts a new ticket into the database
func AddTicket(t *Ticket) (err error) {
	_, err = engine.Insert(t)
	return err
}

// GetTicket fetches a ticket based on the TicketID
func GetTicket(id int) (*Ticket, error) {
	t := new(Ticket)
	has, err := engine.ID(id).Get(t)
	if err != nil {
		return t, err
	} else if !has {
		return t, err
	}

	return t, nil
}

// GetTickets fetches all tickets in the database
func GetTickets() (tickets []Ticket) {
	engine.Find(&tickets)
	return tickets
}

// DelTicket deletes a ticket based on the TicketID
func DelTicket(id int) (err error) {
	_, err = engine.ID(id).Delete(&Ticket{})
	return err
}
