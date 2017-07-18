package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mediocregopher/radix.v2/pool"
)

func qotdHandler(db *gorm.DB, p *pool.Pool, qotdCounter *int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		conn, err := p.Get()

		defer p.Put(conn)
		if err != nil {
			panic(err)
		}

		if req.Method == http.MethodGet {
			param := req.URL.Query().Get("user")

			if param != "" {
				var allUserAnswers []QotdAnswer
				userAnswers := make([]UserAnswer, 10)
				userid, err := strconv.Atoi(param)
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
				q, err := json.Marshal(userAnswers)
				if err != nil {
					fmt.Println(err)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(q)

				// if query string specifies data (i.e. api/qotd?q=data),
				// respond with all user answers from the last 10 QOTDs
			} else if req.URL.Query().Get("q") == "data" {
				var qotdData []QotdData

				db.Raw("SELECT qotds.id as qotd_id, qotds.question_type AS qotd_type, qotds.category AS qotd_category, qotds.text AS qotd_text, qotd_answers.user_auth_id, qotd_answers.text AS answer_text, user_profiles.zip, user_profiles.age, user_profiles.gender, user_profiles.income, user_profiles.education, user_profiles.religiousity, user_profiles.religion, user_profiles.ethnicity, user_profiles.state, user_profiles.party FROM qotds, qotd_answer_options, qotd_answers, user_profiles WHERE qotds.id = qotd_answer_options.qotd_id AND qotds.id = qotd_answers.qotd_id AND qotd_answer_options.text = qotd_answers.text AND qotd_answers.user_auth_id = user_profiles.user_auth_id AND qotds.id <=? AND qotds.id >=?", qotdCounter, *qotdCounter-9).Scan(&qotdData)

				data, err := json.Marshal(qotdData)
				if err != nil {
					panic(err)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(data)

				// if query string specifies dataoptions (i.e. api/qotd?q=dataoptions),
				// respond with answer options for last 10 qotds
			} else if req.URL.Query().Get("q") == "dataoptions" {
				var qotdAnswerOptions []QotdAnswers

				db.Raw("SELECT qotds.text AS qotd_text, qotds.ID AS qotd_id, qotd_answer_options.text AS answer_text FROM qotds, qotd_answer_options WHERE qotds.id = qotd_answer_options.qotd_id AND qotd_id <=? AND qotd_id >=?", qotdCounter, *qotdCounter-9).Scan(&qotdAnswerOptions)

				data, err := json.Marshal(qotdAnswerOptions)
				if err != nil {
					panic(err)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(data)

				// if query string specifies qotd (i.e. api/qotd?q=qotd),
				// respond with today's qotd
			} else if req.URL.Query().Get("q") == "qotd" {
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
				data, err := json.Marshal(qotdWOptions)
				if err != nil {
					panic(err)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write(data)
			}

			// if POST request, add user's qotd answer to db
		} else {
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
			w.Header().Set("Content-Type", "application/json")
			response, _ := json.Marshal(responseString)
			w.Write(response)
		}
	})
}

//----- TWILIO REVERSE PROXY ------//

func twilioProxy(w http.ResponseWriter, r *http.Request) {
	log.Println("receiving twilio request from client")
	r.Host = "localhost:3000"

	j := strings.Join(r.URL.Query()["q"], "")
	u, _ := url.Parse("http://localhost:300/api/twilio" + j)
	proxy := httputil.NewSingleHostReverseProxy(u)

	proxy.Transport = &transport{CapturedTransport: http.DefaultTransport}
	proxy.ServeHTTP(w, r)
}

type transport struct {
	CapturedTransport http.RoundTripper
}

func (t *transport) RoundTrip(request *http.Request) (*http.Response, error) {
	// response, err := http.DefaultTransport.RoundTrip(request)
	response, err := t.CapturedTransport.RoundTrip(request)
	bodyBytes, err := ioutil.ReadAll(response.Body)

	// body, err := httputil.DumpResponse(response, true)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	log.Println("proxy reponse is", response)

	return response, err
}
