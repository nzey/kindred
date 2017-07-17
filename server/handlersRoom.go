package main

import (
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mediocregopher/radix.v2/pool"
)

func roomRemoveHandler(p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		if req.Method == http.MethodPost {
			var r Room
			decoder := json.NewDecoder(req.Body)
			defer req.Body.Close()
			err := decoder.Decode(&r)
			if err != nil {
				panic(err)
			}

			log.Println("room remove is", r)

			out, err := json.Marshal(r)
			if err != nil {
				panic(err)
			}

			log.Println("remove remove out is", string(out))

			conn.Cmd("LREM", "rooms", "-1", string(out))
			w.Write(out)
		}
	})
}

func roomHandler(p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		if req.Method == http.MethodGet {
			var s string
			rl, err := conn.Cmd("LLEN", "rooms").Int()
			r, err := conn.Cmd("LRANGE", "rooms", 0, rl).Array()

			if err != nil {
				panic(err)
			}

			for _, v := range r {
				tempS, _ := v.Str()
				s += tempS + " "
			}

			j, err := json.Marshal(s)
			if err != nil {
				panic(err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(j)
		}

		if req.Method == http.MethodPost {
			var r Room
			decoder := json.NewDecoder(req.Body)
			defer req.Body.Close()
			err := decoder.Decode(&r)
			if err != nil {
				panic(err)
			}

			rc, err := conn.Cmd("GET", "roomCount").Int()
			log.Println("roomCount is:", rc)
			if err != nil {
				panic(err)
			}
			r.RoomNumber = rc + 1

			out, err := json.Marshal(r)
			if err != nil {
				panic(err)
			}

			conn.Cmd("RPUSH", "rooms", string(out))
			conn.Cmd("INCR", "roomCount")

			w.Write(out)
		}
	})
}
