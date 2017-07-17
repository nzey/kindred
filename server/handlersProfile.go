package main

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mediocregopher/radix.v2/pool"
	"net/http"
)

func profileHandler(db *gorm.DB, p *pool.Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()
		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		//handle post
		if req.Method == http.MethodPost {
			var us UserSurvey
			var un UserAuth
			var usp UserProfile

			rh := req.Header.Get("Authorization")[7:]

			decoder := json.NewDecoder(req.Body)
			defer req.Body.Close()
			err := decoder.Decode(&us)
			if err != nil {
				panic(err)
			}

			db.Where(&UserAuth{Token: "\"" + rh + "\""}).First(&un)
			us.ID = un.ID
			db.Model(&un).Related(&usp)

			//create new user profile entry
			if usp.UserAuthID == 0 {
				f := defaultSurvey(us)
				//create new database record
				db.NewRecord(f)
				db.Create(&f)
				//add profile to cache
				out, err := json.Marshal(f)
				if err != nil {
					panic(err)
				}
				conn.Cmd("HMSET", un.Username, "Profile", string(out), "Survey", "true")

				//write response back
				w.Header().Set("Content-Type", "application/json")
				j, _ := json.Marshal("Profile posted")
				w.Write(j)
			} else {
				//update existing user profile entry in db
				f := defaultSurvey(us)
				db.Model(&usp).Updates(f)

				//updata profile in cache
				out, err := json.Marshal(f)
				if err != nil {
					panic(err)
				}

				conn.Cmd("HMSET", un.Username, "Profile", string(out), "Survey", "true")

				//write response back
				w.Header().Set("Content-Type", "application/json")
				j, _ := json.Marshal("Profile updated")
				w.Write(j)
			}
		}

		//handle get
		if req.Method == http.MethodGet {
			u := req.URL.Query()

			res, err := conn.Cmd("HGET", u["q"], "Profile").Str()
			if err != nil {
				panic(err)
			}

			w.Header().Set("Content-Type", "application/json")
			j, _ := json.Marshal(res)
			w.Write(j)
		}

		// delete user profile
		if req.Method == http.MethodDelete {
			///api/profile?user=username
			type UserID struct {
				ID int
			}
			var userID UserID
			var userProfile UserProfile
			var userAuth UserAuth
			var userKinship Kinship
			var r string
			username := req.URL.Query().Get("user")
			db.Table("user_auths").Select("id").Where("username = ?", username).Scan(&userID)
			db.Raw("SELECT * FROM user_auths WHERE id = ?", userID.ID).Scan(&userAuth)
			db.Raw("SELECT * FROM user_profiles WHERE user_auth_id = ?", userID.ID).Scan(&userProfile)
			db.Raw("SELECT * FROM kinships WHERE user_auth_id = ?", userID.ID).Scan(&userKinship)
			if userProfile.UserAuthID != 0 {
				db.Delete(&userProfile)
				db.Delete(&userAuth)
				db.Delete(&userKinship)
				r = "Profile deleted"
			} else {
				r = "User does not exist"
			}
			res, err := json.Marshal(r)
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(res)
		}
	})
}
