package config

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
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
	// LoadedDegrees is a list of degrees discovered in the config.
	LoadedDegrees []string
)

// Configuration represents the configuration file format.
type Configuration struct {
	SiteName        string                // SiteName is the name of the site.
	SiteScope       string                // SiteScope is the campus, department, and university year of the site.
	SitePort        string                // SitePort is the port to run the web server on.
	VoterPepper     string                // VoterPepper is the salt used in the voter ID hash.
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
	Courses      []Course
	Lecturers    []Lecturer
}

// ExternalResource holds the information to a hyperlink
type ExternalResource struct {
	Name, Link string
}

// ClassRepresentative holds the details of a class representative
type ClassRepresentative struct {
	Name, Email, Course, DegreeCode string
}

type Course struct {
	Code, Name string
	DegreeCode []string
}

type Lecturer struct {
	Name, Email, Office, Updated string
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
		SiteName:        "Platform",
		SiteScope:       "Edinburgh · MACS · Year 4",
		SitePort:        "8080",
		VoterPepper:     uuid.New().String(),
		DevMode:         true,
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
			ShowNotice:   true,
			NoticeTitle:  "Privacy Policy Update",
			NoticeText:   "The privacy policy has been updated. Please consider re-reading it for your peace of mind.",
			NoticeLink:   "/privacy",
			NoticeColour: "alert-green", // alert-green, alert-yellow, alert-red, alert-grey
			Links: []ExternalResource{
				{Name: "Example", Link: "https://example.com"},
				{Name: "Courses", Link: "/courses"},
				{Name: "Lecturers & Office Locations", Link: "/lecturers"},
			},
			ClassReps: []ClassRepresentative{
				{Name: "Alakbar", Email: "az40@hw.ac.uk", Course: "Student Officer", DegreeCode: "F291-COS"},
				{Name: "Humaid", Email: "ha82@hw.ac.uk", Course: "Computer Science", DegreeCode: "F291-COS"},
				{Name: "Maleeha", Email: "mr137@hw.ac.uk", Course: "Computer Systems", DegreeCode: "F2CC-CSE"},
				{Name: "James", Email: "jss2@hw.ac.uk", Course: "Information Systems", DegreeCode: "F2IS-ISY"},
			},
			Courses: []Course{
				{Code: "F20GA",
					Name:       "3D Graphics and Animation",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "F20AD",
					Name:       "Advanced Interaction Design",
					DegreeCode: []string{"F291-COS", "F2CC-CSE", "F2IS-ISY"}},

				{Code: "F20AN",
					Name:       "Advanced Network Security",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "F20AA",
					Name:       "Applied Text Analytics [DXB]",
					DegreeCode: []string{"F291-COS"}},

				{Code: "F20BD",
					Name:       "Big Data Management",
					DegreeCode: []string{"F291-COS", "F2CC-CSE", "F2IS-ISY"}},

				{Code: "F20BC",
					Name:       "Biologically Inspired Computation",
					DegreeCode: []string{"F291-COS"}},

				{Code: "F20GP",
					Name:       "Computer Games Programming",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "F20CN",
					Name:       "Computer Network Security",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "F20CA",
					Name:       "Conversational Agents and Spoken Language Processing [EDI]",
					DegreeCode: []string{"F291-COS"}},

				{Code: "F28CD",
					Name:       "Creative Design Project [EDI]",
					DegreeCode: []string{"F2CC-CSE"}},

				{Code: "F20DL",
					Name:       "Data Mining and Machine Learning",
					DegreeCode: []string{"F291-COS"}},

				{Code: "F20DV",
					Name:       "Data Visualisation and Analytics [DXB]",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "F20PB",
					Name:       "Design & Implementation",
					DegreeCode: []string{"F291-COS", "F2CC-CSE", "F2IS-ISY"}},

				{Code: "F20DE",
					Name:       "Digital and Knowledge Economy",
					DegreeCode: []string{"F2CC-CSE", "F2IS-ISY"}},

				{Code: "F20FO",
					Name:       "Digital Forensics [DXB]",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "C10DM",
					Name:       "Digital Marketing [EDI]",
					DegreeCode: []string{"F2IS-ISY"}},

				{Code: "F17SC",
					Name:       "Discrete Mathematics",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "F20EC",
					Name:       "e-Commerce Technology",
					DegreeCode: []string{"F291-COS", "F2CC-CSE", "F2IS-ISY"}},

				{Code: "C17EC",
					Name:       "Enterprise and its Business Environment [DXB]",
					DegreeCode: []string{"F2CC-CSE"}},

				{Code: "F28HS",
					Name:       "Hardware-Software Interface",
					DegreeCode: []string{"F2CC-CSE"}},

				{Code: "F20SC",
					Name:       "Industrial Programming",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "F20IF",
					Name:       "Information Systems Methodologies",
					DegreeCode: []string{"F291-COS", "F2CC-CSE", "F2IS-ISY"}},

				{Code: "F20IM",
					Name:       "Information Technology Master Class [EDI]",
					DegreeCode: []string{"F2IS-ISY"}},

				{Code: "F20RO",
					Name:       "Intelligent Robotics",
					DegreeCode: []string{"F291-COS"}},

				{Code: "F17LP",
					Name:       "Logic and Proof",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "C10SM",
					Name:       "Marketing and Management of SMEs [EDI]",
					DegreeCode: []string{"F2IS-ISY"}},

				{Code: "C18OP",
					Name:       "Operations Management",
					DegreeCode: []string{"F2CC-CSE"}},

				{Code: "F20PC",
					Name:       "Project Testing and Presentation",
					DegreeCode: []string{"F291-COS", "F2CC-CSE", "F2IS-ISY"}},

				{Code: "F20PA",
					Name:       "Research Methods & Requirements Engineering",
					DegreeCode: []string{"F291-COS", "F2CC-CSE", "F2IS-ISY"}},

				{Code: "C10RS",
					Name:       "Retail Marketing [EDI]",
					DegreeCode: []string{"F2IS-ISY"}},

				{Code: "F20RS",
					Name:       "Rigorous Methods for Software Engineering",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "F20SF",
					Name:       "Software Engineering Foundations [EDI]",
					DegreeCode: []string{"F2IS-ISY"}},
				{Code: "F20SA",
					Name:       "Statistical Modelling and Analysis",
					DegreeCode: []string{"F291-COS", "F2CC-CSE"}},

				{Code: "F27TS",
					Name:       "Technology in Society [EDI]",
					DegreeCode: []string{"F2CC-CSE"}},

				{Code: "C10CW",
					Name:       "The Contemporary Workfore [EDI]",
					DegreeCode: []string{"F2IS-ISY"}},
			},
			Lecturers: []Lecturer{

				// EDINBURGH

				{Name: "Alasdair Gray",
					Email:   "A.J.G.Gray@hw.ac.uk",
					Office:  "EM 1.39",
					Updated: "04/08/20"},

				{Name: "Andrew Ireland",
					Email:   "A.Ireland@hw.ac.uk",
					Office:  "EM G.57",
					Updated: "04/08/20"},

				{Name: "Ben Kenwright",
					Email:   "B.Kenwright@hw.ac.uk",
					Office:  "EM 1.43",
					Updated: "04/08/20"},

				{Name: "Christian Dondrup",
					Email:   "C.Dondrup@hw.ac.uk",
					Office:  "EM 1.44",
					Updated: "04/08/20"},

				{Name: "Diana Bental",
					Email:   "D.S.Bental@hw.ac.uk",
					Office:  "EM 1.05",
					Updated: "04/08/20"},

				{Name: "Ekaterina 'Katya' Komendantskaya",
					Email:   "E.Komendantskaya@hw.ac.uk",
					Office:  "EM G.26",
					Updated: "04/08/20"},

				{Name: "Fairouz Kamareddine",
					Email:   "F.D.Kamareddine@hw.ac.uk",
					Office:  "EM 1.65",
					Updated: "24/07/20"},

				{Name: "Hans Wolfgang Loidl",
					Email:   "H.W.Loidl@hw.ac.uk",
					Office:  "EM G.51",
					Updated: "04/08/20"},

				{Name: "Jennifer 'Jenny' Coady",
					Email:   "J.Coady@hw.ac.uk",
					Office:  "EM G.37",
					Updated: "04/08/20"},

				{Name: "Jessica Chen-Burger",
					Email:   "Y.J.ChenBurger@hw.ac.uk",
					Office:  "EM G.38",
					Updated: "04/08/20"},

				{Name: "Lilia Georgieva",
					Email:   "L.Georgieva@hw.ac.uk",
					Office:  "EM G.54",
					Updated: "04/08/20"},

				{Name: "Lynne Baillie",
					Email:   "L.Baillie@hw.ac.uk",
					Office:  "EM G.30",
					Updated: "04/08/20"},

				{Name: "Manuel Maarek",
					Email:   "M.Maarek@hw.ac.uk",
					Office:  "EM 1.63",
					Updated: "04/08/20"},

				{Name: "Marcelo Pereyra",
					Email:   "M.Pereyra@hw.ac.uk",
					Office:  "Unknown",
					Updated: "04/08/20"},

				{Name: "Michael Lones",
					Email:   "M.Lones@hw.ac.uk",
					Office:  "EM G.31",
					Updated: "04/08/20"},

				{Name: "Mike Chantler",
					Email:   "M.J.Chantler@hw.ac.uk",
					Office:  "Unknown",
					Updated: "24/07/20"},

				{Name: "Mike Just",
					Email:   "M.Just@hw.ac.uk",
					Office:  "EM 1.37",
					Updated: "04/08/20"},

				{Name: "Nick Taylor",
					Email:   "N.K.Taylor@hw.ac.uk",
					Office:  "EM 1.62",
					Updated: "24/07/20"},

				{Name: "Oliver Lemon",
					Email:   "O.Lemon@hw.ac.uk",
					Office:  "EM 1.40",
					Updated: "04/08/20"},

				{Name: "Phil Bartie",
					Email:   "Phil.Bartie@hw.ac.uk",
					Office:  "EM G.29",
					Updated: "04/08/20"},

				{Name: "Ron Petrick",
					Email:   "R.Petrick@hw.ac.uk",
					Office:  "EM 1.60",
					Updated: "04/08/20"},

				{Name: "Santiago Chumbe",
					Email:   "S.Chumbe@hw.ac.uk",
					Office:  "EM G.41",
					Updated: "04/08/20"},

				{Name: "Stefano Padilla",
					Email:   "S.Padilla@hw.ac.uk",
					Office:  "EM 1.38",
					Updated: "04/08/20"},

				{Name: "Verena Rieser",
					Email:   "V.T.Rieser@hw.ac.uk",
					Office:  "EM 1.36",
					Updated: "04/08/20"},

				{Name: "Wei Pang",
					Email:   "pang.wei@abdn.ac.uk",
					Office:  "Unknown",
					Updated: "04/08/20"},

				// DUBAI

				{Name: "Adrian Turcanu",
					Email:   "A.Turcanu@hw.ac.uk",
					Office:  "S2-36",
					Updated: "04/08/20"},

				{Name: "Abrar Ullah",
					Email:   "A.Ullah@hw.ac.uk",
					Office:  "Unknown",
					Updated: "04/08/20"},

				{Name: "Hani Ragab Hassen",
					Email:   "H.RagabHassen@hw.ac.uk",
					Office:  "S2-33",
					Updated: "04/08/20"},

				{Name: "Hind Zantout",
					Email:   "H.Zantout@hw.ac.uk",
					Office:  "S3-12",
					Updated: "04/08/20"},

				{Name: "Mohammad Hamdan",
					Email:   "M.Hamdan@hw.ac.uk",
					Office:  "S3-14",
					Updated: "04/08/20"},

				{Name: "Neamat El Gayar",
					Email:   "N.Elgayar@hw.ac.uk",
					Office:  "S2-55",
					Updated: "04/08/20"},

				{Name: "Ryad Soobany",
					Email:   "R.Soobhany@hw.ac.uk",
					Office:  "Unknown",
					Updated: "04/08/20"},

				{Name: "Smitha Kumar",
					Email:   "Smitha.Kumar@hw.ac.uk",
					Office:  "Unknown",
					Updated: "04/08/20"},

				{Name: "Stephen 'Steve' Gill",
					Email:   "S.Gill@hw.ac.uk",
					Office:  "S3-13",
					Updated: "04/08/20"},

				{Name: "Talal Shaikh",
					Email:   "T.A.G.Shaikh@hw.ac.uk",
					Office:  "S3-11",
					Updated: "04/08/20"},
			},
		},
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	var err error
	WorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatal("Cannot get working directory!", err)
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

	has := make(map[string]bool)
	for _, c := range Config.InstanceConfig.Courses {
		for _, dc := range c.DegreeCode {
			if !has[dc] {
				LoadedDegrees = append(LoadedDegrees, dc)
				has[dc] = true
			}
		}
	}
}

// SaveConfig saves the configuration from memory to disk.
func SaveConfig() error {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(Config); err != nil {
		return err
	}

	if err := ioutil.WriteFile(ConfigPath, buf.Bytes(), 0600); err != nil {
		return err
	}
	return nil
}
