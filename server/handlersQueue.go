package main

import (
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mediocregopher/radix.v2/pool"
)

// ----- QUEUE ------ //

func queueRemoveHandler(p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		if req.Method == http.MethodPost {
			var p UserQueue
			decoder := json.NewDecoder(req.Body)
			defer req.Body.Close()
			err := decoder.Decode(&p)
			if err != nil {
				panic(err)
			}

			log.Println("post to queue from queue remove is:", p)

			out, err := json.Marshal(p)
			if err != nil {
				panic(err)
			}

			conn.Cmd("LREM", "queue", "-1", string(out))
			w.Write(out)
		}
	})
}

func queueHandler(p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		if req.Method == http.MethodGet {

			qr, _ := conn.Cmd("LPOP", "queue").Str()
			log.Printf("qr is: v - %v, t - %T", qr, qr)
			if qr == "" {
				log.Println("handling empty queue")
				j, err := json.Marshal("empty")
				if err != nil {
					panic(err)
				}
				w.Write(j)
			} else {
				j, err := json.Marshal(qr)
				if err != nil {
					panic(err)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(j)
			}
		}

		if req.Method == http.MethodPost {
			var p UserQueue
			decoder := json.NewDecoder(req.Body)
			defer req.Body.Close()
			err := decoder.Decode(&p)
			if err != nil {
				panic(err)
			}

			log.Println("post to queue is from queue handler is:", p)

			out, err := json.Marshal(p)
			if err != nil {
				panic(err)
			}

			conn.Cmd("RPUSH", "queue", string(out))
			w.Write(out)
		}

	})
}
