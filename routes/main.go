package routes

import (
	"bytes"
	"html/template"
	"log"
	"strconv"

	"github.com/hw-cs-reps/platform/config"
	"github.com/hw-cs-reps/platform/models"

	"github.com/go-macaron/session"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	macaron "gopkg.in/macaron.v1"
)

// HomepageHandler response for the home page.
func HomepageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Config"] = config.Config.InstanceConfig
	ctx.Data["IsHome"] = 1
	ctx.Data["Title"] = "Class Reps"
	ctx.HTML(200, "index")
}

// ComplaintsHandler response for the complaints page.
func ComplaintsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.HTML(200, "complaints")
}

// PostComplaintsHandler response for the complaints page.
func PostComplaintsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.HTML(200, "complaints")
}

// TicketsHandler response for the tickets listing page.
func TicketsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Tickets"] = models.GetTickets()
	ctx.Data["Title"] = "Tickets"
	ctx.HTML(200, "tickets")
}

// markdownToHTML converts a string (in Markdown) and outputs (X)HTML.
// The input may also contain HTML, and the output is sanitized.
func markdownToHTML(s string) string {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(s), &buf); err != nil {
		panic(err)
	}
	return string(bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes()))
}

// TicketPageHandler response for the a specific ticket..
func TicketPageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ticket, err := models.GetTicket(ctx.ParamsInt("id"))
	if err != nil {
		log.Println(err)
		ctx.Redirect("/tickets")
		return
	}
	ctx.Data["FormattedPost"] = template.HTML(markdownToHTML(ticket.Description))
	ctx.Data["Ticket"] = ticket
	ctx.HTML(200, "ticket")
}

// NewTicketsHandler response for posting new ticket.
func NewTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.HTML(200, "new-ticket")
}

// PostNewTicketsHandler post response for posting new ticket.
func PostNewTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	title := ctx.Query("title")
	text := ctx.Query("text")
	ticket := models.Ticket{
		Title:       title,
		Description: text,
	}
	err := models.AddTicket(&ticket)
	if err != nil {
		log.Println(err)
		f.Error("Failed to add ticket")
		ctx.Redirect("/tickets")
		return
	}
	ctx.Redirect("/tickets/" + strconv.Itoa(ticket.TicketID))
}

// UpvoteTicketHandler response for upvoting a specific ticket.
func UpvoteTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	// Limit upvotes by IP?
}

// TicketEditHandler response for adding posting new ticket.
func TicketEditHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.HTML(200, "new-ticket")
}

// PostTicketEditHandler response for adding posting new ticket.
func PostTicketEditHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
}

// PostTicketDeleteHandler response for deleting a ticket.
func PostTicketDeleteHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
}
