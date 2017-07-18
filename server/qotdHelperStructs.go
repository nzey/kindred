package main

// If the type you're looking for is not here,
// check server/db.go, where schema types are defined

type UserAnswer struct {
	Question string
	Answer   string
}

type QotdData struct {
	QotdID       int
	QotdType     string
	QotdCategory string
	QotdText     string
	UserAuthID   int
	AnswerText   string
	Zip          int
	Age          int
	Gender       int
	Income       int
	Education    int
	Religiousity int
	Ethnicity    int
	State        string
	Party        int
}

type QotdAnswers struct {
	QotdID     int
	QotdText   string
	AnswerText string
}

type QuestionWOptions struct {
	ID       string
	Qtype    string
	Category string
	Text     string
	Options  []string
}
