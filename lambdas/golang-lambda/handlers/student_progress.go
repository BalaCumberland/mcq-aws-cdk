package handlers

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type ProgressSummary struct {
	Category     string  `json:"category"`
	Percentage   float64 `json:"percentage"`
	Attempted    int     `json:"attempted"`
	Unattempted  int     `json:"unattempted"`
}

type TestScore struct {
	QuizName       string  `json:"quizName"`
	Category       string  `json:"category"`
	CorrectCount   int     `json:"correctCount"`
	WrongCount     int     `json:"wrongCount"`
	SkippedCount   int     `json:"skippedCount"`
	TotalCount     int     `json:"totalCount"`
	Percentage     float64 `json:"percentage"`
	TotalAttempts  int     `json:"totalAttempts"`
	LatestScore    float64 `json:"latestScore"`
	AttemptedAt    string  `json:"attemptedAt"`
}

type ProgressResponse struct {
	Email           string                       `json:"email"`
	CategorySummary []ProgressSummary            `json:"categorySummary"`
	IndividualTests map[string][]TestScore       `json:"individualTests"`
}

func SaveQuizAttempt(email, quizName, category string, correctCount, wrongCount, skippedCount, totalCount int, percentage float64, results []QuestionResult) error {
	db, err := ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Convert results to JSON
	resultsJSON, _ := json.Marshal(results)

	// Upsert - increment attempt number on conflict
	_, err = db.Exec(`
		INSERT INTO student_quiz_attempts 
		(email, quiz_name, category, correct_count, wrong_count, skipped_count, total_count, percentage, attempt_number, results)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1, $9)
		ON CONFLICT (email, quiz_name) 
		DO UPDATE SET 
			category = EXCLUDED.category,
			correct_count = EXCLUDED.correct_count,
			wrong_count = EXCLUDED.wrong_count,
			skipped_count = EXCLUDED.skipped_count,
			total_count = EXCLUDED.total_count,
			percentage = EXCLUDED.percentage,
			attempt_number = student_quiz_attempts.attempt_number + 1,
			attempted_at = now(),
			results = EXCLUDED.results
	`, email, quizName, category, correctCount, wrongCount, skippedCount, totalCount, percentage, resultsJSON)

	return err
}

func HandleStudentProgress(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	email, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	db, err := ConnectDB()
	if err != nil {
		log.Printf("❌ Database connection error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer db.Close()

	// Get student's enrolled subjects
	var studentClass string
	err = db.QueryRow("SELECT student_class FROM students WHERE LOWER(email) = LOWER($1)", email).Scan(&studentClass)
	if err != nil {
		log.Printf("❌ Error getting student class: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	// Get enrolled subjects for this student class
	var enrolledSubjects []string
	for _, category := range VALID_CATEGORIES {
		if strings.HasPrefix(category, studentClass) {
			enrolledSubjects = append(enrolledSubjects, category)
		}
	}

	if len(enrolledSubjects) == 0 {
		return CreateErrorResponse(404, "No subjects found for student class"), nil
	}

	// Create placeholders for IN clause
	placeholders := make([]string, len(enrolledSubjects))
	args := []interface{}{email}
	for i, subject := range enrolledSubjects {
		placeholders[i] = "$" + strconv.Itoa(i+2)
		args = append(args, subject)
	}

	// Get category summary for enrolled subjects only
	categorySummary := []ProgressSummary{}
	
	// Create map to track attempted quizzes per category
	attemptedMap := make(map[string]int)
	percentageMap := make(map[string]float64)
	
	rows, err := db.Query(`
		SELECT 
			category,
			COUNT(*) as attempted_count,
			AVG(percentage) as avg_percentage
		FROM student_quiz_attempts 
		WHERE email = $1 AND category IN (`+strings.Join(placeholders, ",")+`)
		GROUP BY category
	`, args...)
	if err != nil {
		log.Printf("❌ Error fetching category summary: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var attempted int
		var percentage float64
		err := rows.Scan(&category, &attempted, &percentage)
		if err != nil {
			continue
		}
		attemptedMap[category] = attempted
		percentageMap[category] = percentage
	}
	
	// Get total quiz count per category
	for _, category := range enrolledSubjects {
		var totalQuizzes int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM quiz_questions 
			WHERE category = $1
		`, category).Scan(&totalQuizzes)
		if err != nil {
			totalQuizzes = 0
		}
		
		attempted := attemptedMap[category]
		unattempted := totalQuizzes - attempted
		percentage := percentageMap[category]
		
		categorySummary = append(categorySummary, ProgressSummary{
			Category:    category,
			Percentage:  percentage,
			Attempted:   attempted,
			Unattempted: unattempted,
		})
	}

	// Get individual test scores for enrolled subjects only
	individualTests := make(map[string][]TestScore)
	rows2, err := db.Query(`
		SELECT 
			quiz_name, 
			category, 
			correct_count, 
			wrong_count, 
			skipped_count, 
			total_count, 
			percentage, 
			attempt_number as total_attempts,
			percentage as latest_score,
			attempted_at
		FROM student_quiz_attempts 
		WHERE email = $1 AND category IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY attempted_at DESC
	`, args...)
	if err != nil {
		log.Printf("❌ Error fetching individual tests: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer rows2.Close()

	for rows2.Next() {
		var test TestScore
		err := rows2.Scan(&test.QuizName, &test.Category, &test.CorrectCount, &test.WrongCount, &test.SkippedCount, &test.TotalCount, &test.Percentage, &test.TotalAttempts, &test.LatestScore, &test.AttemptedAt)
		if err != nil {
			continue
		}
		individualTests[test.Category] = append(individualTests[test.Category], test)
	}

	response := ProgressResponse{
		Email:           email,
		CategorySummary: categorySummary,
		IndividualTests: individualTests,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}

