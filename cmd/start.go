package cmd

import (
	"fmt"

	"github.com/hw-cs-reps/platform/config"
	"github.com/hw-cs-reps/platform/models"
	"github.com/hw-cs-reps/platform/routes"

	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-macaron/cache"
	"github.com/go-macaron/captcha"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/mysql" // MySQL driver for persistent sessions
	"github.com/hako/durafmt"
	"github.com/urfave/cli/v2"
	macaron "gopkg.in/macaron.v1"
)

// CmdStart represents a command-line command
// which starts the bot.
var CmdStart = &cli.Command{
	Name:    "run",
	Aliases: []string{"start", "web"},
	Usage:   "Start the web server",
	Action:  start,
}

func start(clx *cli.Context) (err error) {
	config.LoadConfig()
	engine := models.SetupEngine()
	defer engine.Close()

	// Run macaron
	m := macaron.Classic()
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Funcs: []template.FuncMap{map[string]interface{}{
			"CalcTime": func(sTime time.Time) string {
				return fmt.Sprint(time.Since(sTime).Nanoseconds() / int64(time.Millisecond))
			},
			"EmailToUser": func(s string) string {
				if strings.Contains(s, "@") {
					return strings.Split(s, "@")[0]
				} else {
					return s
				}
			},
			"CalcDurationShort": func(unix int64) string {
				return durafmt.Parse(time.Since(time.Unix(unix, 0))).LimitFirstN(1).String()
			},
			"Len": func(arr []string) int {
				return len(arr)
			},
		}},
		IndentJSON: true,
	}))

	if config.Config.DevMode {
		fmt.Println("In development mode.")
		macaron.Env = macaron.DEV
	} else {
		fmt.Println("In production mode.")
		macaron.Env = macaron.PROD
	}

	m.Use(cache.Cacher())
	sessOpt := session.Options{
		CookieLifeTime: 2629744, // 1 month policy
		Gclifetime:     3600,    // gc every 1 hour
		CookieName:     "hithereimacookie",
	}
	if config.Config.DBConfig.Type == config.MySQL {
		sqlConfig := fmt.Sprintf("%s:%s@tcp(%s)/%s",
			config.Config.DBConfig.User, config.Config.DBConfig.Password,
			config.Config.DBConfig.Host, config.Config.DBConfig.Name)
		sessOpt.Provider = "mysql"
		sessOpt.ProviderConfig = sqlConfig
		sessOpt.CookieLifeTime = 0
	}
	m.Use(session.Sessioner(sessOpt))
	m.Use(csrf.Csrfer())
	m.Use(captcha.Captchaer())
	m.Use(routes.ContextInit())

	m.Get("/", routes.HomepageHandler)
	m.Group("/tickets", func() {
		m.Get("", routes.TicketsHandler)
		m.Get("/new", routes.NewTicketHandler)
		m.Post("/new", csrf.Validate, routes.PostNewTicketHandler)
		m.Group("/:id", func() {
			m.Get("", routes.TicketPageHandler)
			m.Post("", csrf.Validate, routes.PostTicketPageHandler) // comment post
			m.Post("/upvote", csrf.Validate, routes.UpvoteTicketHandler)
			m.Get("/edit", routes.TicketEditHandler)
			m.Post("/edit", csrf.Validate, routes.PostTicketEditHandler)
			m.Post("/delete", csrf.Validate, routes.PostTicketDeleteHandler)
		})
	})
	m.Get("/complaints", routes.ComplaintsHandler)
	m.Post("/complaints", csrf.Validate, routes.PostComplaintsHandler)
	m.Get("/courses", routes.CoursesHandler)
	m.Get("/lecturers", routes.LecturerHandler)
	m.Get("/privacy", routes.PrivacyHandler)

	m.Get("/login", routes.LoginHandler)
	m.Post("/login", csrf.Validate, routes.PostLoginHandler)
	m.Get("/verify", routes.VerifyHandler)
	m.Post("/verify", csrf.Validate, routes.PostVerifyHandler)
	m.Get("/logout", routes.LogoutHandler)
	m.Post("/cancel", routes.CancelHandler)

	log.Printf("Starting web server on port %s\n", config.Config.SitePort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.Config.SitePort), m))
	return nil
}
