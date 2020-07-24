package routes

import (
	"github.com/go-macaron/session"
	"github.com/hw-cs-reps/platform/config"
	"github.com/hw-cs-reps/platform/models"
	"github.com/hw-cs-reps/platform/namegen"

	macaron "gopkg.in/macaron.v1"
	"time"
)

const (
	// LoggedOut is when a user is logged out.
	LoggedOut = iota
	// Verification is when a user is in the verification process.
	Verification
	// LoggedIn is when the user is verified and logged in.
	LoggedIn
)

// ContextInit is a middleware which initialises some global variables, and
// verifies the login status.
func ContextInit() macaron.Handler {
	return func(ctx *macaron.Context, sess session.Store, f *session.Flash) {
		ctx.Data["PageStartTime"] = time.Now()
		if sess.Get("auth") == nil {
			sess.Set("auth", LoggedOut)
		}
		if sess.Get("auth") == LoggedIn {
			ctx.Data["LoggedIn"] = 1
			ctx.Data["Username"] = sess.Get("user")
			if user, err := models.GetUser(sess.Get("user").(string)); err == nil {
				ctx.Data["User"] = user
			} else {
				// Let's log out the user
				ctx.Data["LoggedIn"] = 0
				sess.Set("auth", LoggedOut)
				f.Warning("You have been logged out.")
			}
		}
		ctx.Data["UniEmailDomain"] = config.Config.UniEmailDomain
		if config.Config.DevMode {
			ctx.Data["DevMode"] = 1
		}
		if sess.Get("id") == nil {
			sess.Set("id", namegen.GetName())
		}
		ctx.Data["SiteTitle"] = config.Config.SiteName
	}
}
