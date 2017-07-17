package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func feedbackHandler(db *gorm.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			var randQuestion FeedbackQuestion
			var questionCount int
			db.Table("feedback_questions").Count(&questionCount)
			// TODO: only do next lines if question count is more than 0
			db.Find(&randQuestion, rand.Intn(questionCount)+1)
			q, err := json.Marshal(randQuestion)
			if err != nil {
				fmt.Println(err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(q)
		} else {
			// post feedback answer
			fmt.Println("feedback post")
			var newAnswer FeedbackAnswer
			var user UserAuth
			var question FeedbackQuestion
			decoder := json.NewDecoder(req.Body)
			defer req.Body.Close()
			err := decoder.Decode(&newAnswer)
			if err != nil {
				panic(err)
			}
			db.Model(&newAnswer).Related(&user)
			db.Model(&newAnswer).Related(&question)
			if user.ID != 0 && question.ID != 0 {
				db.NewRecord(newAnswer)
				db.Create(&newAnswer).Create(&newAnswer)
			}
		}
	})
}
