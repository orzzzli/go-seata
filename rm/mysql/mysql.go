package mysql

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

const (
	DBDriver = "root:123456@tcp(127.0.0.1:3306)/user?charset=utf8"
)

var DBPool *sqlx.DB

func init() {
	DBPool, _ = sqlx.Open("mysql",DBDriver)
	DBPool.SetMaxOpenConns(10)
	DBPool.SetMaxIdleConns(5)
	DBPool.SetConnMaxLifetime(time.Second * 300)
	err := DBPool.Ping()
	if err != nil {
		log.Fatalf("init db error: %v",err)
	}
	log.Println("db worked")
}

type Db struct {}
