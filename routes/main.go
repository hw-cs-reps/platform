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
	ctx.HTML(200, "index")
}
