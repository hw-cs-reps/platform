package models

// Ticket represents an issue
type Ticket struct {
	TicketID      int64  `xorm:"pk autoincr"`
	Title         string `xorm:"text"`
	Category      string `xorm:"text"`
	CreatedUnix   int64  `xorm:"created"`
	UpdatedUnix   int64  `xorm:"updated"`
	Description   string `xorm:"text"`
	Voters        []string
	IsRep         bool `xorm:"bool"` // Used for adding badge to emphasise rep tickets
	IsResolved    bool
	CommentsCount int       `xorm:"-"`
	Comments      []Comment `xorm:"-"`
}

// AddTicket inserts a new ticket into the database
func AddTicket(t *Ticket) (err error) {
	_, err = engine.Insert(t)
	return err
}

// UpdateTicket updates a comment in the database.
func UpdateTicket(t *Ticket) (err error) {
	_, err = engine.ID(t.TicketID).Update(t)
	return
}

// LoadComments loads the comments of the ticket into a non-mapped field.
func (t *Ticket) LoadComments() (err error) {
	return engine.Where("ticket_id = ?", t.TicketID).Find(&t.Comments)
}

// GetTicket fetches a ticket based on the TicketID
func GetTicket(id int64) (*Ticket, error) {
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

// GetCategory creates a new array of type Ticket and populates it with tickets of a certain category
func GetCategory(category string) (tickets []Ticket) {
	engine.Where("category = ?", category).Find(&tickets)
	return
}

// DelTicket deletes a ticket based on the TicketID
func DelTicket(id int64) (err error) {
	_, err = engine.ID(id).Delete(&Ticket{})
	return err
}

// UpdateTicketCols updates a ticket in the database including the specified
// columns, even if the fields are empty.
func UpdateTicketCols(t *Ticket, cols ...string) error {
	_, err := engine.ID(t.TicketID).Cols(cols...).Update(t)
	return err
}
