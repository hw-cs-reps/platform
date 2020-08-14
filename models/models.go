package models

import (
	"github.com/hw-cs-reps/platform/config"

	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // MySQL driver support
	_ "github.com/mattn/go-sqlite3"    // SQLite driver support
	"xorm.io/core"
	"xorm.io/xorm"
)

var (
	engine *xorm.Engine
	tables []interface{}
)

func init() {
	tables = append(tables,
		new(Announcement),
		new(Moderation),
		new(Ticket),
		new(Comment),
	)
}

// SetupEngine sets up an XORM engine according to the database configuration
// and syncs the schema.
func SetupEngine() *xorm.Engine {
	var err error
	dbConf := &config.Config.DBConfig

	address := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		dbConf.User, dbConf.Password, dbConf.Host, dbConf.Name)

	switch dbConf.Type {
	case config.MySQL:
		engine, err = xorm.NewEngine("mysql", address)
	case config.SQLite:
		engine, err = xorm.NewEngine("sqlite3", dbConf.Path)
	}

	if err != nil {
		log.Fatal("Unable to connect/load the database! ", err)
	}

	engine.SetMapper(core.GonicMapper{}) // So ID becomes 'id' instead of 'i_d'
	err = engine.Sync(tables...)         // Sync the schema of tables

	//cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 2000)
	//engine.SetDefaultCacher(cacher)

	if err != nil {
		log.Fatal("Unable to sync schema! ", err)
	}

	return engine
}
