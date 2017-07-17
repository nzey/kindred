package main

import (
	"encoding/json"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mediocregopher/radix.v2/pool"
	"golang.org/x/crypto/bcrypt"
)

func signupHandler(db *gorm.DB, p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}
		var u User
		var un UserAuth

		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		err = decoder.Decode(&u)
		if err != nil {
			panic(err)
		}

		//check if username exists
		db.Where(&UserAuth{Username: u.Username}).First(&un)
		if un.Username != "" {
			http.Error(w, "Username already taken", http.StatusForbidden)
			return
		}

		//check if email exists
		db.Where(&UserAuth{Email: u.Email}).First(&un)
		if un.Email != "" {
			http.Error(w, "Email already taken", http.StatusForbidden)
			return
		}

		//generate encrypted password
		bs, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		//generate user in database
		user := UserAuth{Username: u.Username, Name: u.Name, Email: u.Email, Password: string(bs)}

		db.NewRecord(user)
		db.Create(&user)

		conn.Cmd("HMSET", u.Username, "Survey", "false")

		w.Header().Set("Content-Type", "application/json")
		j, _ := json.Marshal("User created")
		w.Write(j)
		return
	})
}
