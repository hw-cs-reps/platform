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
	ctx.Data["Title"] = config.Config.SiteName
	ctx.HTML(200, "index")
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

// ConfigHandler gets courses page
func ConfigHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	if !(sess.Get("auth") == LoggedIn && sess.Get("isadmin") == 1) {
		ctx.Redirect("/")
		return
	}
	ctx.Data["Title"] = "Configuration"
	ctx.HTML(200, "config")
}
