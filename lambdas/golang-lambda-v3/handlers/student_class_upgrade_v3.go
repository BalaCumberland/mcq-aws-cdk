package handlers

import (
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type ClassUpgradeRequest struct {
	NewClass string `json:"newClass"`
}

func HandleStudentClassUpgradeV3(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	uid, err := GetUserFromContext(request)
	if err != nil {
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	var upgradeRequest ClassUpgradeRequest
	err = json.Unmarshal([]byte(request.Body), &upgradeRequest)
	if err != nil {
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if upgradeRequest.NewClass == "" {
		return CreateErrorResponse(400, "Missing 'newClass' parameter"), nil
	}

	log.Printf("📌 Upgrading class for student: %s to %s", uid, upgradeRequest.NewClass)

	student, err := GetStudentFromDynamoDB(uid)
	if err != nil || student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	currentClass := student.StudentClass
	newClass := upgradeRequest.NewClass

	// Validate upgrade path
	if !isValidUpgrade(currentClass, newClass) {
		return CreateErrorResponse(400, "Invalid class upgrade path"), nil
	}

	// Delete all quiz attempts for this student
	log.Printf("🗑️ Deleting all quiz attempts for student: %s", uid)
	queryResult, err := dynamoClient.Query(&dynamodb.QueryInput{
		TableName: aws.String("student_quiz_attempts_v3"),
		KeyConditionExpression: aws.String("uid = :uid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uid": {S: aws.String(uid)},
		},
	})

	if err == nil {
		for _, item := range queryResult.Items {
			if quizName := item["quiz_name"]; quizName != nil && quizName.S != nil {
				_, _ = dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
					TableName: aws.String("student_quiz_attempts_v3"),
					Key: map[string]*dynamodb.AttributeValue{
						"uid": {S: aws.String(uid)},
						"quiz_name": {S: quizName.S},
					},
				})
			}
		}
	}

	// Update student class
	student.StudentClass = newClass

	// Save updated student
	err = SaveStudentToDynamoDB(*student)
	if err != nil {
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	response := map[string]interface{}{
		"message":     "Class upgraded successfully",
		"uid":         uid,
		"oldClass":    currentClass,
		"newClass":    newClass,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    GetCORSHeaders(),
		Body:       string(responseJSON),
	}, nil
}

func isValidUpgrade(currentClass, newClass string) bool {
	upgradableClasses := getUpgradableClasses(currentClass)
	for _, allowed := range upgradableClasses {
		if newClass == allowed {
			return true
		}
	}
	return false
}