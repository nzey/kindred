package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mediocregopher/radix.v2/pool"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var upgrader = websocket.Upgrader{}

func main() {
	//qotdCounter for 24hr worker
	var qotdCounter int
	qotdCounter = 10

	//redis
	p, err := pool.New("tcp", "localhost:6379", 10)
	if err != nil {
		log.Panic(err)
	}
	conn, err := p.Get()
	if err != nil {
		panic(err)
	}
	conn.Cmd("INCR", "roomCount")
	p.Put(conn)

	//db
	var db *gorm.DB
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_USER, DB_PASSWORD, DB_NAME)
	db, err = gorm.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&UserAuth{}, &UserProfile{}, &QotdAnswer{}, &FeedbackQuestion{}, &FeedbackAnswer{}, &Kinship{}, &Chat{})

	seedQotds(db)
	seedUsers(db, p, 500) // mock data

	defer db.Close()

	//routes
	http.Handle("/", http.FileServer(http.Dir("../public/")))
	http.Handle("/public/assets/", http.StripPrefix("/public/assets/", http.FileServer(http.Dir("../public/assets/"))))
	http.Handle("/bundles/", http.StripPrefix("/bundles/", http.FileServer(http.Dir("../bundles/"))))
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/api/profile", profileHandler(db, p))
	http.Handle("/api/signup", signupHandler(db, p))
	http.Handle("/api/login", loginHandler(db, p))
	http.Handle("/api/logout", logoutHandler(p))
	http.Handle("/api/tokenCheck", tokenHandler(p))
	http.Handle("/api/twilio", http.HandlerFunc(twilioProxy))
	http.Handle("/api/feedback", feedbackHandler(db))
	http.Handle("/api/visitCheck", visitHandler(p))
	http.Handle("/api/queue", queueHandler(p))
	http.Handle("/api/queueRemove", queueRemoveHandler(p))
	http.Handle("/api/room", roomHandler(p))
	http.Handle("/api/roomRemove", roomRemoveHandler(p))
	http.Handle("/api/qotd", qotdHandler(db, p, &qotdCounter))

	go worker(db, p, &qotdCounter)

	//Initialize
	//if on localhost, use ListenAndServe, if on deployment server, use ListenAndServeTLS.
	http.ListenAndServe(":8080", nil)
	// err = http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/www.kindredchat.io/fullchain.pem", "/etc/letsencrypt/live/www.kindredchat.io/privkey.pem", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
