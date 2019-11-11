package mysql

import (
	"github.com/luobangkui/golang-IM/utils/log"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"github.com/luobangkui/golang-IM/config"
)

func NewDatabase(config config.Config) *sql.DB {
	db := config.Db
	return newDb(db.Addr,db.User,db.Pass,db.Database)
}



func newDb(addr string, user string, pass string,database string ) *sql.DB  {

	conn_addr := fmt.Sprintf("%s:%s@tcp(%s)/%s",user,pass,addr,database)
	db, err := sql.Open("mysql", conn_addr)
	if err != nil {
		log.Fatal(err)
	}

	return db
}






