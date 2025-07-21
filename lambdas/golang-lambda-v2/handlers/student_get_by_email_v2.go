package handlers

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentGetByEmailV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get email from authenticated user context
	userEmail, err := GetUserFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user from context: %v", err)
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	email := strings.ToLower(userEmail)
	log.Printf("📌 Fetching student: %s", email)

	student, err := GetStudentFromDynamoDB(email)
	if err != nil {
		log.Printf("❌ Error fetching student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	// Format response exactly like v1
	studentData := map[string]interface{}{
		"id":            1, // Default ID since DynamoDB doesn't have auto-increment
		"email":         student.Email,
		"name":          student.Name,
		"student_class":  student.StudentClass,
		"phone_number":   student.PhoneNumber,
		"sub_exp_date":   nil,
		"updated_by":     nil,
		"amount":         nil,
		"payment_time":   nil,
		"role":           nil,
		"payment_status": "UNPAID",
		"subjects":       []string{},
	}

	// Handle optional fields like v1
	if student.SubExpDate != nil {
		if dateStr, ok := student.SubExpDate.(string); ok && dateStr != "" {
			studentData["sub_exp_date"] = dateStr
		}
	}
	if student.UpdatedBy != nil {
		if updatedBy, ok := student.UpdatedBy.(string); ok {
			studentData["updated_by"] = updatedBy
		}
	}
	if student.Amount != nil {
		if amountStr, ok := student.Amount.(string); ok {
			studentData["amount"] = amountStr
		} else if amountNum, ok := student.Amount.(float64); ok {
			studentData["amount"] = amountNum
		}
	}
	if student.PaymentTime != nil {
		if paymentTime, ok := student.PaymentTime.(string); ok && paymentTime != "" {
			studentData["payment_time"] = paymentTime
		}
	}
	if student.Role != nil {
		if role, ok := student.Role.(string); ok {
			studentData["role"] = role
		}
	}

	// Calculate payment status like v1
	today := time.Now().Format("2006-01-02")
	if student.SubExpDate != nil {
		if dateStr, ok := student.SubExpDate.(string); ok && dateStr != "" {
			if dateStr >= today {
				studentData["payment_status"] = "PAID"
			}
		}
	}

	// Add subjects like v1 using VALID_CATEGORIES
	classPrefix := student.StudentClass
	for _, category := range VALID_CATEGORIES {
		if strings.HasPrefix(category, classPrefix) {
			studentData["subjects"] = append(studentData["subjects"].([]string), category)
		}
	}

	responseJSON, _ := json.Marshal(studentData)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}