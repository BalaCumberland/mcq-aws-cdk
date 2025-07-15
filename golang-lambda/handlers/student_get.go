package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentGetByEmail(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	email := request.QueryStringParameters["email"]
	if email == "" {
		return CreateErrorResponse(400, "Missing 'email' query parameter"), nil
	}

	// URL decode the email parameter
	decodedEmail, err := url.QueryUnescape(email)
	if err != nil {
		log.Printf("❌ URL decode error: %v", err)
		return CreateErrorResponse(400, "Invalid email parameter"), nil
	}
	email = decodedEmail

	normalizedEmail := strings.ToLower(email)

	db, err := ConnectDB()
	if err != nil {
		log.Printf("❌ Database connection error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer db.Close()

	query := `
		SELECT id, email, name, student_class, phone_number, sub_exp_date, updated_by, amount, payment_time, role 
		FROM students 
		WHERE LOWER(email) = LOWER($1)
	`

	var student struct {
		ID            int      `json:"id"`
		Email         string   `json:"email"`
		Name          string   `json:"name"`
		StudentClass  string   `json:"student_class"`
		PhoneNumber   string   `json:"phone_number"`
		SubExpDate    *string  `json:"sub_exp_date"`
		UpdatedBy     *string  `json:"updated_by"`
		Amount        *float64 `json:"amount"`
		PaymentTime   *string  `json:"payment_time"`
		Role          *string  `json:"role"`
		PaymentStatus string   `json:"payment_status"`
		Subjects      []string `json:"subjects"`
	}

	var subExpDate, updatedBy, paymentTime, role sql.NullString
	var amount sql.NullFloat64

	err = db.QueryRow(query, normalizedEmail).Scan(
		&student.ID, &student.Email, &student.Name, &student.StudentClass,
		&student.PhoneNumber, &subExpDate, &updatedBy, &amount, &paymentTime, &role)

	if err != nil {
		if err == sql.ErrNoRows {
			return CreateErrorResponse(404, "Student not found"), nil
		}
		log.Printf("❌ Database error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if subExpDate.Valid {
		student.SubExpDate = &subExpDate.String
	}
	if updatedBy.Valid {
		student.UpdatedBy = &updatedBy.String
	}
	if amount.Valid {
		student.Amount = &amount.Float64
	}
	if paymentTime.Valid {
		student.PaymentTime = &paymentTime.String
	}
	if role.Valid {
		student.Role = &role.String
	}

	today := time.Now().Format("2006-01-02")
	if student.SubExpDate == nil || *student.SubExpDate < today {
		student.PaymentStatus = "UNPAID"
	} else {
		student.PaymentStatus = "PAID"
	}

	classPrefix := student.StudentClass
	for _, category := range VALID_CATEGORIES {
		if strings.HasPrefix(category, classPrefix) {
			student.Subjects = append(student.Subjects, category)
		}
	}

	responseJSON, _ := json.Marshal(student)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}
