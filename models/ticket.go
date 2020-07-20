package models

// Ticket represents an issue
type Ticket struct {
	TicketID    int    `xorm:"pk"`
	Title       string `xorm:"text"`
	Tags        string `xorm:"text"`
	CreatedUnix int64  `xorm:"created"`
	Description string `xorm:"text"`
	Upvotes     int    `xorm:"default 0"`
	IsRep       bool   `xorm:"bool"` // Used for adding badge to emphasise rep tickets
}

// NewTicket inserts a new ticket into the database
func NewTicket(t Ticket) (err error) { // does it need to be *Ticket?
	_, err = engine.Insert(t)
	return err
}

// GetTicket fetches a ticket based on the TicketID
func GetTicket(id int) (ticket Ticket, err error) {

	has, err = engine.ID(id).Get()

	if err != nil {
		return ticket, err
	} else if !has {
		return ticket, err
	}

	return ticket, err
}

// GetTickets fetches all tickets in the database
func GetTickets() (tickets []Ticket) {
	engine.Find(&tickets)
	return tickets
}

// DelTicket deletes a ticket based on the TicketID
func DelTicket(id int) (err error) {
	_, err = engine.ID(id).Delete()
	return err
}
