package main

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

//remove once db is created
var dbUsers = map[string]UserBcrypt{}

func init() {
	bs, _ := bcrypt.GenerateFromPassword([]byte("password1234"), bcrypt.MinCost)
	dbUsers["j-s-o"] = UserBcrypt{"j-s-o", "jonathan so", "jso.jonathan@gmail.com", bs}
}

//-----

func signup(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var u User
	var ub UserBcrypt
	defer req.Body.Close()

	err := decoder.Decode(&u)
	if err != nil {
		panic(err)
	}

	bs, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ub = UserBcrypt{u.Username, u.Name, u.Email, bs}
	// db.NewRecord(ub)

	//store userbcrypt details in database
	log.Println("User Bcrypt to store is", ub)
}

func login(w http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {
		// implement JSON web token to see if user is already logged in

		decoder := json.NewDecoder(req.Body)
		var u User
		defer req.Body.Close()
		err := decoder.Decode(&u)
		if err != nil {
			panic(err)
		}

		_, ok := dbUsers[u.Username]
		if !ok {
			http.Error(w, "Username and/or password does not match", http.StatusForbidden)
		}

		err = bcrypt.CompareHashAndPassword(dbUsers[u.Username].Password, []byte(u.Password))
		if err != nil {
			panic(err)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"Username": u.Username,
			"Time":     time.Now().Add(time.Hour * 72).Unix(),
		})
		tokenString, err := token.SignedString(mySigningKey)
		w.Header().Set("Content-Type", "application/json")
		j, _ := json.Marshal(tokenString)
		w.Write(j)
	}
}

func protected(w http.ResponseWriter, req *http.Request) {
	log.Println("Protected resource served")
}
