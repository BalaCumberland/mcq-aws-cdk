package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentUpdateV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Check admin permissions
	userRole, err := CheckAdminRole(request)
	if err != nil {
		log.Printf("❌ Permission denied: %v", err)
		return CreateErrorResponse(403, err.Error()), nil
	}

	var updateRequest StudentUpdateRequest
	err = json.Unmarshal([]byte(request.Body), &updateRequest)
	if err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if updateRequest.UID == "" {
		return CreateErrorResponse(400, "Missing 'uid' parameter"), nil
	}

	log.Printf("📌 Updating student: %s", updateRequest.UID)

	// Additional check for subscription updates - only super
	isSubscriptionUpdate := updateRequest.Amount > 0
	if isSubscriptionUpdate && userRole != "super" {
		return CreateErrorResponse(403, "Only 'super' role can update subscription amounts"), nil
	}

	// Get existing student
	student, err := GetStudentInfoByUID(updateRequest.UID)
	if err != nil {
		log.Printf("❌ Error fetching student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	// Update fields like v1
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
		// Auto-update payment_time and extend sub_exp_date like v1
		now := time.Now()
		student.PaymentTime = now.Format("2006-01-02T15:04:05.000Z")
		
		// Extend subscription by 1 year from existing date or today
		today := now.Format("2006-01-02")
		var newExpiry time.Time
		
		if student.SubExpDate != nil {
			if dateStr, ok := student.SubExpDate.(string); ok && dateStr != "" {
				if existingDate, err := time.Parse("2006-01-02T15:04:05Z", dateStr); err == nil && existingDate.Format("2006-01-02") >= today {
					newExpiry = existingDate.AddDate(1, 0, 0)
				} else {
					newExpiry = now.AddDate(1, 0, 0)
				}
			} else {
				newExpiry = now.AddDate(1, 0, 0)
			}
		} else {
			newExpiry = now.AddDate(1, 0, 0)
		}
		
		student.SubExpDate = newExpiry.Format("2006-01-02T15:04:05Z")
		
		if updateRequest.UpdatedBy != "" {
			student.UpdatedBy = updateRequest.UpdatedBy
		}
	}

	// Save updated student
	err = SaveStudentInfoToDynamoDB(*student)
	if err != nil {
		log.Printf("❌ Error updating student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	return CreateSuccessResponse("Student updated successfully"), nil
}