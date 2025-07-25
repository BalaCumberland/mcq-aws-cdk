package handlers

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentGetProfile(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Debug: log the entire authorizer context
	log.Printf("🔍 Authorizer context: %+v", request.RequestContext.Authorizer)

	// Get UID from authenticated user context
	userUID, err := GetUserUIDFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user UID from context: %v", err)
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	log.Printf("📌 Fetching student: %s", userUID)

	student, err := GetStudentInfoByUID(userUID)
	if err != nil {
		log.Printf("❌ Error fetching student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		log.Printf("❌ Student not found for UID: %s", userUID)
		return CreateErrorResponse(404, "Student not found"), nil
	}

	log.Printf("✅ Student found: %+v", student)

	// Format response
	studentData := map[string]interface{}{
		"id":             1, // Default ID since DynamoDB doesn't have auto-increment
		"email":          student.Email,
		"uid":            student.UID,
		"name":           student.Name,
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

	// Calculate payment status with proper expiration check
	if student.SubExpDate != nil {
		if subExpStr, ok := student.SubExpDate.(string); ok && subExpStr != "" {
			if subExpTime, err := time.Parse(time.RFC3339, subExpStr); err == nil {
				if subExpTime.After(time.Now()) {
					studentData["payment_status"] = "PAID"
				} else {
					studentData["payment_status"] = "EXPIRED"
				}
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

	// Add upgradable classes
	studentData["upgradable_classes"] = getUpgradableClasses(student.StudentClass)

	responseJSON, _ := json.Marshal(studentData)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}
