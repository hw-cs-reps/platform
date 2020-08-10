package routes

import (
	"bytes"
	"log"

	"github.com/hw-cs-reps/platform/config"

	"github.com/BurntSushi/toml"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	macaron "gopkg.in/macaron.v1"
)

// HomepageHandler response for the home page.
func HomepageHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Config"] = config.Config.InstanceConfig
	ctx.Data["IsHome"] = 1
	ctx.Data["Title"] = config.Config.SiteName
	ctx.Data["HasScope"] = 1
	ctx.HTML(200, "index")
}

// CoursesHandler gets courses page
func CoursesHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Courses"] = config.Config.InstanceConfig.Courses
	ctx.Data["Title"] = "Courses"
	ctx.Data["HasScope"] = 1
	ctx.HTML(200, "courses")
}

// LecturerHandler gets courses page
func LecturerHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Lecturers"] = config.Config.InstanceConfig.Lecturers
	ctx.Data["Title"] = "Lecturers"
	ctx.Data["HasScope"] = 1
	ctx.HTML(200, "lecturers")
}

// PrivacyHandler gets the privacy policy page
func PrivacyHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Title"] = "Privacy Policy"
	ctx.HTML(200, "privacy")
}

// ConfigHandler gets courses page
func ConfigHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["Title"] = "Configuration"
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config.Config.InstanceConfig); err != nil {
		log.Println(err)
	}
	ctx.Data["Conf"] = buf.String()
	ctx.Data["HasScope"] = 1
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.HTML(200, "config")
}

// PostConfigHandler gets courses page
func PostConfigHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	var conf config.InstanceSettings
	err := toml.Unmarshal([]byte(ctx.Query("conf")), &conf)
	if err != nil {
		f.Error("Incorrect syntax in config! " + err.Error())
	}

	f.Success("Configuration updated correctly!")

	config.Config.InstanceConfig = conf
	config.SaveConfig()
	ctx.Redirect("/config")
}

// PreviewHandler gets the privacy policy page
func PreviewHandler(ctx *macaron.Context, sess session.Store, f *session.Flash) {
	ctx.Data["Title"] = "Preview"
	ctx.HTML(200, "preview")
}
