package mailer

import (
	"github.com/hw-cs-reps/platform/config"

	"net/smtp"
	"strings"
	"time"
)

// Email sends an email to a specific email address.
func Email(to string, title string, message string) (err error) {
	from := config.Config.EmailAddress
	full := "From: <" + from + ">\n" +
		"To: <" + to + ">\n" +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\n" +
		"Subject: " + title + "\n\n" +
		message

	err = smtp.SendMail(config.Config.EmailSMTPServer,
		smtp.PlainAuth("", from, config.Config.EmailPassword, strings.Split(config.Config.EmailSMTPServer, ":")[0]),
		from, []string{to}, []byte(full))

	return err
}
