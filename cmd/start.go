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

	"github.com/go-emmanuel/cache"
	"github.com/go-emmanuel/captcha"
	"github.com/go-emmanuel/csrf"
	"github.com/go-emmanuel/emmanuel"
	"github.com/go-emmanuel/session"
	_ "github.com/go-emmanuel/session/mysql" // MySQL driver for persistent sessions
	"github.com/hako/durafmt"
	"github.com/urfave/cli/v2"
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

	// Run emmanuel
	m := emmanuel.Classic()
	m.Use(emmanuel.Renderer(emmanuel.RenderOptions{
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
			"Date": func(unix int64) string {
				return time.Unix(unix, 0).Format("Jan 2 2006")
			},
			"DateFull": func(unix int64) string {
				return time.Unix(unix, 0).Format("2006-01-02 15:04 -0700")
			},
			"Len": func(arr []string) int {
				return len(arr)
			},
			"Csv": func(s string) []string {
				return strings.Split(s, ",")
			},
			"Sep": func(sep string, s []string) string {
				str := strings.Builder{}
				for i, k := range s {
					str.WriteString(k)
					if i < len(s)-1 {
						str.WriteString(sep)
					}
				}
				return str.String()
			},
		}},
		IndentJSON: true,
	}))

	if config.Config.DevMode {
		fmt.Println("In development mode.")
		emmanuel.Env = emmanuel.DEV
	} else {
		fmt.Println("In production mode.")
		emmanuel.Env = emmanuel.PROD
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
	m.Get("/preview", routes.PreviewHandler)
	m.Group("/tickets", func() {
		m.Get("", routes.TicketsHandler)
		m.Get("/cat/:category", routes.TicketsHandler)
		m.Get("/deg/:degree", routes.TicketsHandler)
		m.Post("", csrf.Validate, routes.PostTicketSortHandler)
		m.Post("/cat/:category", csrf.Validate, routes.PostTicketSortHandler)
		m.Post("/deg/:degree", csrf.Validate, routes.PostTicketSortHandler)
		m.Get("/new", routes.NewTicketHandler)
		m.Post("/new", csrf.Validate, routes.PostNewTicketHandler)
		m.Group("/:id", func() {
			m.Get("", routes.TicketPageHandler)
			m.Post("", csrf.Validate, routes.PostTicketPageHandler) // comment post
			m.Post("/upvote", csrf.Validate, routes.UpvoteTicketHandler)

			// Admin
			m.Post("/resolve", routes.RequireAdmin, csrf.Validate, routes.ResolveTicketHandler)
			m.Post("/edit", routes.RequireAdmin, csrf.Validate, routes.PostTicketEditHandler)
			m.Post("/delete", routes.RequireAdmin, csrf.Validate, routes.PostTicketDeleteHandler)
			m.Post("/del/:cid", routes.RequireAdmin, csrf.Validate, routes.PostCommentDeleteHandler)
		})
	})

	m.Group("/a", func() {
		m.Get("", routes.AnnouncementsHandler)
		m.Group("/:id", func() {
			m.Get("", routes.AnnouncementHandler)
			m.Post("/edit", routes.RequireAdmin, csrf.Validate, routes.PostAnnouncementEditHandler)
			m.Post("/delete", routes.RequireAdmin, csrf.Validate, routes.PostAnnouncementDeleteHandler)
		})

		// Admin
		m.Get("/new", routes.RequireAdmin, routes.NewAnnouncementHandler)
		m.Post("/new", routes.RequireAdmin, routes.PostNewAnnouncementHandler)
	})

	m.Get("/complaints", routes.ComplaintsHandler)
	m.Post("/complaints", csrf.Validate, routes.PostComplaintsHandler)
	m.Get("/courses", routes.CoursesHandler)
	m.Get("/lecturers", routes.LecturerHandler)
	m.Get("/privacy", routes.PrivacyHandler)
	m.Get("/logs", routes.ModLogsHandler)

	m.Get("/login", routes.LoginHandler)
	m.Post("/login", csrf.Validate, routes.PostLoginHandler)
	m.Get("/verify", routes.VerifyHandler)
	m.Post("/verify", csrf.Validate, routes.PostVerifyHandler)
	m.Get("/logout", routes.LogoutHandler)
	m.Post("/cancel", csrf.Validate, routes.CancelHandler)

	// Admin
	m.Get("/config", routes.RequireAdmin, routes.ConfigHandler)
	m.Post("/config", routes.RequireAdmin, csrf.Validate, routes.PostConfigHandler)

	log.Printf("Starting web server on port %s\n", config.Config.SitePort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.Config.SitePort), m))
	return nil
}
