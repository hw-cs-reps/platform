package routes

import (
	"github.com/go-emmanuel/emmanuel"
	"github.com/go-emmanuel/session"

	"github.com/hw-cs-reps/platform/config"
	"github.com/hw-cs-reps/platform/namegen"

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
func ContextInit() emmanuel.Handler {
	return func(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
		ctx.Data["PageStartTime"] = time.Now()
		if sess.Get("auth") == nil {
			sess.Set("auth", LoggedOut)
		}
		if sess.Get("auth") == LoggedIn {
			ctx.Data["LoggedIn"] = 1
			ctx.Data["IsAdmin"] = sess.Get("isadmin")
			for _, c := range config.Config.InstanceConfig.ClassReps {
				if c.Email == sess.Get("user") {
					ctx.Data["User"] = c
				}
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
		ctx.Data["SiteScope"] = config.Config.SiteScope
	}
}

// RequireAdmin redirects if user is not an administrator
func RequireAdmin(ctx *emmanuel.Context, sess session.Store) {
	if !(sess.Get("auth") == LoggedIn && sess.Get("isadmin") == 1) {
		ctx.Redirect("/")
		return
	}
}
