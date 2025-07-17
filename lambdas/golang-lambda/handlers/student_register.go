package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentRegister(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var studentRegister StudentRegisterRequest
	err := json.Unmarshal([]byte(request.Body), &studentRegister)
	if err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if studentRegister.Email == "" {
		return CreateErrorResponse(400, "Missing required field: 'email'"), nil
	}

	normalizedEmail := strings.ToLower(studentRegister.Email)

	db, err := ConnectDB()
	if err != nil {
		log.Printf("❌ Database connection error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}
	defer db.Close()

	query := `
		INSERT INTO students (email, name, phone_number, student_class) 
		VALUES ($1, $2, $3, $4) 
		ON CONFLICT (email) DO NOTHING 
		RETURNING id, email, name, student_class;
	`

	studentClass := studentRegister.StudentClass
	if studentClass == "" {
		studentClass = "DEMO"
	}

	var id int
	var returnedEmail, returnedName, returnedClass string
	err = db.QueryRow(query, normalizedEmail, studentRegister.Name, studentRegister.PhoneNumber, studentClass).Scan(
		&id, &returnedEmail, &returnedName, &returnedClass)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return CreateErrorResponse(409, "Student already exists"), nil
		}
		log.Printf("❌ Database error: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	studentData := map[string]interface{}{
		"id":           id,
		"email":        returnedEmail,
		"name":         returnedName,
		"studentClass": returnedClass,
	}

	response := map[string]interface{}{
		"message": "Student created successfully",
		"student": studentData,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}