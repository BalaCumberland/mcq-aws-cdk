package handlers

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentUpdateV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userEmail, err := GetUserFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user from context: %v", err)
		return CreateErrorResponse(401, "Unauthorized"), nil
	}
	log.Printf("🔐 Authenticated user: %s", userEmail)

	var updateRequest StudentUpdateRequest
	err = json.Unmarshal([]byte(request.Body), &updateRequest)
	if err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if updateRequest.Email == "" {
		return CreateErrorResponse(400, "Missing 'email' parameter"), nil
	}

	email := strings.ToLower(updateRequest.Email)
	log.Printf("📌 Updating student: %s", email)

	// Get user role
	userStudent, _ := GetStudentFromDynamoDB(strings.ToLower(userEmail))
	userRole := "student"
	if userStudent != nil && userStudent.Role != nil {
		if roleStr, ok := userStudent.Role.(string); ok {
			userRole = roleStr
		}
	}

	// Check permissions
	isSubscriptionUpdate := updateRequest.Amount > 0
	if isSubscriptionUpdate && userRole != "super" {
		return CreateErrorResponse(403, "Only 'super' role can update subscription"), nil
	}
	if !isSubscriptionUpdate && userRole != "admin" && userRole != "super" {
		return CreateErrorResponse(403, "Only 'admin' or 'super' role can update student fields"), nil
	}

	// Get existing student
	student, err := GetStudentFromDynamoDB(email)
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
	err = SaveStudentToDynamoDB(*student)
	if err != nil {
		log.Printf("❌ Error updating student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	return CreateSuccessResponse("Student updated successfully"), nil
}