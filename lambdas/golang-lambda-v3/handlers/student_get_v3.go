package handlers

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentGetV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get UID from authenticated user context
	uid, err := GetUserFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user from context: %v", err)
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	log.Printf("📌 Fetching student: %s", uid)

	student, err := GetStudentFromDynamoDB(uid)
	if err != nil {
		log.Printf("❌ Error fetching student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	// Get email and phone from Firebase token
	email := GetEmailFromContext(request)
	phoneNumber := GetPhoneFromContext(request)

	// Format response
	studentData := map[string]interface{}{
		"uid":            student.UID,
		"email":          email,
		"phone_number":   phoneNumber,
		"name":           student.Name,
		"student_class":  student.StudentClass,
		"sub_exp_date":   nil,
		"updated_by":     nil,
		"amount":         nil,
		"payment_time":   nil,
		"role":           nil,
		"payment_status": "UNPAID",
		"subjects":       []string{},
	}

	// Handle optional fields
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

	// Calculate payment status
	today := time.Now().Format("2006-01-02")
	if student.SubExpDate != nil {
		if dateStr, ok := student.SubExpDate.(string); ok && dateStr != "" {
			if dateStr >= today {
				studentData["payment_status"] = "PAID"
			}
		}
	}

	// Add subjects using VALID_CATEGORIES
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

func getUpgradableClasses(currentClass string) []string {
	validUpgrades := map[string][]string{
		"CLS6":       {"CLS7"},
		"CLS7":       {"CLS8"},
		"CLS8":       {"CLS9"},
		"CLS9":       {"CLS10"},
		"CLS10":      {"CLS11-MPC", "CLS11-BIPC"},
		"CLS11-MPC":  {"CLS12-MPC"},
		"CLS11-BIPC": {"CLS12-BIPC"},
	}

	allowedUpgrades, exists := validUpgrades[currentClass]
	if !exists {
		return []string{}
	}

	return allowedUpgrades
}