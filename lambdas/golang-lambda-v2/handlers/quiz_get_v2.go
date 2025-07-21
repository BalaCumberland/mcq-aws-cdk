package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func HandleQuizGetByNameV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	email := request.QueryStringParameters["email"]
	quizName := request.QueryStringParameters["quizName"]

	if email == "" {
		return CreateErrorResponse(400, "Missing 'email' parameter"), nil
	}
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}

	email = strings.ToLower(email)
	log.Printf("📌 Fetching quiz questions for: %s, Email: %s", quizName, email)

	// Check student exists and is paid
	student, err := GetStudentFromDynamoDB(email)
	if err != nil {
		log.Printf("❌ Error fetching student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	// Check payment status (skip for now during migration)
	// today := time.Now().Format("2006-01-02")
	// if student.SubExpDate == nil || *student.SubExpDate < today {
	// 	return CreateErrorResponse(400, "Student not paid"), nil
	// }

	// Fetch quiz data and remove correctAnswer from questions
	quiz, err := GetQuizFromDynamoDB(quizName)
	if err != nil {
		log.Printf("❌ Error fetching quiz: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if quiz == nil {
		return CreateErrorResponse(404, "Quiz not found"), nil
	}

	// Remove correctAnswer and explanation from questions
	var cleanQuestions []map[string]interface{}
	for i, question := range quiz.Questions {
		questionMap := map[string]interface{}{
			"qno":        i + 1,
			"question":   question.Question,
			"allAnswers": question.AllAnswers,
		}
		cleanQuestions = append(cleanQuestions, questionMap)
	}

	quizData := map[string]interface{}{
		"quizName":  quiz.QuizName,
		"duration":  quiz.Duration,
		"category":  quiz.Category,
		"questions": cleanQuestions,
	}

	response := map[string]interface{}{
		"message": "Quiz fetched successfully",
		"quiz":    quizData,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}