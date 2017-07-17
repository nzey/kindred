package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mediocregopher/radix.v2/pool"
	"golang.org/x/crypto/bcrypt"
)

//----- LOGIN -----//

func loginHandler(db *gorm.DB, p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		if req.Method == http.MethodPost {
			var u User
			var un UserAuth
			var usp UserProfile

			decoder := json.NewDecoder(req.Body)
			defer req.Body.Close()
			err := decoder.Decode(&u)
			if err != nil {
				panic(err)
			}

			//check if username is valid
			db.Where(&UserAuth{Username: u.Username}).First(&un)
			log.Println(un.Username)
			if un.Username == "" {
				http.Error(w, "Username or password does not match", http.StatusForbidden)
				return
			}

			//compare passwords
			err = bcrypt.CompareHashAndPassword([]byte(un.Password), []byte(u.Password))

			if err != nil {
				http.Error(w, "Username or password does not match", http.StatusForbidden)
				return
			}

			//issue token upon successful login
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"username": u.Username,
				"iss":      "https://kindredchat.io",
				"exp":      time.Now().Add(time.Hour * 72).Unix(),
			})

			//sign token
			tokenString, err := token.SignedString(mySigningKey)

			//store in db and set header
			w.Header().Set("Content-Type", "application/json")
			j, _ := json.Marshal(tokenString)
			db.Model(&un).Update("Token", j)

			//store token and user profile in cache
			rj := j[1 : len(j)-1]
			db.Where("user_auth_id = ?", un.ID).First(&usp)
			out, err := json.Marshal(usp)
			if err != nil {
				panic(err)
			}

			conn.Cmd("HMSET", u.Username, "Token", rj, "Name", un.Name, "Profile", string(out))
			//send token back as response
			w.Write(j)
		}
	})
}

//----- LOGOUT -----//

func logoutHandler(p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		if req.Method == http.MethodPost {
			var c Cookie

			decoder := json.NewDecoder(req.Body)
			defer req.Body.Close()
			err := decoder.Decode(&c)
			if err != nil {
				panic(err)
			}

			conn.Cmd("HDEL", c.Username, "Token")

			j, err := json.Marshal("User logged out")
			w.Write(j)
		}
	})
}

//----- CHECK TOKEN -----//
func tokenHandler(p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		var c Cookie
		var b bool

		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		err = decoder.Decode(&c)
		if err != nil {
			panic(err)
		}

		if req.Method == http.MethodPost {
			res, err := conn.Cmd("HGET", c.Username, "Token").Str()
			if err != nil {
				panic(err)
			}
			b = c.Token == res

			w.Header().Set("Content-Type", "application/json")
			j, _ := json.Marshal(b)
			w.Write(j)
		}
	})
}
