package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var (
	host string
	user string
	pass string
)

func init() {
	flag.StringVar(&host, "host", "127.0.0.1:3306", "DB host")
	flag.StringVar(&user, "user", "golang", "DB user")
	flag.StringVar(&pass, "pass", "golang", "DB pass")
	flag.Parse()
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("failed to create new logger")
		os.Exit(1)
	}
	defer logger.Sync()

	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/sample", user, pass, host)
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		logger.Fatal("failed to connect to mysql", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("failed to establish DB connection", zap.Error(err))
	}

	rows, err := db.Query("select id, name from sample_table")
	if err != nil {
		logger.Fatal("failed to execute select statement", zap.Error(err))
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			logger.Fatal("error occured while fetching data", zap.Error(err))
		}
		fmt.Printf("%d\t%s\n", id, name)
	}
	if err := rows.Err(); err != nil {
		logger.Fatal("error", zap.Error(err))
	}
}
