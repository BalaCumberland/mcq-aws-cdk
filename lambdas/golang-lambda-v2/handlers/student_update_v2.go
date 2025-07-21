package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentUpdateV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var updateRequest StudentUpdateRequest
	err := json.Unmarshal([]byte(request.Body), &updateRequest)
	if err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if updateRequest.Email == "" {
		return CreateErrorResponse(400, "Missing required field: 'email'"), nil
	}

	email := strings.ToLower(updateRequest.Email)
	log.Printf("📌 Updating student: %s", email)

	// Get existing student
	student, err := GetStudentFromDynamoDB(email)
	if err != nil {
		log.Printf("❌ Error fetching student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	// Update fields
	if updateRequest.Name != "" {
		student.Name = updateRequest.Name
	}
	if updateRequest.PhoneNumber != "" {
		student.PhoneNumber = updateRequest.PhoneNumber
	}
	if updateRequest.StudentClass != "" {
		student.StudentClass = updateRequest.StudentClass
	}
	if updateRequest.SubExpDate != "" {
		student.SubExpDate = updateRequest.SubExpDate
	}
	if updateRequest.UpdatedBy != "" {
		student.UpdatedBy = updateRequest.UpdatedBy
	}
	if updateRequest.Amount > 0 {
		student.Amount = updateRequest.Amount
	}
	if updateRequest.PaymentTime != "" {
		student.PaymentTime = updateRequest.PaymentTime
	}
	if updateRequest.Role != "" {
		student.Role = updateRequest.Role
	}

	// Save updated student
	err = SaveStudentToDynamoDB(*student)
	if err != nil {
		log.Printf("❌ Error updating student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	response := map[string]interface{}{
		"message": "Student updated successfully",
		"student": student,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}