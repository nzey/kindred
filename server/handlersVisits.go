package main

import (
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mediocregopher/radix.v2/pool"
)

// ----- CHECK VISITS (first time users) -----//

func visitHandler(p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		if req.Method == http.MethodGet {
			u := req.URL.Query()
			log.Println("uri is", u["q"])
			res, err := conn.Cmd("HGET", u["q"], "Survey").Str()
			if err != nil {
				panic(err)
			}

			w.Header().Set("Content-Type", "application/json")
			j, _ := json.Marshal(res)
			w.Write(j)
		}
	})
}
