package handlers

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentLookupV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	targetUID, err := GetTargetUIDFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	student, err := GetStudentFromDynamoDB(targetUID)
	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	responseJSON, _ := json.Marshal(student)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}