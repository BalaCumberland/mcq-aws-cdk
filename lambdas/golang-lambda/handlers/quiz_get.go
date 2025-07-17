package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleQuizGetByName(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	db, err := ConnectDB()
	if err != nil {
		log.Printf("❌ Database connection error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer db.Close()

	// Check student exists and is paid
	studentQuery := `
		SELECT id, email, name, student_class, phone_number, sub_exp_date, updated_by, amount, payment_time 
		FROM students 
		WHERE LOWER(email) = LOWER($1)`

	var student struct {
		ID          int
		Email       string
		Name        string
		Class       string
		Phone       string
		SubExpDate  sql.NullString
		UpdatedBy   sql.NullString
		Amount      sql.NullFloat64
		PaymentTime sql.NullString
	}

	err = db.QueryRow(studentQuery, email).Scan(
		&student.ID, &student.Email, &student.Name, &student.Class,
		&student.Phone, &student.SubExpDate, &student.UpdatedBy,
		&student.Amount, &student.PaymentTime)

	if err != nil {
		if err == sql.ErrNoRows {
			return CreateErrorResponse(404, "Student not found"), nil
		}
		log.Printf("❌ Database error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Check payment status
	today := time.Now().Format("2006-01-02")
	if !student.SubExpDate.Valid || student.SubExpDate.String < today {
		return CreateErrorResponse(400, "Student not paid"), nil
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("❌ Failed to begin transaction: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer tx.Rollback()

	// Fetch quiz data and remove correctAnswer from questions
	var quizData struct {
		QuizName  string                   `json:"quizName"`
		Duration  int                      `json:"duration"`
		Category  string                   `json:"category"`
		Questions []map[string]interface{} `json:"questions"`
	}

	var questionsJSON json.RawMessage
	err = tx.QueryRow(
		`SELECT quiz_name AS "quizName", duration, category, questions FROM quiz_questions WHERE quiz_name = $1`,
		quizName).Scan(&quizData.QuizName, &quizData.Duration, &quizData.Category, &questionsJSON)

	if err == nil {
		// Parse questions and remove correctAnswer
		var questions []map[string]interface{}
		json.Unmarshal(questionsJSON, &questions)

		for i := range questions {
			delete(questions[i], "correctAnswer")
			delete(questions[i], "explanation")
		}
		quizData.Questions = questions
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return CreateErrorResponse(404, fmt.Sprintf("Quiz not found: %s", quizName)), nil
		}
		log.Printf("❌ Database error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// // Update student_quizzes table
	// quizUpdateQuery := `
	// 	INSERT INTO student_quizzes (email, quiz_names)
	// 	VALUES ($1, to_jsonb(ARRAY[$2]::text[]))
	// 	ON CONFLICT (email)
	// 	DO UPDATE SET quiz_names = (
	// 		SELECT jsonb_agg(DISTINCT q)
	// 		FROM jsonb_array_elements(
	// 			COALESCE(student_quizzes.quiz_names, '[]'::jsonb) || to_jsonb(ARRAY[$2]::text[])
	// 		) AS q
	// 	)
	// 	RETURNING quiz_names`

	// var updatedQuizNames json.RawMessage
	// err = tx.QueryRow(quizUpdateQuery, email, quizName).Scan(&updatedQuizNames)
	// if err != nil {
	// 	log.Printf("❌ Failed to update student_quizzes: %v", err)
	// 	return CreateErrorResponse(500, "Internal Server Error"), nil
	// }

	// log.Printf("✅ Updated student_quizzes: %s", string(updatedQuizNames))

	// // Commit transaction
	// err = tx.Commit()
	// if err != nil {
	// 	log.Printf("❌ Failed to commit transaction: %v", err)
	// 	return CreateErrorResponse(500, "Internal Server Error"), nil
	// }

	response := map[string]interface{}{
		"message": "Quiz fetched and updated successfully",
		"quiz":    quizData,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}
