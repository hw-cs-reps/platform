package config

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

var (
	// WorkingDir is the current working directory of the project.
	WorkingDir string
	// ConfigPath is the configuration file name.
	ConfigPath = "config.toml"
	// Config is where the current configuration is loaded.
	Config Configuration
	// StartTime is the time when the server started.
	StartTime = time.Now()
)

// Configuration represents the configuration file format.
type Configuration struct {
	SiteName        string                // SiteName is the name of the site.
	SitePort        string                // SitePort is the port to run the web server on.
	DevMode         bool                  // DevMode is whether to disable authentication for development.
	UniEmailDomain  string                // UniEmailDomain is the university domain for login.
	EmailAddress    string                // EmailAddress is the email address which sends the OTPs.
	EmailPassword   string                // EmailPassword is the password of the email used to send OTPs.
	EmailSMTPServer string                // EmailSMTPServer is the SMTP server including the port.
	DBConfig        DatabaseConfiguration // DBConfig is the database configuration.
	InstanceConfig  InstanceSettings      // InstanceSettings is instance-specific configuration.
}

type InstanceSettings struct {
	ShowNotice   bool
	NoticeText   string
	NoticeLink   string
	NoticeTitle  string
	NoticeColour string
	Links        []ExternalResource
	ClassReps    []ClassRepresentative
}

// ExternalResource holds the information to a hyperlink
type ExternalResource struct {
	Name, Link string
}

// ClassRepresentative holds the details of a class representative
type ClassRepresentative struct {
	Name, Email, Course string
}

// DBType represents the type of the database driver which will be used.
type DBType int

const (
	// MySQL indicates to use the MySQL database driver.
	MySQL = iota
	// SQLite indicates to use the SQLite database driver.
	SQLite
)

// DatabaseConfiguration represents the general database configuration for all
// database drivers.
type DatabaseConfiguration struct {
	Type     DBType // Type refers to which database driver to use.
	Host     string // Host refers to the host of the database (MySQL only).
	Name     string // Name refers to the name of the database (MySQL only).
	User     string // User refers to the user of the database (MySQL only).
	Password string // Password refers to the database passsword (MySQL only).
	Path     string // Path refers to the database file path (SQLite only).
}

func newConfig() Configuration {
	return Configuration{
		SiteName:        "Class Representatives",
		SitePort:        "8080",
		DevMode:         false,
		UniEmailDomain:  "@hw.ac.uk",
		EmailAddress:    "noreply@example.com",
		EmailPassword:   "emailpasswordhere",
		EmailSMTPServer: "smtp.migadu.com:587",
		DBConfig: DatabaseConfiguration{
			Type:     SQLite,
			Host:     "localhost:3306",
			Name:     "notes",
			User:     "notes",
			Password: "passwordhere",
			Path:     "data.db",
		},
		InstanceConfig: InstanceSettings{
			ShowNotice: true,
			NoticeText: "This is an example notice!",
			NoticeLink: "https://example.com",
			Links: []ExternalResource{
				{Name: "Example", Link: "https://example.com"},
			},
			ClassReps: []ClassRepresentative{
				{Name: "Flan bin Flan", Email: "flan@example.com", Course: "Flani"},
			},
		},
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	var err error
	WorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatal("Cannot get working directory! ", err)
	}
}

// LoadConfig loads the configuration file from disk. It will also generate one
// if it doesn't exist.
func LoadConfig() {
	var err error
	if _, err = toml.DecodeFile(WorkingDir+"/"+ConfigPath, &Config); err != nil {
		log.Printf("Cannot load config file. Error: %s", err)
		if os.IsNotExist(err) {
			log.Println("Generating new configuration file, as it doesn't exist")
			var err error

			buf := new(bytes.Buffer)
			if err = toml.NewEncoder(buf).Encode(newConfig()); err != nil {
				log.Fatal(err)
			}

			err = ioutil.WriteFile(ConfigPath, buf.Bytes(), 0600)
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}
	}
}
