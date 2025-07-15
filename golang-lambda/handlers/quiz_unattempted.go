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

var dateFilteredCategories = map[string]bool{
	"CLS11-MPC-EAMCET": true, "CLS11-MPC-JEEMAINS": true, "CLS11-MPC-JEEADV": true,
	"CLS12-MPC-EAMCET": true, "CLS12-MPC-JEEMAINS": true, "CLS12-MPC-JEEADV": true,
	"CLS11-BIPC-EAPCET": true, "CLS11-BIPC-NEET": true,
	"CLS12-BIPC-EAPCET": true, "CLS12-BIPC-NEET": true,
}

func HandleUnattemptedQuizzes(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	category := request.QueryStringParameters["category"]
	studentEmail := request.QueryStringParameters["email"]

	if category == "" || studentEmail == "" {
		return CreateErrorResponse(400, "Category and student email are required"), nil
	}

	studentEmail = strings.ToLower(studentEmail)
	log.Printf("📌 Fetching quizzes for category: %s, excluding attempted quizzes for: %s", category, studentEmail)

	db, err := ConnectDB()
	if err != nil {
		log.Printf("❌ Database connection error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer db.Close()

	// Check if student exists and is paid
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

	err = db.QueryRow(studentQuery, studentEmail).Scan(
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

	// Build quiz filter query
	quizFilterQuery := `SELECT quiz_name FROM quiz_questions WHERE category = $1`
	queryParams := []interface{}{category}

	// Apply date filtering if category matches
	if dateFilteredCategories[category] {
		now := time.Now()
		currentMonth := int(now.Month())
		currentDate := now.Day()

		quizPattern := fmt.Sprintf("%s-%d-%d-%%", category, currentMonth, currentDate)
		quizFilterQuery = `SELECT quiz_name FROM quiz_questions WHERE category = $1 AND quiz_name LIKE $2`
		queryParams = append(queryParams, quizPattern)
	}

	log.Printf("🔍 Executing Query: %s", quizFilterQuery)
	log.Printf("🔍 Query Parameters: %v", queryParams)

	// Get all quizzes for category
	allQuizzesRows, err := db.Query(quizFilterQuery, queryParams...)
	if err != nil {
		log.Printf("❌ Error fetching quizzes: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer allQuizzesRows.Close()

	var allQuizNames []string
	for allQuizzesRows.Next() {
		var quizName string
		if err := allQuizzesRows.Scan(&quizName); err != nil {
			log.Printf("❌ Error scanning quiz name: %v", err)
			continue
		}
		allQuizNames = append(allQuizNames, quizName)
	}

	if len(allQuizNames) == 0 {
		response := map[string]interface{}{
			"unattempted_quizzes": []string{},
		}
		responseJSON, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    GetCORSHeaders(),
			Body:       string(responseJSON),
		}, nil
	}

	// Get attempted quizzes
	attemptedQuizzesRows, err := db.Query(
		`SELECT jsonb_array_elements_text(quiz_names) AS quiz_name FROM student_quizzes WHERE LOWER(email) = $1`,
		studentEmail)
	if err != nil {
		log.Printf("❌ Error fetching attempted quizzes: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer attemptedQuizzesRows.Close()

	var attemptedQuizNames []string
	for attemptedQuizzesRows.Next() {
		var quizName string
		if err := attemptedQuizzesRows.Scan(&quizName); err != nil {
			log.Printf("❌ Error scanning attempted quiz name: %v", err)
			continue
		}
		attemptedQuizNames = append(attemptedQuizNames, quizName)
	}

	log.Printf("✅ Attempted quizzes: %v", attemptedQuizNames)

	// If no attempted quizzes, return all quizzes
	if len(attemptedQuizNames) == 0 {
		log.Printf("✅ No attempted quizzes found for %s. Returning all quizzes.", studentEmail)
		response := map[string]interface{}{
			"unattempted_quizzes": allQuizNames,
		}
		responseJSON, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    GetCORSHeaders(),
			Body:       string(responseJSON),
		}, nil
	}

	// Filter out attempted quizzes
	attemptedSet := make(map[string]bool)
	for _, name := range attemptedQuizNames {
		attemptedSet[name] = true
	}

	var unattemptedQuizzes []string
	for _, name := range allQuizNames {
		if !attemptedSet[name] {
			unattemptedQuizzes = append(unattemptedQuizzes, name)
		}
	}

	response := map[string]interface{}{
		"unattempted_quizzes": unattemptedQuizzes,
	}
	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}