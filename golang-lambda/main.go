package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"



	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	_ "github.com/lib/pq"
	"github.com/xuri/excelize/v2"
)



func getUserFromContext(request events.APIGatewayProxyRequest) (string, error) {
	// Get user email from authorizer context
	if request.RequestContext.Authorizer == nil {
		return "", fmt.Errorf("no authorizer context")
	}
	email, ok := request.RequestContext.Authorizer["email"].(string)
	if !ok || email == "" {
		return "", fmt.Errorf("missing user email from authorizer")
	}
	return email, nil
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

func getDBConfig() (*DBConfig, error) {
	log.Printf("🔐 Getting DB config from Secrets Manager...")
	sess := session.Must(session.NewSession())
	svc := secretsmanager.New(sess)

	log.Printf("📡 Calling GetSecretValue...")
	result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String("dbcredentials/postgres"),
	})
	if err != nil {
		log.Printf("❌ Failed to get secret: %v", err)
		return nil, err
	}

	log.Printf("✅ Secret retrieved successfully")
	var config DBConfig
	err = json.Unmarshal([]byte(*result.SecretString), &config)
	if err != nil {
		log.Printf("❌ Failed to parse secret JSON: %v", err)
		return nil, err
	}
	log.Printf("✅ DB config parsed successfully")
	return &config, nil
}

// ✅ Structs
type QuizData struct {
	QuizName  string     `json:"quizName"`
	Duration  int        `json:"duration"`
	Category  string     `json:"category"`
	Questions []Question `json:"questions"`
}

type Question struct {
	Explanation      string `json:"explanation"`
	Question         string `json:"question"`
	CorrectAnswer    string `json:"correctAnswer"`
	IncorrectAnswers string `json:"incorrectAnswers"`
}

type StudentUpdateRequest struct {
	Email        string   `json:"email"`
	PhoneNumber  *string  `json:"phoneNumber,omitempty"`
	Name         *string  `json:"name,omitempty"`
	StudentClass *string  `json:"studentClass,omitempty"`
	Amount       *float64 `json:"amount,omitempty"`
	UpdatedBy    *string  `json:"updatedBy,omitempty"`
}

// ✅ Connect to PostgreSQL
func connectDB() (*sql.DB, error) {
	log.Printf("🔌 Connecting to database...")
	config, err := getDBConfig()
	if err != nil {
		log.Printf("❌ Failed to get DB config: %v", err)
		return nil, err
	}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		config.Host, config.Port, config.Username, config.Password, config.DBName)
	log.Printf("📡 Opening database connection...")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("❌ Failed to open DB connection: %v", err)
		return nil, err
	}
	log.Printf("✅ Database connection established")
	return db, nil
}

// ✅ CORS Headers Helper Function
func getCORSHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "OPTIONS, POST, PUT",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}
}

// ✅ AWS Lambda Handler for API Gateway
func lambdaHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("🚀 Lambda function started")
	log.Printf("📌 Received request: Path = %s, Method = %s", request.Path, request.HTTPMethod)

	// ✅ Handle CORS Preflight
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    getCORSHeaders(),
			Body:       `{"message":"CORS preflight response"}`,
		}, nil
	}

	// User already authorized by API Gateway authorizer

	// ✅ Route API Requests
	switch request.Path {
	case "/upload/questions":
		return handleQuizUpload(request)
	case "/students/update":
		return handleStudentUpdate(request)
	default:
		log.Printf("❌ Invalid API Path: %s", request.Path)
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers:    getCORSHeaders(),
			Body:       fmt.Sprintf(`{"error":"Invalid API endpoint", "receivedPath": "%s"}`, request.Path),
		}, nil
	}
}

// ✅ Get User Role from Database
func getUserRole(db *sql.DB, email string) (string, error) {
	var role sql.NullString
	err := db.QueryRow("SELECT role FROM students WHERE LOWER(email) = LOWER($1)", email).Scan(&role)
	if err != nil {
		return "", err
	}
	if !role.Valid {
		return "", nil
	}
	return role.String, nil
}

