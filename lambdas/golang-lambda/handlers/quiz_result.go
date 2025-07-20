package handlers

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

type QuizResultResponse struct {
	Email        string           `json:"email"`
	QuizName     string           `json:"quizName"`
	Category     string           `json:"category"`
	CorrectCount int              `json:"correctCount"`
	WrongCount   int              `json:"wrongCount"`
	SkippedCount int              `json:"skippedCount"`
	TotalCount   int              `json:"totalCount"`
	Percentage   float64          `json:"percentage"`
	AttemptedAt  string           `json:"attemptedAt"`
	Results      []QuestionResult `json:"results"`
}

func HandleQuizResult(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	email, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	quizName := request.QueryStringParameters["quizName"]
	if quizName == "" {
		return CreateErrorResponse(400, "Missing 'quizName' parameter"), nil
	}

	db, err := ConnectDB()
	if err != nil {
		log.Printf("❌ Database connection error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer db.Close()

	var response QuizResultResponse
	var resultsJSON json.RawMessage
	response.Email = email

	err = db.QueryRow(`
		SELECT quiz_name, category, correct_count, wrong_count, skipped_count, total_count, percentage, attempted_at, results
		FROM student_quiz_attempts 
		WHERE email = $1 AND quiz_name = $2
	`, email, quizName).Scan(
		&response.QuizName, &response.Category, &response.CorrectCount,
		&response.WrongCount, &response.SkippedCount, &response.TotalCount,
		&response.Percentage, &response.AttemptedAt, &resultsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return CreateErrorResponse(404, "No result found for this quiz"), nil
		}
		log.Printf("❌ Error fetching quiz result: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Parse results JSON
	err = json.Unmarshal(resultsJSON, &response.Results)
	if err != nil {
		log.Printf("❌ Error parsing results JSON: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}