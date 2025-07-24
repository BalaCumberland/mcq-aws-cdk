package handlers

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentRegisterV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get UID from context
	uid, err := GetUserFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user from context: %v", err)
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	var registerRequest StudentRegisterRequest
	err = json.Unmarshal([]byte(request.Body), &registerRequest)
	if err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if registerRequest.Name == "" || registerRequest.PhoneNumber == "" || registerRequest.StudentClass == "" {
		return CreateErrorResponse(400, "Missing required fields"), nil
	}

	log.Printf("📌 Registering student: UID=%s, Name=%s, Class=%s", uid, registerRequest.Name, registerRequest.StudentClass)

	// Check if student already exists
	existingStudent, err := GetStudentFromDynamoDB(uid)
	if err != nil {
		log.Printf("❌ Error checking existing student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if existingStudent != nil {
		return CreateErrorResponse(409, "Student already registered"), nil
	}

	// Create new student
	student := StudentItem{
		UID:          uid,
		Name:         registerRequest.Name,
		StudentClass: registerRequest.StudentClass,
		PhoneNumber:  registerRequest.PhoneNumber,
	}

	// Save to DynamoDB
	err = SaveStudentToDynamoDB(student)
	if err != nil {
		log.Printf("❌ Error saving student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	return CreateSuccessResponse("Student registered successfully"), nil
}