package routes

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"log"
	"sort"
	"strconv"

	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	"github.com/hw-cs-reps/platform/config"
	"github.com/hw-cs-reps/platform/models"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	macaron "gopkg.in/macaron.v1"
)

// TicketsHandler response for the tickets listing page.
func TicketsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {

	var tickets []models.Ticket

	if ctx.Params("category") != "" {
		tickets = models.GetCategory(ctx.Params("category"))
	} else {
		tickets = models.GetTickets()
	}

	sort.Sort(models.HotTickets(tickets))
	for i := range tickets {
		tickets[i].LoadComments()
		tickets[i].CommentsCount = len(tickets[i].Comments)
	}
	ctx.Data["Tickets"] = tickets
	ctx.Data["IsTickets"] = 1
	ctx.Data["Title"] = "Tickets"
	ctx.Data["Category"] = ctx.Params("category")
	ctx.Data["Courses"] = config.Config.InstanceConfig.Courses
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

// TicketPageHandler response for the a specific ticket.
func TicketPageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["IsTickets"] = 1
	ctx.Data["Title"] = "Ticket"
	ticket, err := models.GetTicket(ctx.ParamsInt64("id"))
	if err != nil {
		log.Println(err)
		ctx.Redirect("/tickets")
		return
	}

	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["FormattedPost"] = template.HTML(markdownToHTML(ticket.Description))
	ticket.LoadComments()
	ctx.Data["Ticket"] = ticket
	voterHash := userHash(getIP(ctx), ctx.Req.Header.Get("User-Agent"))
	ctx.Data["Upvoted"] = containsString(voterHash, ticket.Voters)
	ctx.HTML(200, "ticket")
}

// PostTicketPageHandler handles posting a new comment on a ticket.
func PostTicketPageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ticket, err := models.GetTicket(ctx.ParamsInt64("id"))
	if err != nil {
		log.Println(err)
		ctx.Redirect("/tickets")
		return
	}
	if ticket.IsResolved {
		ctx.Redirect("/tickets/" + ctx.Params("id"))
		return
	}

	comment := models.Comment{
		TicketID: ticket.TicketID,
		PosterID: sess.Get("id").(string),
		Text:     ctx.QueryTrim("text"),
	}

	err = models.AddComment(&comment)
	if err != nil {
		log.Println(err)
	}

	ctx.Redirect("/tickets/" + ctx.Params("id"))
}

// PostTicketSortHandler handles redirecting to a page of filtered tickets by category
func PostTicketSortHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Redirect("/tickets/cat/" + ctx.Query("category"))
}

// NewTicketHandler response for posting new ticket.
func NewTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["IsTickets"] = 1
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["Courses"] = config.Config.InstanceConfig.Courses
	ctx.HTML(200, "new-ticket")
}

// PostNewTicketHandler post response for posting new ticket.
func PostNewTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	title := ctx.QueryTrim("title")
	text := ctx.QueryTrim("text")
	category := ctx.QueryTrim("category")
	voterHash := userHash(getIP(ctx), ctx.Req.Header.Get("User-Agent"))
	ticket := models.Ticket{
		Title:       title,
		Description: text,
		Voters:      []string{voterHash},
		Category:    category,
	}
	err := models.AddTicket(&ticket)
	if err != nil {
		log.Println(err)
		f.Error("Failed to add ticket")
		ctx.Redirect("/tickets")
		return
	}
	ctx.Redirect(fmt.Sprintf("/tickets/%d", ticket.TicketID))
}

func userHash(ip string, useragent string) string {
	h := sha256.New()
	h.Write([]byte(ip + useragent + config.Config.VoterPepper))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getIP(ctx *macaron.Context) string {
	xf := ctx.Header().Get("X-Forwarded-For")
	if len(xf) > 0 {
		return xf
	}
	return ctx.RemoteAddr()
}

func containsString(s string, arr []string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

// UpvoteTicketHandler response for upvoting a specific ticket.
func UpvoteTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ticket, err := models.GetTicket(ctx.ParamsInt64("id"))
	if err != nil {
		log.Println(err)
		ctx.Redirect("/tickets")
		return
	}

	voterHash := userHash(getIP(ctx), ctx.Req.Header.Get("User-Agent"))
	if !ticket.IsResolved && !containsString(voterHash, ticket.Voters) {
		ticket.Voters = append(ticket.Voters, voterHash)
		models.UpdateTicketCols(ticket, "voters")
	}

	ctx.Redirect("/tickets/" + strconv.Itoa(ctx.ParamsInt("id")))
}

// PostTicketEditHandler response for adding posting new ticket.
func PostTicketEditHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	if !(sess.Get("auth") == LoggedIn && sess.Get("isadmin") == 1) {
		ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
	}
	if ctx.QueryTrim("title") == "" || ctx.QueryTrim("title") == "" {
		ctx.Data["IsTickets"] = 1
		ticket, err := models.GetTicket(ctx.ParamsInt64("id"))
		if err != nil {
			log.Println(err)
			ctx.Redirect("/tickets")
			return
		}
		ctx.Data["csrf_token"] = x.GetToken()
		ctx.Data["Ticket"] = ticket
		ctx.Data["ptitle"] = ticket.Title
		ctx.Data["ptext"] = ticket.Description
		ctx.Data["edit"] = 1

		ctx.HTML(200, "new-ticket")
		return
	}
	ticket, err := models.GetTicket(ctx.ParamsInt64("id"))
	if err != nil {
		ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
		return
	}

	title := ctx.QueryTrim("title")
	text := ctx.QueryTrim("text")

	err = models.UpdateTicket(&models.Ticket{
		TicketID:    ticket.TicketID,
		Title:       title,
		Description: text,
	})
	if err != nil {
		panic(err)
	}

	ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
}

// PostTicketDeleteHandler response for deleting a ticket.
func PostTicketDeleteHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	if !(sess.Get("auth") == LoggedIn && sess.Get("isadmin") == 1) {
		ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
	}
	models.DelTicket(ctx.ParamsInt64("id"))
	f.Success("Ticket deleted!")
	ctx.Redirect("/tickets")
}

// PostCommentDeleteHandler response for deleting a ticket's comment.
func PostCommentDeleteHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	if !(sess.Get("auth") == LoggedIn && sess.Get("isadmin") == 1) {
		ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
	}
	models.DeleteComment(ctx.ParamsInt64("cid"))
	f.Success("Comment deleted!")
	ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
}
