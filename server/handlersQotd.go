package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mediocregopher/radix.v2/pool"
)

func lastXAnswersFromOneUser(db *gorm.DB, userIdStr string, x int) []UserAnswer {
	var allUserAnswers []QotdAnswer
	userAnswers := make([]UserAnswer, 10)
	userid, err := strconv.Atoi(userIdStr)
	if err != nil {
		fmt.Println(err)
	}
	db.Where("user_auth_id = ?", userid).Find(&allUserAnswers)
	numOfAnswers := len(allUserAnswers)

	if numOfAnswers > 10 {
		allUserAnswers = allUserAnswers[0:10]
	}
	if numOfAnswers < 10 {
		userAnswers = userAnswers[0:numOfAnswers]
	}

	for i, answer := range allUserAnswers {
		var question Qotd
		db.Model(&answer).Related(&question)
		userAnswers[i] = UserAnswer{question.Text, answer.Text}
	}
	return userAnswers
}

func lastXAnswersFromAllUsers(db *gorm.DB, qotdCounter *int, x int) []QotdData {
	var qotdData []QotdData
	db.Raw("SELECT qotds.id as qotd_id, qotds.question_type AS qotd_type, qotds.category AS qotd_category, qotds.text AS qotd_text, qotd_answers.user_auth_id, qotd_answers.text AS answer_text, user_profiles.zip, user_profiles.age, user_profiles.gender, user_profiles.income, user_profiles.education, user_profiles.religiousity, user_profiles.religion, user_profiles.ethnicity, user_profiles.state, user_profiles.party FROM qotds, qotd_answer_options, qotd_answers, user_profiles WHERE qotds.id = qotd_answer_options.qotd_id AND qotds.id = qotd_answers.qotd_id AND qotd_answer_options.text = qotd_answers.text AND qotd_answers.user_auth_id = user_profiles.user_auth_id AND qotds.id <=? AND qotds.id >=?", qotdCounter, *qotdCounter-9).Scan(&qotdData)
	return qotdData
}

func lastXQuestionsAndAnswerOptions(db *gorm.DB, qotdCounter *int, x int) []QotdAnswers {
	var qotdAnswerOptions []QotdAnswers
	db.Raw("SELECT qotds.text AS qotd_text, qotds.ID AS qotd_id, qotd_answer_options.text AS answer_text FROM qotds, qotd_answer_options WHERE qotds.id = qotd_answer_options.qotd_id AND qotd_id <=? AND qotd_id >=?", qotdCounter, *qotdCounter-9).Scan(&qotdAnswerOptions)
	return qotdAnswerOptions
}

func getQotd(p *pool.Pool) QuestionWOptions {
	conn, err := p.Get()
	defer p.Put(conn)
	if err != nil {
		panic(err)
	}

	var qotdWOptions QuestionWOptions
	qotd, err := conn.Cmd("HGETALL", "qotd").Map()
	options, err := conn.Cmd("HGETALL", "options").List()
	qotdWOptions.ID = qotd["id"]
	qotdWOptions.Qtype = qotd["qtype"]
	qotdWOptions.Text = qotd["text"]
	qotdWOptions.Category = qotd["category"]
	for i := 1; i < len(options); i += 2 {
		qotdWOptions.Options = append(qotdWOptions.Options, options[i])
	}

	return qotdWOptions
}

func addOrUpdateQotdAnswer(db *gorm.DB, req *http.Request) string {
	var newAnswer QotdAnswer
	var oldAnswer QotdAnswer
	var user UserAuth
	var questionAnswered Qotd
	var responseString string

	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	err := decoder.Decode(&newAnswer)
	if err != nil {
		panic(err)
	}
	db.Model(&newAnswer).Related(&questionAnswered)
	db.Model(&newAnswer).Related(&user)
	if questionAnswered.ID != 0 && user.ID != 0 {
		db.Where(&QotdAnswer{UserAuthID: user.ID, QotdID: questionAnswered.ID}).First(&oldAnswer)
		if &oldAnswer == nil {
			db.NewRecord(newAnswer)
			db.Create(&newAnswer)
		} else {
			db.Model(&oldAnswer).Update("text", newAnswer.Text)
		}
		responseString = "Answer successfully posted or updated in db"
	} else {
		responseString = "Failed to post answer. Incorrect user or question id."
	}
	return responseString
}

func writeResponse(w http.ResponseWriter, unencodedJson interface{}) {
	q, err := json.Marshal(unencodedJson)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(q)
}

func qotdHandler(db *gorm.DB, p *pool.Pool, qotdCounter *int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		if req.Method == http.MethodGet {
			user := req.URL.Query().Get("user")
			if user != "" {
				writeResponse(w, lastXAnswersFromOneUser(db, user, 10))
			} else if req.URL.Query().Get("q") == "data" {
				writeResponse(w, lastXAnswersFromAllUsers(db, qotdCounter, 10))
			} else if req.URL.Query().Get("q") == "dataoptions" {
				writeResponse(w, lastXQuestionsAndAnswerOptions(db, qotdCounter, 10))
			} else if req.URL.Query().Get("q") == "qotd" {
				writeResponse(w, getQotd(p))
			}
		} else if req.Method == http.MethodPost {
			writeResponse(w, addOrUpdateQotdAnswer(db, req))
		}
	})
}