// ✅ Handle Student Update
func handleStudentUpdate(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get authenticated user from authorizer context
	userEmail, err := getUserFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user from context: %v", err)
		return createErrorResponse(401, "Unauthorized"), nil
	}
	log.Printf("🔐 Authenticated user: %s", userEmail)

	var studentUpdate StudentUpdateRequest
	err = json.Unmarshal([]byte(request.Body), &studentUpdate)
	if err != nil {
		log.Println("❌ Error parsing JSON:", err)
		return createErrorResponse(400, "Invalid JSON format"), nil
	}

	// ✅ Validate Required Fields
	if studentUpdate.Email == "" {
		return createErrorResponse(400, "Missing 'email' parameter"), nil
	}

	// ✅ Connect to Database
	db, err := connectDB()
	if err != nil {
		log.Println("❌ Database connection error:", err)
		return createErrorResponse(500, "Database connection failed"), nil
	}
	defer db.Close()

	// ✅ Get User Role
	userRole, err := getUserRole(db, userEmail)
	if err != nil {
		log.Printf("❌ Failed to get user role: %v", err)
		return createErrorResponse(500, "Failed to verify user permissions"), nil
	}

	// ✅ Check Role-Based Permissions
	isSubscriptionUpdate := studentUpdate.Amount != nil
	if isSubscriptionUpdate && userRole != "super" {
		return createErrorResponse(403, "Only 'super' role can update subscription"), nil
	}
	if !isSubscriptionUpdate && userRole != "admin" && userRole != "super" {
		return createErrorResponse(403, "Only 'admin' or 'super' role can update student fields"), nil
	}

	// ✅ Perform Partial Update
	rowsAffected, err := updateStudent(db, studentUpdate)
	if err != nil {
		log.Println("❌ Error updating student:", err)
		return createErrorResponse(500, "Internal server error"), nil
	}

	// ✅ Handle No Matching Record
	if rowsAffected == 0 {
		return createErrorResponse(404, "No student found with the provided email"), nil
	}

	// ✅ Success Response
	return createSuccessResponse("Student updated successfully"), nil
}

// ✅ Function to Update Student in Database
func updateStudent(db *sql.DB, student StudentUpdateRequest) (int64, error) {
	normalizedEmail := strings.ToLower(student.Email)
	log.Printf("🔍 Updating student: Email = %s", normalizedEmail)

	// ✅ Fetch existing sub_exp_date before updating
	var existingSubExpDate sql.NullString
	err := db.QueryRow("SELECT sub_exp_date FROM students WHERE LOWER(email) = $1", normalizedEmail).Scan(&existingSubExpDate)
	if err != nil {
		log.Printf("❌ Failed to fetch existing sub_exp_date for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to fetch existing sub_exp_date: %w", err)
	}

	log.Printf("📅 Existing sub_exp_date: %v", existingSubExpDate.String)

	// ✅ Get today's date in YYYY-MM-DD format
	today := time.Now().Format("2006-01-02")

	// ✅ Start Transaction
	tx, err := db.Begin()
	if err != nil {
		log.Printf("❌ Failed to begin transaction for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if an error occurs

	// ✅ Prepare Dynamic Update Query
	query := "UPDATE students SET "
	params := []interface{}{normalizedEmail} // Email is always first
	paramIndex := 2
	updateFields := []string{}

	// ✅ Handle Name Update
	if student.Name != nil && *student.Name != "" {
		log.Printf("📝 Updating name: %s", *student.Name)
		updateFields = append(updateFields, fmt.Sprintf("name = $%d", paramIndex))
		params = append(params, *student.Name)
		paramIndex++
	}

	// ✅ Handle Phone Number Update
	if student.PhoneNumber != nil && *student.PhoneNumber != "" {
		log.Printf("📞 Updating phone number: %s", *student.PhoneNumber)
		updateFields = append(updateFields, fmt.Sprintf("phone_number = $%d", paramIndex))
		params = append(params, *student.PhoneNumber)
		paramIndex++
	}

	// ✅ Handle Student Class Update
	if student.StudentClass != nil && *student.StudentClass != "" {
		log.Printf("🏫 Updating student class: %s", *student.StudentClass)
		updateFields = append(updateFields, fmt.Sprintf("student_class = $%d", paramIndex))
		params = append(params, *student.StudentClass)
		paramIndex++
	}

	// ✅ Handle Amount Update and Modify sub_exp_date Logic
	if student.Amount != nil {
		log.Printf("💰 Updating amount: %f", *student.Amount)
		updateFields = append(updateFields, fmt.Sprintf("amount = $%d", paramIndex))
		params = append(params, *student.Amount)
		paramIndex++

		// ✅ Check if amount > 0 to update `payment_time`
		if *student.Amount > 0 {
			log.Printf("⏳ Updating payment_time to NOW() since amount > 0")
			updateFields = append(updateFields, "payment_time = NOW()")

			var newSubExpDate string
			if existingSubExpDate.Valid && existingSubExpDate.String >= today {
				// ✅ sub_exp_date is today or future → Extend by 1 year
				log.Printf("📅 Extending sub_exp_date by 1 year from %s", existingSubExpDate.String)
				newSubExpDate = fmt.Sprintf("DATE '%s' + INTERVAL '1 year'", existingSubExpDate.String)
			} else {
				// ✅ sub_exp_date is NULL or past → Set to today + 1 year
				log.Printf("📅 Setting new sub_exp_date as today + 1 year")
				newSubExpDate = fmt.Sprintf("DATE '%s' + INTERVAL '1 year'", today)
			}

			// ✅ Append sub_exp_date update
			updateFields = append(updateFields, fmt.Sprintf("sub_exp_date = %s", newSubExpDate))

			// ✅ Ensure UpdatedBy is set if amount > 0
			if student.UpdatedBy != nil && *student.UpdatedBy != "" {
				log.Printf("👤 Updated by: %s", *student.UpdatedBy)
				updateFields = append(updateFields, fmt.Sprintf("updated_by = $%d", paramIndex))
				params = append(params, *student.UpdatedBy)
				paramIndex++
			}
		} else {
			log.Printf("💰 Amount is 0, skipping sub_exp_date & payment_time update")
		}
	}

	// ✅ If No Fields Provided, Return Error
	if len(updateFields) == 0 {
		log.Printf("⚠️ No valid fields to update for email: %s", normalizedEmail)
		return 0, fmt.Errorf("no valid fields to update")
	}

	// ✅ Construct Final Query
	query += fmt.Sprintf("%s WHERE LOWER(email) = $1", strings.Join(updateFields, ", "))

	log.Printf("📡 Executing query: %s", query)

	// ✅ Execute Query
	result, err := tx.Exec(query, params...)
	if err != nil {
		log.Printf("❌ Failed to execute update for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to execute update: %w", err)
	}

	// ✅ Commit Transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("❌ Failed to commit transaction for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// ✅ Get Number of Updated Rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("❌ Failed to fetch affected rows for email %s: %v", normalizedEmail, err)
		return 0, fmt.Errorf("failed to fetch affected rows: %w", err)
	}

	log.Printf("✅ Successfully updated %d row(s) for email %s", rowsAffected, normalizedEmail)
	return rowsAffected, nil
}

// ✅ Handle Quiz Upload
func handleQuizUpload(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	queryParams := request.QueryStringParameters
	category := queryParams["category"]
	durationStr := queryParams["duration"]
	quizName := queryParams["quizName"]

	if category == "" || durationStr == "" || quizName == "" {
		return createErrorResponse(400, "Missing required query parameters"), nil
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		return createErrorResponse(400, "Invalid duration format"), nil
	}

	fileContent, err := base64.StdEncoding.DecodeString(request.Body)
	if err != nil {
		return createErrorResponse(400, "Invalid file encoding"), nil
	}

	quizData, err := processExcel(fileContent, category, duration, quizName)
	if err != nil {
		return createErrorResponse(500, "Failed to process Excel file"), nil
	}

	err = saveToPostgres(quizData)
	if err != nil {
		return createErrorResponse(500, "Failed to save to database"), nil
	}

	return createSuccessResponse("Quiz uploaded successfully"), nil
}

func processExcel(fileBytes []byte, category string, duration int, quizName string) (QuizData, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return QuizData{}, err
	}

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return QuizData{}, err
	}

	if len(rows) < 2 {
		return QuizData{}, errors.New("insufficient data in the file")
	}

	// Read headers from the first row
	headerMap := make(map[string]int)
	for i, header := range rows[0] {
		headerMap[header] = i
	}

	// Required headers
	requiredHeaders := []string{"Question", "CorrectAnswer", "IncorrectAnswers", "Explanation"}
	for _, header := range requiredHeaders {
		if _, exists := headerMap[header]; !exists {
			return QuizData{}, fmt.Errorf("missing required column: %s", header)
		}
	}

	var questions []Question
	for _, row := range rows[1:] {
		questions = append(questions, Question{
			Question:         getCellValue(row, headerMap, "Question"),
			CorrectAnswer:    getCellValue(row, headerMap, "CorrectAnswer"),
			IncorrectAnswers: getCellValue(row, headerMap, "IncorrectAnswers"),
			Explanation:      getCellValue(row, headerMap, "Explanation"),
		})
	}

	return QuizData{QuizName: quizName, Duration: duration, Category: category, Questions: questions}, nil
}

