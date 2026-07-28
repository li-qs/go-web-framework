package db

import "myapi/pkg/mysql"

var DB *mysql.DB

func InitMySQL(dsn string) error {
	db, err := mysql.InitDB(dsn)
	if err != nil {
		return err
	}
	DB = db
	return nil
}
