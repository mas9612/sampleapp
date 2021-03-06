package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var (
	host   string
	user   string
	pass   string
	dbname string
)

const (
	window = 10
)

func init() {
	flag.StringVar(&host, "host", "127.0.0.1:3306", "DB host")
	flag.StringVar(&user, "user", "golang", "DB user")
	flag.StringVar(&pass, "pass", "golang", "DB pass")
	flag.StringVar(&dbname, "dbname", "sample", "DB name")
	flag.Parse()

	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		fmt.Println("failed to create new logger")
		os.Exit(1)
	}

	// If environment variable set, command line argument is overridden by it.
	if value, ok := os.LookupEnv("SAMPLEAPP_DB_HOST"); ok {
		logger.Info("configuration overridden by environment variable 'SAMPLEAPP_DB_HOST'")
		host = value
	}
	if value, ok := os.LookupEnv("SAMPLEAPP_DB_USER"); ok {
		logger.Info("configuration overridden by environment variable 'SAMPLEAPP_DB_USER'")
		user = value
	}
	if value, ok := os.LookupEnv("SAMPLEAPP_DB_PASS"); ok {
		logger.Info("configuration overridden by environment variable 'SAMPLEAPP_DB_PASS'")
		pass = value
	}
	if value, ok := os.LookupEnv("SAMPLEAPP_DB_NAME"); ok {
		logger.Info("configuration overridden by environment variable 'SAMPLEAPP_DB_NAME'")
		dbname = value
	}
}

var (
	logger *zap.Logger

	istioTracingHeaders = []string{
		"x-request-id",
		"x-b3-traceid",
		"x-b3-spanid",
		"x-b3-parentspanid",
		"x-b3-sampled",
		"x-b3-flags",
		"x-ot-span-context",
	}
)

func main() {
	defer logger.Sync()
	http.HandleFunc("/", indexHandler)

	logger.Info("listening on 0.0.0.0:8080")
	err := http.ListenAndServe(":8080", nil)
	logger.Fatal("failed to serve HTTP server", zap.Error(err))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, dbname)
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		logger.Fatal("failed to connect to mysql", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("failed to establish DB connection", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("select id, name from sample_table")
	if err != nil {
		logger.Error("failed to execute select statement", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	i := 0
	result := make([]string, window)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			logger.Error("error occured while fetching data", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		result[i] = fmt.Sprintf("%d\t%s\n", id, name)
		i++
		if i >= window {
			break
		}
	}
	if err := rows.Err(); err != nil {
		logger.Error("error occured while fetching data", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, header := range istioTracingHeaders {
		w.Header().Add(header, r.Header.Get(header))
	}
	w.WriteHeader(http.StatusOK)
	for _, r := range result {
		if n, err := io.WriteString(w, r); err != nil || n < len(r) {
			logger.Error("error occured while sending data", zap.Error(err))
			break
		}
	}
}
