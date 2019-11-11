package server

import (
	"database/sql"
	"log"
)

type UserPass struct {
	Username string `json:"username"`
	Password string `json:"password"`
}




func ValidateUserPass(db *sql.DB, user, pass string )( string ,error ){
	stmtOut, err := db.Prepare("SELECT uid FROM IM_USER WHERE username = ? AND password = ?")
	if err != nil {
		log.Println(err)
		return "-1" ,err
	}
	defer stmtOut.Close()

	var uid string

	err = stmtOut.QueryRow(user,pass).Scan(&uid)
	log.Println(user,pass,uid)

	return uid, nil
}






