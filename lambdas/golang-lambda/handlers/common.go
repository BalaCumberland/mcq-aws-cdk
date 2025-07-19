package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	_ "github.com/lib/pq"
)

type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type QuizData struct {
	QuizName  string     `json:"quizName"`
	Duration  int        `json:"duration"`
	Category  string     `json:"category"`
	Questions []Question `json:"questions"`
}

type Question struct {
	Explanation   string   `json:"explanation"`
	Question      string   `json:"question"`
	CorrectAnswer string   `json:"correctAnswer"`
	AllAnswers    []string `json:"allAnswers"`
}

type StudentUpdateRequest struct {
	Email        string   `json:"email"`
	PhoneNumber  *string  `json:"phoneNumber,omitempty"`
	Name         *string  `json:"name,omitempty"`
	StudentClass *string  `json:"studentClass,omitempty"`
	Amount       *float64 `json:"amount,omitempty"`
	UpdatedBy    *string  `json:"updatedBy,omitempty"`
}

type StudentRegisterRequest struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	PhoneNumber  string `json:"phoneNumber"`
	StudentClass string `json:"studentClass"`
}

var VALID_CATEGORIES = []string{
	"CLS6-TELUGU", "CLS6-HINDI", "CLS6-ENGLISH", "CLS6-MATHS", "CLS6-SCIENCE", "CLS6-SOCIAL",
	"CLS7-TELUGU", "CLS7-HINDI", "CLS7-ENGLISH", "CLS7-MATHS", "CLS7-SCIENCE", "CLS7-SOCIAL",
	"CLS8-TELUGU", "CLS8-HINDI", "CLS8-ENGLISH", "CLS8-MATHS", "CLS8-SCIENCE", "CLS8-SOCIAL",
	"CLS9-TELUGU", "CLS9-HINDI", "CLS9-ENGLISH", "CLS9-MATHS", "CLS9-SCIENCE", "CLS9-SOCIAL",
	"CLS10-TELUGU", "CLS10-HINDI", "CLS10-ENGLISH", "CLS10-MATHS", "CLS10-SCIENCE", "CLS10-SOCIAL",
	"CLS10-BRIDGE", "CLS10-POLYTECHNIC", "CLS10-FORMULAS",
	"CLS11-MPC-PHYSICS", "CLS11-MPC-MATHS1A", "CLS11-MPC-MATHS1B", "CLS11-MPC-CHEMISTRY",
	"CLS11-MPC-EAMCET", "CLS11-MPC-JEEMAINS", "CLS11-MPC-JEEADV",
	"CLS12-MPC-PHYSICS", "CLS12-MPC-MATHS2A", "CLS12-MPC-MATHS2B", "CLS12-MPC-CHEMISTRY",
	"CLS12-MPC-EAMCET", "CLS12-MPC-JEEMAINS", "CLS12-MPC-JEEADV",
	"CLS11-BIPC-PHYSICS", "CLS11-BIPC-BOTANY", "CLS11-BIPC-ZOOLOGY", "CLS11-BIPC-CHEMISTRY",
	"CLS11-BIPC-EAPCET", "CLS11-BIPC-NEET",
	"CLS12-BIPC-PHYSICS", "CLS12-BIPC-BOTANY", "CLS12-BIPC-ZOOLOGY", "CLS12-BIPC-CHEMISTRY",
	"CLS12-BIPC-EAPCET", "CLS12-BIPC-NEET",
}

func GetUserFromContext(request events.APIGatewayProxyRequest) (string, error) {
	if request.RequestContext.Authorizer == nil {
		return "", fmt.Errorf("no authorizer context")
	}
	email, ok := request.RequestContext.Authorizer["email"].(string)
	if !ok || email == "" {
		return "", fmt.Errorf("missing user email from authorizer")
	}
	return email, nil
}

func getDBConfig() (*DBConfig, error) {
	log.Printf("🔐 Getting DB config from environment variables...")
	
	host := os.Getenv("DB_HOST")
	portStr := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	
	if host == "" || portStr == "" || username == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("missing database environment variables")
	}
	
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}
	
	config := &DBConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		DBName:   dbname,
	}
	
	log.Printf("✅ DB config loaded from environment")
	return config, nil
}

func ConnectDB() (*sql.DB, error) {
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
	
	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)
	
	log.Printf("✅ Database connection established")
	return db, nil
}

func GetCORSHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "OPTIONS, POST, PUT",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}
}

func CreateSuccessResponse(message string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       fmt.Sprintf(`{"message":"%s"}`, message),
	}
}

func CreateErrorResponse(statusCode int, errorMessage string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    GetCORSHeaders(),
		Body:       fmt.Sprintf(`{"error":"%s"}`, errorMessage),
	}
}

func GetUserRole(db *sql.DB, email string) (string, error) {
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

func SaveToPostgres(quiz QuizData) error {
	db, err := ConnectDB()
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