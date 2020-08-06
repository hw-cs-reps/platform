package routes

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/hw-cs-reps/platform/models"

	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

// AnnouncementsHandler response for the announcements listing page.
func AnnouncementsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	announcements := models.GetAnnouncements()
	ctx.Data["Title"] = "Announcements"
	ctx.Data["IsAnnouncements"] = 1
	ctx.Data["Announcements"] = announcements

	ctx.HTML(200, "announcements")
}

// AnnouncementHandler response for the announcements listing page.
func AnnouncementHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["Title"] = "Announcement"
	ctx.Data["IsAnnouncements"] = 1
	announcement, err := models.GetAnnouncement(ctx.ParamsInt64("id"))
	if err != nil {
		log.Println(err)
		ctx.Redirect("/tickets")
		return
	}

	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["FormattedPost"] = template.HTML(markdownToHTML(announcement.Description))
	ctx.Data["Announcement"] = announcement
	ctx.HTML(200, "announcement")
}

// NewAnnouncementHandler response for posting new announcement.
func NewAnnouncementHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["IsAnnouncements"] = 1
	ctx.Data["Announcement"] = 1
	ctx.HTML(200, "new-ticket")
}

// PostNewAnnouncementHandler post response for posting new announcement.
func PostNewAnnouncementHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	title := strings.TrimFunc(ctx.QueryTrim("title"), IsImproperChar)
	text := strings.TrimFunc(ctx.QueryTrim("text"), IsImproperChar)

	if len(title) == 0 || len(text) == 0 {
		f.Error("Title or body cannot be empty!")
		ctx.Redirect("/a/new")
		return
	}

	announcement := models.Announcement{
		Title:       title,
		Description: text,
	}

	err := models.AddAnnouncement(&announcement)
	if err != nil {
		log.Println(err)
		f.Error("Failed to add ticket")
		ctx.Redirect("/a")
		return
	}
	ctx.Redirect(fmt.Sprintf("/a/%d", announcement.AnnouncementID))
}

// PostAnnouncementEditHandler response for adding posting a new announcement.
func PostAnnouncementEditHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	if ctx.QueryTrim("title") == "" || ctx.QueryTrim("text") == "" {
		announcement, err := models.GetAnnouncement(ctx.ParamsInt64("id"))
		if err != nil {
			log.Println(err)
			ctx.Redirect("/tickets")
			return
		}
		ctx.Data["csrf_token"] = x.GetToken()
		ctx.Data["Announcement"] = announcement
		ctx.Data["ptitle"] = announcement.Title
		ctx.Data["ptext"] = announcement.Description
		ctx.Data["edit"] = 1

		ctx.HTML(200, "new-ticket")
		return
	}
	announcement, err := models.GetAnnouncement(ctx.ParamsInt64("id"))
	if err != nil {
		ctx.Redirect(fmt.Sprintf("/a/%d", ctx.ParamsInt64("id")))
		return
	}

	title := ctx.QueryTrim("title")
	text := ctx.QueryTrim("text")

	err = models.UpdateAnnouncement(&models.Announcement{
		AnnouncementID: announcement.AnnouncementID,
		Title:          title,
		Description:    text,
	})
	if err != nil {
		panic(err)
	}

	ctx.Redirect(fmt.Sprintf("/a/%d", ctx.ParamsInt64("id")))
}

// PostAnnouncementDeleteHandler response for deleting an announcement.
func PostAnnouncementDeleteHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	models.DelAnnouncement(ctx.ParamsInt64("id"))
	f.Success("Announcement deleted!")
	ctx.Redirect("/a")
}
