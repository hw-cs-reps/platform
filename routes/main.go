package routes

import (
	"github.com/hw-cs-reps/platform/config"

	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

// HomepageHandler response for the home page.
func HomepageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Config"] = config.Config.InstanceConfig
	ctx.Data["IsHome"] = 1
	ctx.HTML(200, "index")
}

// TicketsHandler response for the tickets listing page.
func TicketsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.HTML(200, "tickets")
}

// TicketPageHandler response for the a specific ticket..
func TicketPageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.HTML(200, "ticket")
}

// NewTicketsHandler response for posting new ticket.
func NewTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.HTML(200, "new-tickets")
}

// UpvoteTicketHandler response for upvoting a specific ticket.
func UpvoteTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	// Limit upvotes by IP?
}

// TicketEditHandler response for adding posting new ticket.
func TicketEditHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.HTML(200, "new-tickets")
}

// PostTicketEditHandler response for adding posting new ticket.
func PostTicketEditHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
}

// PostTicketDeleteHandler response for deleting a ticket.
func PostTicketDeleteHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
}
