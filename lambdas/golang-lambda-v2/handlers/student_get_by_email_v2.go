package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func HandleStudentGetByEmailV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	email := request.QueryStringParameters["email"]
	if email == "" {
		return CreateErrorResponse(400, "Missing 'email' parameter"), nil
	}

	email = strings.ToLower(email)
	log.Printf("📌 Fetching student: %s", email)

	student, err := GetStudentFromDynamoDB(email)
	if err != nil {
		log.Printf("❌ Error fetching student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	// Format response to match v1 structure
	studentData := map[string]interface{}{
		"email":        student.Email,
		"name":         student.Name,
		"student_class": student.StudentClass,
		"phone_number": student.PhoneNumber,
		"payment_status": "PAID", // Default for migrated data
	}

	// Handle optional fields safely
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

	// Add subjects based on student class (simplified)
	if strings.Contains(student.StudentClass, "BIPC") {
		studentData["subjects"] = []string{
			student.StudentClass + "-PHYSICS",
			student.StudentClass + "-BOTANY", 
			student.StudentClass + "-ZOOLOGY",
			student.StudentClass + "-CHEMISTRY",
			student.StudentClass + "-EAPCET",
			student.StudentClass + "-NEET",
		}
	} else if strings.Contains(student.StudentClass, "MPC") {
		studentData["subjects"] = []string{
			student.StudentClass + "-PHYSICS",
			student.StudentClass + "-MATHS1A",
			student.StudentClass + "-MATHS1B", 
			student.StudentClass + "-CHEMISTRY",
			student.StudentClass + "-EAMCET",
			student.StudentClass + "-JEEMAINS",
		}
	}

	responseJSON, _ := json.Marshal(studentData)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}