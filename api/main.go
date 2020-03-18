package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goapt/dotenv"
	"github.com/ilibs/gosql/v2"

	"github.com/fifsky/genstruct/generator"
)

func main() {
	m, err := dotenv.Read()

	if err != nil {
		log.Fatal("load env file error:", err)
	}

	configs := make(map[string]*gosql.Config)
	configs["default"] = &gosql.Config{
		Enable:  true,
		Driver:  "mysql",
		Dsn:     m["database.dsn"],
		ShowSql: false,
	}
	gosql.FatalExit = false
	err = gosql.Connect(configs)

	if err != nil {
		log.Fatal(err)
	}

	db := gosql.Use("default")
	gen := generator.NewGenerator(db)

	http.HandleFunc("/api/struct/gen", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("request body read error \n%s", err)))
			return
		}

		p := &struct {
			Table string   `json:"table"`
			Tags  []string `json:"tags"`
		}{}

		err = json.Unmarshal(body, p)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("request body json Unmarshal error \n%s", err)))
			return
		}

		_, err = db.Exec(p.Table)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("create table error \n%s", err)))
			return
		}

		var tableName string
		defer func() {
			_, err = db.Exec(fmt.Sprintf("drop table `%s`", tableName))
			if err != nil {
				log.Println("drop table error", err)
			}
		}()

		rows := db.QueryRowx("show tables")
		if err != nil {
			w.Write([]byte(fmt.Sprintf("show tables error \n%s", err)))
			return
		}

		err = rows.Scan(&tableName)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("scan table name error \n%s", err)))
			return
		}

		st, err := gen.ShowStruct(tableName, p.Tags)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("generate struct error \n%s", err)))
			return
		}
		w.Write(st)
	})

	err = http.ListenAndServe(":8989", nil)

	if err != nil {
		log.Fatal("ListenAndServe", err)
	}
}