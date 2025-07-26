package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentRegisterV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
	studentClass := studentRegister.StudentClass
	if studentClass == "" {
		studentClass = "DEMO"
	}



	// Check if student already exists
	existingStudent, err := GetStudentInfoByEmail(normalizedEmail)
	if err != nil {
		log.Printf("❌ Error checking existing student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if existingStudent != nil {
		return CreateErrorResponse(409, "Student already exists"), nil
	}

	// Generate UID for new student (temporary until Firebase creates user)
	uid := fmt.Sprintf("temp_%d", time.Now().UnixNano())
	studentInfo := StudentInfoItem{
		UID:          uid,
		Email:        normalizedEmail,
		Name:         studentRegister.Name,
		PhoneNumber:  studentRegister.PhoneNumber,
		StudentClass: studentClass,
	}

	// Save new student
	err = SaveStudentInfoToDynamoDB(studentInfo)
	if err != nil {
		log.Printf("❌ Error saving student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	studentData := map[string]interface{}{
		"uid":          studentInfo.UID,
		"name":         studentInfo.Name,
		"studentClass": studentInfo.StudentClass,
		"phoneNumber":  studentInfo.PhoneNumber,
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