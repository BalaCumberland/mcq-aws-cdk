package handlers

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func HandleStudentLookup(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Check admin role
	_, err := CheckAdminRole(request)
	if err != nil {
		log.Printf("❌ Access denied: %v", err)
		return CreateErrorResponse(403, "Access denied"), nil
	}

	identifier := request.QueryStringParameters["identifier"]
	if identifier == "" {
		return CreateErrorResponse(400, "Missing 'identifier' parameter"), nil
	}

	identifier = strings.ToLower(strings.TrimSpace(identifier))
	log.Printf("🔍 Looking up student: %s", identifier)

	var student *StudentInfoItem

	// Try email lookup first
	if strings.Contains(identifier, "@") {
		student, err = GetStudentInfoByEmail(identifier)
	} else {
		// Phone number lookup - scan table
		result, err := dynamoClient.Scan(&dynamodb.ScanInput{
			TableName:        aws.String("students_info"),
			FilterExpression: aws.String("phone_number = :phone"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":phone": {S: aws.String(identifier)},
			},
		})
		if err == nil && len(result.Items) > 0 {
			student = &StudentInfoItem{}
			err = dynamodbattribute.UnmarshalMap(result.Items[0], student)
		}
	}

	if err != nil {
		log.Printf("❌ Error looking up student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	// Get subjects for student class
	subjects, _ := FetchSubjects(student.StudentClass)
	upgradableClasses := getUpgradableClasses(student.StudentClass)

	studentData := map[string]interface{}{
		"id":                1,
		"uid":               student.UID,
		"email":             student.Email,
		"name":              student.Name,
		"student_class":     student.StudentClass,
		"phone_number":      student.PhoneNumber,
		"subjects":          subjects,
		"upgradable_classes": upgradableClasses,
	}

	if student.SubExpDate != nil {
		studentData["sub_exp_date"] = student.SubExpDate
		// Check if subscription is not expired
		if subExpStr, ok := student.SubExpDate.(string); ok {
			if subExpTime, err := time.Parse(time.RFC3339, subExpStr); err == nil {
				if subExpTime.After(time.Now()) {
					studentData["payment_status"] = "PAID"
				} else {
					studentData["payment_status"] = "EXPIRED"
				}
			}
		}
	}
	if student.UpdatedBy != nil {
		studentData["updated_by"] = student.UpdatedBy
	}
	if student.Amount != nil {
		studentData["amount"] = student.Amount
	}
	if student.PaymentTime != nil {
		studentData["payment_time"] = student.PaymentTime
	}
	if student.Role != nil {
		studentData["role"] = student.Role
	}

	responseJSON, _ := json.Marshal(studentData)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}