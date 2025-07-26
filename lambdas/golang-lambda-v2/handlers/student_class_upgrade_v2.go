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

func HandleStudentClassUpgradeV2(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get authenticated user UID
	userUID, err := GetUserUIDFromContext(request)
	if err != nil {
		log.Printf("❌ Failed to get user UID from context: %v", err)
		return CreateErrorResponse(401, "Unauthorized"), nil
	}

	var upgradeRequest ClassUpgradeRequest
	err = json.Unmarshal([]byte(request.Body), &upgradeRequest)
	if err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		return CreateErrorResponse(400, "Invalid JSON format"), nil
	}

	if upgradeRequest.NewClass == "" {
		return CreateErrorResponse(400, "Missing 'newClass' parameter"), nil
	}

	log.Printf("📌 Upgrading class for student: %s to %s", userUID, upgradeRequest.NewClass)

	// Get existing student
	student, err := GetStudentInfoByUID(userUID)
	if err != nil {
		log.Printf("❌ Error fetching student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	if student == nil {
		return CreateErrorResponse(404, "Student not found"), nil
	}

	currentClass := student.StudentClass
	newClass := upgradeRequest.NewClass

	// Validate upgrade path
	if !isValidUpgrade(currentClass, newClass) {
		return CreateErrorResponse(400, "Invalid class upgrade path"), nil
	}

	// Delete all quiz attempts for this student
	log.Printf("🗑️ Deleting all quiz attempts for student: %s", userUID)
	queryResult, err := dynamoClient.Query(&dynamodb.QueryInput{
		TableName: aws.String("student_quiz_attempts_v2"),
		KeyConditionExpression: aws.String("uid = :uid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":uid": {S: aws.String(userUID)},
		},
	})

	if err == nil {
		for _, item := range queryResult.Items {
			if quizName := item["quiz_name"]; quizName != nil && quizName.S != nil {
				_, _ = dynamoClient.DeleteItem(&dynamodb.DeleteItemInput{
					TableName: aws.String("student_quiz_attempts_v2"),
					Key: map[string]*dynamodb.AttributeValue{
						"uid": {S: aws.String(userUID)},
						"quiz_name": {S: quizName.S},
					},
				})
			}
		}
	}

	// Update student class
	student.StudentClass = newClass

	// Save updated student
	err = SaveStudentInfoToDynamoDB(*student)
	if err != nil {
		log.Printf("❌ Error updating student: %v", err)
		return CreateErrorResponse(500, "Internal Server Error"), nil
	}

	response := map[string]interface{}{
		"message":     "Class upgraded successfully",
		"uid":         student.UID,
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

func isValidUpgrade(currentClass, newClass string) bool {
	upgradableClasses := getUpgradableClasses(currentClass)
	for _, allowed := range upgradableClasses {
		if newClass == allowed {
			return true
		}
	}
	return false
}