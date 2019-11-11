package mysql

import (
	"testing"
	"database/sql"
	"fmt"
)

var (
	USER = "root"
	PASSWORD = "xiaoluo1370"
	ADDR = "127.0.0.1:3306"
	DATABASE = "imdata"
)
func TestDatabase(t *testing.T) {
	conn_addr := fmt.Sprintf("%s:%s@tcp(%s)/%s",USER,PASSWORD,ADDR,DATABASE)

	db, err := sql.Open("mysql", conn_addr)

	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		t.Fatal(err.Error()) // proper error handling instead of panic in your app
	}
}
