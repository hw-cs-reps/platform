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

	"github.com/go-emmanuel/csrf"
	"github.com/go-emmanuel/emmanuel"
	"github.com/go-emmanuel/session"
	"github.com/hw-cs-reps/platform/config"
	"github.com/hw-cs-reps/platform/models"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func getUsedCourses() (courses []config.Course) {
	for _, c := range config.Config.InstanceConfig.Courses {
		if models.HasTicketWithCategory(c.Code) {
			courses = append(courses, c)
		}
	}
	return
}

func isCourseOfDegree(code string, deg string) bool {
	for _, c := range config.Config.InstanceConfig.Courses {
		if c.Code == code {
			for _, d := range c.DegreeCode {
				if d == deg {
					return true
				}
			}

			break
		}
	}
	return false
}

// TicketsHandler response for the tickets listing page.
func TicketsHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	var tickets []models.Ticket

	if ctx.Params("category") != "" {
		tickets = models.GetCategory(ctx.Params("category"))
	} else if ctx.Params("degree") != "" {
		t := models.GetTickets()
		for i, tic := range t {
			if isCourseOfDegree(tic.Category, ctx.Params("degree")) {
				tickets = append(tickets, t[i])
			}
		}
	} else {
		tickets = models.GetTickets()
	}

	sort.Sort(models.HotTickets(tickets))
	hasResolved := false

	for i := range tickets {
		tickets[i].LoadComments()
		tickets[i].CommentsCount = len(tickets[i].Comments)
		if tickets[i].IsResolved {
			hasResolved = true
		}
	}

	ctx.Data["Tickets"] = tickets
	ctx.Data["IsTickets"] = 1
	ctx.Data["HasResolved"] = hasResolved
	ctx.Data["Title"] = "Tickets"
	ctx.Data["Category"] = ctx.Params("category")
	ctx.Data["Degree"] = ctx.Params("degree")
	ctx.Data["Courses"] = getUsedCourses()
	ctx.Data["LoadedDegrees"] = config.LoadedDegrees
	ctx.Data["csrf_token"] = x.GetToken()
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
func TicketPageHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["IsTickets"] = 1
	ticket, err := models.GetTicket(ctx.ParamsInt64("id"))
	if err != nil {
		log.Println(err)
		ctx.Redirect("/tickets")
		return
	}
	ctx.Data["Title"] = ticket.Title + " - Ticket"
	ctx.Data["Description"] = summariseMarkdown(ticket.Description)

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
func PostTicketPageHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
	ticket, err := models.GetTicket(ctx.ParamsInt64("id"))
	if err != nil {
		log.Println(err)
		ctx.Redirect("/tickets")
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

	if sess.Get("isadmin") == 1 && ctx.Query("as_admin") == "on" {
		comment.IsAdmin = true
		comment.PosterID = ctx.Data["User"].(config.ClassRepresentative).Name
	}

	err = models.AddComment(&comment)
	if err != nil {
		log.Println(err)
	}
	ctx.Redirect("/tickets/" + ctx.Params("id"))
}

// PostTicketSortHandler handles redirecting to a page of filtered tickets by category
func PostTicketSortHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
	category := ctx.Query("category")

	switch ctx.Query("type") {
	case "category":
		if !hasCategory(category) {
			f.Error("Can't sort by that category")
			ctx.Redirect("/tickets")
			return
		}

		ctx.Redirect("/tickets/cat/" + category)
	case "degree":
		if !hasDegree(category) {
			f.Error("Can't sort by that degree")
			ctx.Redirect("/tickets")
			return
		}

		ctx.Redirect("/tickets/deg/" + category)
	default:
		f.Error("Unknown filter")
		ctx.Redirect("/tickets")
	}
}

// NewTicketHandler response for posting new ticket.
func NewTicketHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["IsTickets"] = 1
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["Courses"] = config.Config.InstanceConfig.Courses
	ctx.Data["HasScope"] = 1
	ctx.HTML(200, "new-ticket")
}

// hasCategory checks if a category is listed in the configuration..
func hasCategory(category string) bool {
	for _, c := range config.Config.InstanceConfig.Courses {
		if category == c.Code || category == "General" {
			return true
		}
	}
	return false
}

// hasDegree checks if a degree is listed in the configuration.
func hasDegree(deg string) bool {
	for _, d := range config.LoadedDegrees {
		if d == deg {
			return true
		}
	}
	return false
}

