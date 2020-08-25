package routes

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-emmanuel/csrf"
	"github.com/go-emmanuel/emmanuel"
	"github.com/go-emmanuel/session"
	"github.com/hw-cs-reps/platform/config"
	"github.com/hw-cs-reps/platform/mailer"
)

// LoginHandler response for the login page.
func LoginHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
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
func PostLoginHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
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

	f.Error("You are not registered.")
	ctx.Redirect("/login")
}

// VerifyHandler post response for the login page.
func VerifyHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
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
func PostVerifyHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash) {
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
func CancelHandler(ctx *emmanuel.Context, sess session.Store) {
	if sess.Get("auth") != Verification {
		ctx.Redirect("/login")
		return
	}

	sess.Set("auth", LoggedOut)
	ctx.Redirect("/login")
}

// LogoutHandler response for the login page.
func LogoutHandler(ctx *emmanuel.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	sess.Set("auth", LoggedOut)
	sess.Set("isadmin", 0)
	sess.Set("user", "")
	ctx.Redirect("/")
}
