package routes

import (
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	"github.com/hw-cs-reps/platform/config"
	"github.com/hw-cs-reps/platform/mailer"
	macaron "gopkg.in/macaron.v1"
)

// ComplaintsHandler response for the complaints page.
func ComplaintsHandler(ctx *macaron.Context, sess session.Store, f *session.Flash, x csrf.CSRF) {
	ctx.Data["Title"] = "Complaints"
	ctx.Data["IsComplaints"] = 1
	ctx.Data["Courses"] = config.Config.InstanceConfig.Courses
	ctx.Data["csrf_token"] = x.GetToken()
	ctx.Data["HasScope"] = 1
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

		var recipients []string
		if ctx.QueryTrim("category") == "General" {
			for _, c := range config.Config.InstanceConfig.ClassReps {
				recipients = append(recipients, c.Email)
			}
		} else {
			crs := getClassRepsByCourseCode(ctx.QueryTrim("category"))
			for _, c := range crs {
				recipients = append(recipients, c.Email)
			}
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

	var recipients []string
	if ctx.QueryTrim("category") == "General" {
		for _, c := range config.Config.InstanceConfig.ClassReps {
			recipients = append(recipients, c.Name)
		}
	} else {
		crs := getClassRepsByCourseCode(ctx.QueryTrim("category"))
		if len(crs) == 0 {
			f.Error("Sorry, no class representatives are available for the selected course/category")
			ctx.Redirect("/complaints")
			return
		}

		for _, c := range crs {
			recipients = append(recipients, c.Name)
		}
	}

	ctx.Data["Recipients"] = recipients
	ctx.Data["HasScope"] = 1
	ctx.HTML(200, "complaints-confirm")
}
