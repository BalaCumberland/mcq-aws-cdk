package handlers

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

func HandleQuizGetByNameV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get UID from auth context
	userUID, err := GetUserUIDFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}
	
	quizName := request.QueryStringParameters["quizName"]
	className := request.QueryStringParameters["className"]
	subjectName := request.QueryStringParameters["subjectName"]
	topic := request.QueryStringParameters["topic"]

	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}
	if className == "" {
		return CreateErrorResponse(400, "Missing 'className' parameter"), nil
	}
	if subjectName == "" {
		return CreateErrorResponse(400, "Missing 'subjectName' parameter"), nil
	}
	if topic == "" {
		return CreateErrorResponse(400, "Missing 'topic' parameter"), nil
	}

	log.Printf("📌 Fetching quiz questions for: %s (%s-%s-%s), UID: %s", quizName, className, subjectName, topic, userUID)

	// Check student exists and is paid
	student, err := GetStudentInfoByUID(userUID)
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
	quiz, err := GetQuizFromDynamoDB(quizName, className, subjectName, topic)
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
		"quizName":    quiz.QuizName,
		"duration":    quiz.Duration,
		"className":   quiz.ClassName,
		"subjectName": quiz.SubjectName,
		"topic":       quiz.Topic,
		"questions":   cleanQuestions,
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