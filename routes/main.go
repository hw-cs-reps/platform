package routes

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/hw-cs-reps/platform/config"
	"github.com/hw-cs-reps/platform/mailer"
	"github.com/hw-cs-reps/platform/models"

	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	macaron "gopkg.in/macaron.v1"
)

// LoginHandler response for the login page.
func LoginHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["Title"] = config.Config.SiteName
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "login")
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randIntRange(min, max int) int {
	return rand.Intn(max-min) + min
}

// PostLoginHandler post response for the login page.
func PostLoginHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	if sess.Get("auth") == Verification {
		ctx.Redirect("/verify")
		return
	} else if sess.Get("auth") == LoggedIn {
		ctx.Redirect("/")
		return
	}
	for _, c := range config.Config.InstanceConfig.ClassReps {
		if c.Email == ctx.QueryTrim("email")+config.Config.UniEmailDomain {
			// Generate code
			code := fmt.Sprint(randIntRange(100000, 999999))
			// Note: We assume email in config to be correct

			if !config.Config.DevMode {
				go mailer.EmailCode(c.Email, code)
			}
			sess.Set("auth", Verification)
			sess.Set("code", code)
			sess.Set("user", c.Email)
			sess.Set("attempts", 0)

			ctx.Redirect("/verify")
			return
		}
	}

	f.Error("This is for class representatives only")
	ctx.Redirect("/login")
}

// VerifyHandler post response for the login page.
func VerifyHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	if sess.Get("auth") == LoggedOut {
		ctx.Redirect("/login")
		return
	} else if sess.Get("auth") == LoggedIn || sess.Get("auth") != Verification {
		f.Info("You are already logged in!")
		ctx.Redirect("/")
		return
	}
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["email"] = sess.Get("user")
	ctx.Data["Title"] = "Verification"
	ctx.HTML(200, "verify_login")
}

// PostVerifyHandler post response for the login page.
func PostVerifyHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	if sess.Get("auth") == LoggedOut {
		ctx.Redirect("/login")
		return
	} else if sess.Get("auth") == LoggedIn {
		ctx.Redirect("/")
		return
	}
	sess.Set("attempts", sess.Get("attempts").(int)+1)
	if sess.Get("attempts").(int) > 3 {
		f.Error("You reached the maximum number of attempts. Please try again later.")
		sess.Set("auth", LoggedOut)
		ctx.Redirect("/")
		return
	}
	if ctx.QueryTrim("code") != sess.Get("code") && !config.Config.DevMode {
		f.Error("The code you entered is invalid, make sure you use the latest code sent to you.")
		ctx.Redirect("/verify")
		return
	}

	sess.Set("auth", LoggedIn)
	sess.Set("isadmin", 1)
	ctx.Redirect("/")
}

// CancelHandler post response for canceling verification.
func CancelHandler(ctx *macaron.Context, sess session.Store) {
	if sess.Get("auth") != Verification {
		ctx.Redirect("/login")
		return
	}

	sess.Set("auth", LoggedOut)
	ctx.Redirect("/login")
}

// LogoutHandler response for the login page.
func LogoutHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	sess.Set("auth", LoggedOut)
	sess.Set("isadmin", 0)
	sess.Set("user", "")
	ctx.Redirect("/")
}

// HomepageHandler response for the home page.
func HomepageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Config"] = config.Config.InstanceConfig
	ctx.Data["IsHome"] = 1
	ctx.Data["Title"] = config.Config.SiteName
	ctx.HTML(200, "index")
}

// ComplaintsHandler response for the complaints page.
func ComplaintsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["Title"] = "Complaints"
	ctx.Data["IsComplaints"] = 1
	ctx.Data["Courses"] = config.Config.InstanceConfig.Courses
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "complaints")
}

func getClassRepsByCourseCode(code string) (recipients []*config.ClassRepresentative) {
	var degrees []string
	for _, c := range config.Config.InstanceConfig.Courses {
		if c.Code == code {
			degrees = c.DegreeCode
			break
		}
	}

	for i, c := range config.Config.InstanceConfig.ClassReps {
		for _, d := range degrees {
			if c.DegreeCode == d {
				recipients = append(recipients, &config.Config.InstanceConfig.ClassReps[i])
				break
			}
		}
	}
	return
}

// PostComplaintsHandler response for the complaints page.
func PostComplaintsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["IsComplaints"] = 1
	if ctx.QueryTrim("confirm") == "1" { // confirm sending
		var sender string
		if ctx.QueryTrim("email") == "" {
			sender = "anonymous"
		} else {
			sender = ctx.QueryTrim("email")
		}
		// TODO send to respective class reps

		crs := getClassRepsByCourseCode(ctx.QueryTrim("category"))
		var recipients []string
		for _, c := range crs {
			recipients = append(recipients, c.Email)
		}

		mailer.Email(recipients, "Complaint submission", `A complaint submission
From: `+sender+`
Category: `+ctx.QueryTrim("category")+`
Subject: `+ctx.QueryTrim("subject")+`
Message:
`+ctx.QueryTrim("message"))

		f.Success("Your complaint was sent!")
		ctx.Redirect("/complaints")
		return
	}

	ctx.Data["Category"] = ctx.QueryTrim("category")
	ctx.Data["Subject"] = ctx.QueryTrim("subject")
	ctx.Data["Message"] = ctx.QueryTrim("message")
	ctx.Data["Email"] = ctx.QueryTrim("Email")
	ctx.Data["csrf_token"] = x.GetToken()

	crs := getClassRepsByCourseCode(ctx.QueryTrim("category"))

	if len(crs) == 0 {
		f.Error("Sorry, no class representatives are available for the selected course/category.")
		ctx.Redirect("/complaints")
		return
	}

	var recipients []string
	for _, c := range crs {
		recipients = append(recipients, c.Name)
	}

	ctx.Data["Recipients"] = recipients
	ctx.HTML(200, "complaints-confirm")
}

// TicketsHandler response for the tickets listing page.
func TicketsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	tickets := models.GetTickets()
	sort.Sort(models.HotTickets(tickets))
	for i := range tickets {
		tickets[i].LoadComments()
		tickets[i].CommentsCount = len(tickets[i].Comments)
	}
	ctx.Data["Tickets"] = tickets
	ctx.Data["IsTickets"] = 1
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

// CoursesHandler gets courses page
func CoursesHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Courses"] = config.Config.InstanceConfig.Courses
	ctx.Data["Title"] = "Courses"
	ctx.HTML(200, "courses")
}

// LecturerHandler gets courses page
func LecturerHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Lecturers"] = config.Config.InstanceConfig.Lecturers
	ctx.Data["Title"] = "Lecturers"
	ctx.HTML(200, "lecturers")
}

// PrivacyHandler gets the privacy policy page
func PrivacyHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Title"] = "Privacy Policy"
	ctx.HTML(200, "privacy")
}

// TicketPageHandler response for the a specific ticket..
func TicketPageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["IsTickets"] = 1
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

// NewTicketsHandler response for posting new ticket.
func NewTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["IsTickets"] = 1
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "new-ticket")
}

// PostNewTicketsHandler post response for posting new ticket.
func PostNewTicketHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	title := ctx.QueryTrim("title")
	text := ctx.QueryTrim("text")
	voterHash := userHash(getIP(ctx), ctx.Req.Header.Get("User-Agent"))
	ticket := models.Ticket{
		Title:       title,
		Description: text,
		Voters:      []string{voterHash},
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