// PostNewTicketHandler post response for posting new ticket.
func PostNewTicketHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
	title := strings.TrimFunc(ctx.QueryTrim("title"), IsImproperChar)
	text := strings.TrimFunc(ctx.QueryTrim("text"), IsImproperChar)
	category := ctx.QueryTrim("category")

	if !hasCategory(category) {
		f.Error("There was an error in creating your ticket")
		ctx.Redirect("/tickets")
		return
	}

	voterHash := userHash(getIP(ctx), ctx.Req.Header.Get("User-Agent"))

	if len(title) == 0 || len(text) < 4 {
		f.Error("Title or body cannot be empty!")
		ctx.Redirect("/tickets/new")
		return
	}

	if len(title) > 80 || len(text) > 2048 {
		f.Error("Title or body is too long!")
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
	//h.Write([]byte(ip + useragent + config.Config.VoterPepper))
	h.Write([]byte(ip + config.Config.VoterPepper))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getIP(ctx *emmanuel.Context) string {
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
func UpvoteTicketHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
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
func ResolveTicketHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
	ticket, err := models.GetTicket(ctx.ParamsInt64("id"))
	if err != nil {
		log.Println(err)
		ctx.Redirect("/tickets")
		return
	}

	ticket.IsResolved = !ticket.IsResolved
	models.UpdateTicketCols(ticket, "is_resolved")

	m := models.Moderation{
		Admin: ctx.Data["User"].(config.ClassRepresentative).Name,
		Title: "Ticket \"" + ticket.Title + "\"",
	}
	if ticket.IsResolved {
		m.Description = "Marked ticket as resolved"
	} else {
		m.Description = "Marked ticket as unresolved"
	}
	models.AddModeration(&m)

	ctx.Redirect("/tickets/" + strconv.Itoa(ctx.ParamsInt("id")))
}

// PostTicketEditHandler response for adding posting new ticket.
func PostTicketEditHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
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
		ctx.Data["Courses"] = config.Config.InstanceConfig.Courses
		ctx.Data["pcategory"] = ticket.Category
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

	m := models.Moderation{
		Admin: ctx.Data["User"].(config.ClassRepresentative).Name,
		Title: "Ticket \"" + ticket.Title + "\"",
	}

	title := ctx.QueryTrim("title")
	text := ctx.QueryTrim("text")
	category := ctx.QueryTrim("category")

	modDesc := strings.Builder{}

	if title != "" && title != ticket.Title {
		modDesc.WriteString("changed title from \"" + ticket.Title + "\" to \"" + title + "\"")
	}
	if text != "" && text != ticket.Description {
		if modDesc.Len() > 0 {
			modDesc.WriteString(" and also ")
		}
		modDesc.WriteString("changed description from \"" + ticket.Description + "\" to \"" + text + "\"")
	}
	if category != "" && category != ticket.Category {
		if modDesc.Len() > 0 {
			modDesc.WriteString(" and also ")
		}
		modDesc.WriteString("changed category from \"" + ticket.Category + "\" to \"" + category + "\"")
	}

	m.DescriptionSensitive = (ctx.Query("sensitive") == "on")

	m.Reason = ctx.QueryTrim("reason")
	m.Description = modDesc.String()

	if !hasCategory(category) {
		f.Error("Invalid category, ticket unchanged")
		ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
		return
	}

	models.AddModeration(&m)

	// TODO update category
	err = models.UpdateTicket(&models.Ticket{
		TicketID:    ticket.TicketID,
		Title:       title,
		Description: text,
		Category:    category,
	})
	if err != nil {
		panic(err)
	}

	ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
}

// PostTicketDeleteHandler response for deleting a ticket.
func PostTicketDeleteHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
	t, err := models.GetTicket(ctx.ParamsInt64("id"))
	if err != nil {
		f.Error("Ticket not found!")
		ctx.Redirect("/tickets")
		return
	}
	m := models.Moderation{
		Admin:       ctx.Data["User"].(config.ClassRepresentative).Name,
		Title:       "Ticket \"" + t.Title + "\"",
		Description: "Deleted",
	}
	models.AddModeration(&m)

	models.DelTicket(ctx.ParamsInt64("id"))
	f.Success("Ticket deleted!")
	ctx.Redirect("/tickets")
}

// PostCommentDeleteHandler response for deleting a ticket's comment.
func PostCommentDeleteHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
	c, err := models.GetComment(ctx.ParamsInt64("cid"))
	if err != nil {
		f.Error("Comment not found!")
		ctx.Redirect("/tickets")
		return
	}
	t, err := models.GetTicket(c.TicketID)
	if err != nil {
		f.Error("Ticket of comment not found!")
		ctx.Redirect("/tickets")
		return
	}
	m := models.Moderation{
		Admin:       ctx.Data["User"].(config.ClassRepresentative).Name,
		Title:       "Comment by \"" + c.PosterID + "\" on \"" + t.Title + "\"",
		Description: "Deleted",
	}
	models.AddModeration(&m)

	f.Success("Comment deleted!")
	models.DeleteComment(ctx.ParamsInt64("cid"))
	ctx.Redirect(fmt.Sprintf("/tickets/%d", ctx.ParamsInt64("id")))
}
