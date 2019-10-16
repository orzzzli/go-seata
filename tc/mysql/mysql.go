package mysql

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/orzzzli/orzconfiger"
	"log"
	"time"
)

type DBPools struct {
	Pools map[string]*sqlx.DB
}

var DBPoolsInstance DBPools

func init() {
	DBPoolsInstance.Pools = make(map[string]*sqlx.DB)
}

func (p *DBPools) InitPool (name string) error {
	notFoundError := errors.New("config "+name+" not found.")
	host, find := orzconfiger.GetString(name,"host")
	if !find {
		return notFoundError
	}
	port, find := orzconfiger.GetString(name,"port")
	if !find {
		return notFoundError
	}
	user, find := orzconfiger.GetString(name,"user")
	if !find {
		return notFoundError
	}
	pass, find := orzconfiger.GetString(name,"pass")
	if !find {
		return notFoundError
	}
	database, find := orzconfiger.GetString(name,"database")
	if !find {
		return notFoundError
	}
	maxOpen, find := orzconfiger.GetInt(name,"maxOpen")
	if !find {
		return notFoundError
	}
	maxIdle, find := orzconfiger.GetInt(name,"maxIdle")
	if !find {
		return notFoundError
	}
	maxLife, find := orzconfiger.GetInt(name,"maxLife")
	if !find {
		return notFoundError
	}

	driverStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8",user,pass,host,port,database)
	pool, _ := sqlx.Open("mysql",driverStr)
	pool.SetMaxOpenConns(maxOpen)
	pool.SetMaxIdleConns(maxIdle)
	pool.SetConnMaxLifetime(time.Second * time.Duration(maxLife))
	err := pool.Ping()
	if err != nil {
		log.Fatalf("init db "+name+" error: %v",err)
	}
	p.Pools[name] = pool

	log.Println("DB "+name+" start.")
	return nil
}

func (p *DBPools) GetPool (name string) (*sqlx.DB,bool) {
	pool,find := p.Pools[name]
	return pool,find
}
