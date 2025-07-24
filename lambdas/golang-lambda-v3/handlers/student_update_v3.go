package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentUpdateV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userRole, err := CheckAdminRole(request)
	if err != nil {
		return CreateErrorResponse(403, err.Error()), nil
	}

	var updateRequest StudentUpdateRequest
	err = json.Unmarshal([]byte(request.Body), &updateRequest)
	if err != nil {
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	// Get target UID from authorizer context
	targetUID, err := GetTargetUIDFromContext(request)
	if err != nil {
		return CreateErrorResponse(400, "Missing target student identifier"), nil
	}

	log.Printf("📌 Admin updating student: %s", targetUID)

	isSubscriptionUpdate := updateRequest.Amount > 0
	if isSubscriptionUpdate && userRole != "super" {
		return CreateErrorResponse(403, "Only 'super' role can update subscription amounts"), nil
	}

	student, err := GetStudentFromDynamoDB(targetUID)
	if err != nil || student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	if updateRequest.Name != "" {
		student.Name = updateRequest.Name
	}
	if updateRequest.PhoneNumber != "" {
		student.PhoneNumber = updateRequest.PhoneNumber
	}
	if updateRequest.StudentClass != "" {
		student.StudentClass = updateRequest.StudentClass
	}

	if updateRequest.Amount > 0 {
		student.Amount = updateRequest.Amount
		now := time.Now()
		student.PaymentTime = now.Format("2006-01-02T15:04:05.000Z")
		student.SubExpDate = now.AddDate(1, 0, 0).Format("2006-01-02T15:04:05Z")
		if updateRequest.UpdatedBy != "" {
			student.UpdatedBy = updateRequest.UpdatedBy
		}
	}

	err = SaveStudentToDynamoDB(*student)
	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	return CreateSuccessResponse("Student updated successfully"), nil
}