package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
)

var (
	dsn string
)

func main() {
	var port int
	flag.IntVar(&port, "port", 5000, "http listen port number")
	flag.StringVar(&dsn, "dsn", "root:@tcp(127.0.0.1:3306)/?charset=utf8", "MySQL DSN")
	flag.Parse()
	log.Printf("Listing port %d", port)
	log.Printf("dsn %s", dsn)

	http.HandleFunc("/", handler)
	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", dsn)
	defer db.Close()

	if err != nil {
		serverError(w, err)
		return
	}
	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		serverError(w, err)
		return
	}
	if !rows.Next() {
		serverError(w, errors.New("No slave status"))
		return
	}
	defer rows.Close()

	// カラム数と同じ要素数のsliceを用意して sql.RawBytes のポインタで初期化しておく
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	for i, _ := range values {
		var v sql.RawBytes
		values[i] = &v
	}

	err = rows.Scan(values...)
	if err != nil {
		serverError(w, err)
		return
	}

	// 結果を返す用のmapに値を詰める
	slaveInfo := make(map[string]interface{})
	for i, name := range columns {
		bp := values[i].(*sql.RawBytes)
		vs := string(*bp)
		vi, err := strconv.ParseInt(vs, 10, 64)
		if err != nil {
			slaveInfo[name] = vs
		} else {
			slaveInfo[name] = vi
		}
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(slaveInfo)
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("error: %s", err)
	code := http.StatusInternalServerError
	http.Error(w, fmt.Sprintf("%s", err), code)
}