// Helper function to get cell value safely
func getCellValue(row []string, headerMap map[string]int, key string) string {
	index, exists := headerMap[key]
	if !exists || index >= len(row) {
		return ""
	}
	return row[index]
}

// ✅ Utility: Create Success Response
func createSuccessResponse(message string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    getCORSHeaders(),
		Body:       fmt.Sprintf(`{"message":"%s"}`, message),
	}
}

// ✅ Utility: Create Error Response
func createErrorResponse(statusCode int, errorMessage string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    getCORSHeaders(),
		Body:       fmt.Sprintf(`{"error":"%s"}`, errorMessage),
	}
}

// ✅ Save Data to PostgreSQL
func saveToPostgres(quiz QuizData) error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	questionsJSON, err := json.Marshal(quiz.Questions)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO quiz_questions (quiz_name, duration, category, questions)
		VALUES ($1, $2, $3, $4::jsonb)
		ON CONFLICT (quiz_name)
		DO UPDATE SET duration = EXCLUDED.duration, category = EXCLUDED.category, questions = EXCLUDED.questions;
	`

	_, err = db.Exec(query, quiz.QuizName, quiz.Duration, quiz.Category, questionsJSON)
	return err
}

// ✅ Main Function
func main() {
	lambda.Start(lambdaHandler)
}
