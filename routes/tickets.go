package routes

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"log"
	"sort"
	"strconv"
	"strings"
	"unicode"

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

func getUsedCourses() (courses []config.Course) {
	for _, c := range config.Config.InstanceConfig.Courses {
		if models.HasTicketWithCategory(c.Code) {
			courses = append(courses, c)
		}
	}
	return
}

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
	ctx.Data["Courses"] = getUsedCourses()
	ctx.Data["HasScope"] = 1
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
	ctx.Data["HasScope"] = 1
	ctx.HTML(200, "ticket")
}

// IsImproperChar checks to make sure that an empty message or ticket is not being submitted
func IsImproperChar(r rune) bool {
	return unicode.IsSpace(r) || !unicode.IsGraphic(r)
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

	text := strings.TrimFunc(ctx.QueryTrim("text"), IsImproperChar)
	if len(text) == 0 {
		f.Error("Comment cannot be empty!")
		ctx.Redirect("/tickets/" + ctx.Params("id"))
		return
	}

	comment := models.Comment{
		TicketID: ticket.TicketID,
		PosterID: sess.Get("id").(string),
		Text:     text,
	}

	err = models.AddComment(&comment)
	if err != nil {
		log.Println(err)
	}
	ctx.Redirect("/tickets/" + ctx.Params("id"))
}

// PostTicketSortHandler handles redirecting to a page of filtered tickets by category
func PostTicketSortHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	category := ctx.Query("category")

	if !isCategory(category) {
		f.Error("Can't sort by that category")
		ctx.Redirect("/tickets")
		return
	}

	ctx.Redirect("/tickets/cat/" + category)
}

// NewTicketHandler response for posting new ticket.
func NewTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["IsTickets"] = 1
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["Courses"] = config.Config.InstanceConfig.Courses
	ctx.Data["HasScope"] = 1
	ctx.HTML(200, "new-ticket")
}

// Checks if a category is listed in the configuration
func isCategory(category string) bool {
	var found bool

	for _, c := range config.Config.InstanceConfig.Courses {
		if category == c.Code {
			return true
		} else if category == "General" {
			return true
		}
	}
	return found
}

// PostNewTicketHandler post response for posting new ticket.
func PostNewTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	title := strings.TrimFunc(ctx.QueryTrim("title"), IsImproperChar)
	text := strings.TrimFunc(ctx.QueryTrim("text"), IsImproperChar)
	category := ctx.QueryTrim("category")

	if !isCategory(category) {
		f.Error("There was an error in creating your ticket")
		ctx.Redirect("/tickets")
		return
	}

	voterHash := userHash(getIP(ctx), ctx.Req.Header.Get("User-Agent"))

	if len(title) == 0 || len(text) == 0 {
		f.Error("Title or body cannot be empty!")
		ctx.Redirect("/tickets/new")
		return
	}

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

// ResolveTicketHandler response for resolving and unresolving a specific ticket.
func ResolveTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ticket, err := models.GetTicket(ctx.ParamsInt64("id"))
	if err != nil {
		log.Println(err)
		ctx.Redirect("/tickets")
		return
	}

	ticket.IsResolved = !ticket.IsResolved
	models.UpdateTicketCols(ticket, "is_resolved")

	ctx.Redirect("/tickets/" + strconv.Itoa(ctx.ParamsInt("id")))
}

// PostTicketEditHandler response for adding posting new ticket.
func PostTicketEditHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	if ctx.QueryTrim("title") == "" || ctx.QueryTrim("text") == "" {
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
	models.DelTicket(ctx.ParamsInt64("id"))
	f.Success("Ticket deleted!")
	ctx.Redirect("/tickets")
}

// PostCommentDeleteHandler response for deleting a ticket's comment.
func PostCommentDeleteHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	models.DeleteComment(ctx.ParamsInt64("cid"))
	f.Success("Comment deleted!")
	ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
}
