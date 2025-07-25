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

	// Add target email from authorizer context
	targetEmail := GetTargetEmailFromContext(request)
	response := map[string]interface{}{
		"uid":           student.UID,
		"name":          student.Name,
		"student_class": student.StudentClass,
		"phone_number":  student.PhoneNumber,
		"sub_exp_date":  student.SubExpDate,
		"updated_by":    student.UpdatedBy,
		"amount":        student.Amount,
		"payment_time":  student.PaymentTime,
		"role":          student.Role,
		"email":         targetEmail,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}